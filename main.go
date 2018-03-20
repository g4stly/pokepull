package main

import (
	"os"
	"flag"
	"github.com/g4stly/pokepull/pokepull"
)

var name = flag.String("n", "bulbasaur", "indicate by name which pokemon to fetch json for")

func main() {
	flag.Parse()
	pkmn := pokepull.Pull(*name)
	json, err := pkmn.ToJson()
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		return
	}
	os.Stdout.Write(json)
	os.Exit(0)
}
