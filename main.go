package main

import (
	"fmt"
	"flag"
	"./pokepull"
)

var name = flag.String("name", "garchomp", "use to set name of pokemon to pull")
var verbose = flag.Bool("v", false, "verbose mode")

func main() {
	flag.Parse()
	pkmn := pokepull.Pull(*name)
	json, err := pkmn.ToJson()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(json))
}
