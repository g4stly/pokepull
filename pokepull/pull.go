package pokepull

import (
	"fmt"
	"log"
	"os"
	"encoding/json"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"bytes"
	"github.com/andelf/go-curl"
)

//==============================================================
type innerStat struct {
	Name	string
}

type stat struct {
	Base	int		`json:"base_stat"`
	Inner	innerStat	`json:"stat"`
}

type Pokemon struct {
	Name	string
	Stats	[]stat		`json:"stats"`
}

type FmtPokemon struct {
	Name		string	`json:"name"`
	Hp		int	`json:"hp"`
	Attack		int	`json:"attack"`
	Defense		int	`json:"defense"`
	Spattack	int	`json:"spattack"`
	Spdefense	int	`json:"spdefense"`
	Speed		int	`json:"speed"`
	stats		map[string]int
}

func (pkmn *Pokemon) ToFmtPokemon() *FmtPokemon {
	new_pkmn := FmtPokemon{
		Name: pkmn.Name,
		stats: make(map[string]int)}
	for i := 0; i < len(pkmn.Stats); i++ {
		target := pkmn.Stats[i].Inner.Name
		switch(target) {
		case "attack":
			new_pkmn.Attack = pkmn.Stats[i].Base
			break;
		case "defense":
			new_pkmn.Defense	= pkmn.Stats[i].Base
			break;
		case "special-attack":
			new_pkmn.Spdefense	= pkmn.Stats[i].Base
			target = "spattack"
			break;
		case "special-defense":
			new_pkmn.Spattack	= pkmn.Stats[i].Base
			target = "spdefense"
			break;
		case "speed":
			new_pkmn.Speed		= pkmn.Stats[i].Base
			break;
		case "hp":
			new_pkmn.Hp		= pkmn.Stats[i].Base
			break;
		}
		new_pkmn.stats[target] = pkmn.Stats[i].Base
	}
	return &new_pkmn
}

func (pkmn *FmtPokemon) Print() {
	fmt.Printf("format name: %v\n", pkmn.Name)
	for key, value := range pkmn.stats {
		fmt.Printf("%v: %v\n", key, value)
	}
}

func (pkmn *FmtPokemon) ToJson() ([]byte, error) {
	output, err := json.Marshal(pkmn)
	if err != nil {
		return []byte(""), err
	}
	return output, nil
}
//==============================================================
type PokemonJSON struct {
	url	string
	json	[]byte
}

func (self *PokemonJSON) Process(buffer []byte, userdata interface{}) bool {
	self.json = append(self.json, buffer...)
	return true
}

func (self *PokemonJSON) Fetch() error {
	easy := curl.EasyInit()
	defer easy.Cleanup()

	easy.Setopt(curl.OPT_URL, self.url)
	easy.Setopt(curl.OPT_SSL_VERIFYPEER, false)
	easy.Setopt(curl.OPT_WRITEFUNCTION, self.Process)

	return easy.Perform()
}

func (self *PokemonJSON) Parse() *Pokemon {
	var pkmn Pokemon
	log.Printf("parse(): %v", string(self.json))
	json.Unmarshal(self.json, &pkmn)
	return &pkmn
}

func (self *PokemonJSON) FmtParse() *FmtPokemon {
	var pkmn FmtPokemon
	json.Unmarshal(self.json, &pkmn)
	return &pkmn
}
//==============================================================
var baseURL = "https://pokeapi.co/api/v2/pokemon/"
func Pull(name string) *FmtPokemon {
	var json PokemonJSON
	// do we have it in our database?
	// first, open database
	db, err := sql.Open("sqlite3", fmt.Sprintf("/home/%v/.pokepull/database.db", os.Getenv("USER")))
	if err != nil { log.Fatal(err) }
	defer db.Close();
	// next, query for json
	stmt, err := db.Prepare("SELECT json FROM pokemon WHERE name=?")
	if err != nil { log.Fatal(err) }
	defer stmt.Close()
	err = stmt.QueryRow(name).Scan(&json.json)
	if err != nil && err != sql.ErrNoRows { // if we error'd out
		log.Fatal(err)
	} else if err == nil {			// if we did not error out
		return json.FmtParse()
	}					// if we error'd out cuz there was no entry

	// set up url
	var url bytes.Buffer
	for i := 0; i < len(baseURL); i++ { url.WriteByte(baseURL[i]) }
	url.WriteString(fmt.Sprintf("%v/", name))


	json.url = url.String()
	json.Fetch()
	pkmn := json.Parse()
	fmt_pkmn := pkmn.ToFmtPokemon()

	// insert new formatted pokemon into database
	insertStmt, err := db.Prepare("INSERT INTO pokemon (name, json) VALUES (?, ?);")
	if err != nil { log.Fatal(err) }
	defer insertStmt.Close()
	literalJson, err := fmt_pkmn.ToJson()
	if err != nil { log.Fatal(err) }
	_ , err = insertStmt.Exec(fmt_pkmn.Name, literalJson)
	if err != nil {
		log.Fatal(err)
	}
	return fmt_pkmn
}
