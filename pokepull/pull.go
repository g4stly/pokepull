package pokepull

import (
	"fmt"
	"log"
	"os"
	"errors"
	"encoding/json"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/andelf/go-curl"
)

//==============================================================
// Pokemon type and co.
// innerStat, stat, and criticalStats are all types that
// are used to translate strings stored as json to
// primitives. this is why Pokemon type does not contain a
// slice of stat types in favor of ints, against what you might
// expect.
//==============================================================

type innerStat struct {
	Name	string
}

type stat struct {
	Base	int		`json:"base_stat"`
	Inner	innerStat	`json:"stat"`
}

type criticalStats struct {
	Name	string
	Stats	[]stat		`json:"stats"`
}

type Pokemon struct {
	Name		string	`json:"name"`
	Hp		int	`json:"hp"`
	Attack		int	`json:"attack"`
	Defense		int	`json:"defense"`
	Spattack	int	`json:"spattack"`
	Spdefense	int	`json:"spdefense"`
	Speed		int	`json:"speed"`
	stats		map[string]int
}

func (self *criticalStats) ToPokemon() *Pokemon {
	pkmn := Pokemon {
		Name: self.Name,
		stats: make(map[string]int)}

	// loop over stats and copy as we go
	for i:= 0; i < len(self.Stats); i++ {
		stat_name := self.Stats[i].Inner.Name
		switch(stat_name) {
		case "attack":
			pkmn.Attack		= self.Stats[i].Base
			break;
		case "defense":
			pkmn.Defense		= self.Stats[i].Base
			break;
		case "special-attack":
			pkmn.Spdefense		= self.Stats[i].Base
			stat_name = "spattack"
			break;
		case "special-defense":
			pkmn.Spattack		= self.Stats[i].Base
			stat_name = "spdefense"
			break;
		case "speed":
			pkmn.Speed		= self.Stats[i].Base
			break;
		case "hp":
			pkmn.Hp			= self.Stats[i].Base
			break;
		}
		pkmn.stats[stat_name] = self.Stats[i].Base
	}

	return &pkmn
}

func (pkmn *Pokemon) Print() {
	fmt.Printf("format name: %v\n", pkmn.Name)
	for key, value := range pkmn.stats {
		fmt.Printf("%v: %v\n", key, value)
	}
}

func (pkmn *Pokemon) ToJson() ([]byte, error) {
	output, err := json.Marshal(pkmn)
	if err != nil {
		return []byte(""), err
	}
	return output, nil
}

//==============================================================
// PokemonJSON
// this type is the representation of raw json data
//==============================================================
type PokemonJSON struct {
	url	string
	rawJson	[]byte
}

func (self *PokemonJSON) Process(buffer []byte, userdata interface{}) bool {
	self.rawJson = append(self.rawJson, buffer...)
	return true
}

func (self *PokemonJSON) Fetch(name string) *PokemonJSON {
	log.Printf("now fetching pokemon: %v\n", name)
	easy := curl.EasyInit()
	defer easy.Cleanup()

	easy.Setopt(curl.OPT_URL, self.url)
	easy.Setopt(curl.OPT_SSL_VERIFYPEER, false)
	easy.Setopt(curl.OPT_WRITEFUNCTION, self.Process)

	err := easy.Perform()
	if err != nil {
		log.Fatalf("Fetch(): %v\n", err)
	}

	// after we've fetched the json,
	// let's store it in our database
	err = self.Store(name)
	if err != nil {
		log.Fatalf("Fetch(): %v\n", err)
	}

	// method chaining is sexy
	return self
}

func (self *PokemonJSON) Store(name string) error {
	// open db connection
	db, err := sql.Open("sqlite3", fmt.Sprintf("/home/%v/.pokepull/database.db", os.Getenv("USER")))
	if err != nil { log.Fatal(err) }
	defer db.Close();

	// sql statement 
	stmt, err := db.Prepare("INSERT INTO pokemon (name, json) VALUES (?, ?);")
	if err != nil { log.Fatalf("Store(): %v\n", err) }
	defer stmt.Close()

	// push the json inside the database 
	res, err := stmt.Exec(name, self.rawJson)
	if err != nil { return err }
	numAffected, err := res.RowsAffected()
	if err != nil { return err }
	if numAffected != 1 { return errors.New("RowsAffected() did not return 1") }

	return nil
}
func (self *PokemonJSON) Parse() *Pokemon {
	// read into critical stats object
	statsObj := criticalStats{}
	err := json.Unmarshal(self.rawJson, &statsObj)
	if err != nil { log.Fatal(err) }

	return statsObj.ToPokemon()
}

//==============================================================
// Main package functions
//==============================================================
func pullFromDB(name string) (*Pokemon, bool) {
	var json PokemonJSON

	// open db connection
	db, err := sql.Open("sqlite3", fmt.Sprintf("/home/%v/.pokepull/database.db", os.Getenv("USER")))
	if err != nil { log.Fatal(err) }
	defer db.Close();

	// sql statement 
	stmt, err := db.Prepare("SELECT json FROM pokemon WHERE name=?;")
	if err != nil { log.Fatal(err) }
	defer stmt.Close()

	// pull the json out of the db
	err = stmt.QueryRow(name).Scan(&json.rawJson)
	if err != nil && err != sql.ErrNoRows {
		// if we error'd out
		log.Fatal(err)
	} else if err == nil {
		// if we did not error out
		return json.Parse(), true
	}
	// if we error'd out cuz there was no entry
	return nil, false
}

var baseURL = "https://pokeapi.co/api/v2/pokemon/"

func Pull(name string) *Pokemon{
	// attempt to pull stats from our local database
	pkmn, ok := pullFromDB(name)
	if ok { return pkmn }

	// things are !ok, pull from pokeapi.co
	var json PokemonJSON
	json.url = fmt.Sprintf("%v%v/", baseURL, name)
	return json.Fetch(name).Parse()
}
