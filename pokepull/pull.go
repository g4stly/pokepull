package pokepull

import (
	"fmt"
	"encoding/json"
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

func (pkmn *Pokemon) Print() {
	fmt.Printf("name: %v\n", pkmn.Name)
	for i := 0; i < len(pkmn.Stats); i++ {
		fmt.Printf("%v: %v\n", pkmn.Stats[i].Inner.Name, pkmn.Stats[i].Base)
	}
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
	json.Unmarshal(self.json, &pkmn)
	return &pkmn
}
//==============================================================
var baseURL = "https://pokeapi.co/api/v2/pokemon/"
func Pull(name string) *FmtPokemon {
	// set up url
	var url bytes.Buffer
	for i := 0; i < len(baseURL); i++ { url.WriteByte(baseURL[i]) }
	url.WriteString(fmt.Sprintf("%v/", name))


	json := &PokemonJSON { url: url.String() }
	json.Fetch()
	pkmn := json.Parse()
	return pkmn.ToFmtPokemon()
}
