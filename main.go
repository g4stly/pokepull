package main

import (
	"os"
	"fmt"
	"github.com/g4stly/pokepull/pokepull"
)

func main() {
	if len(os.Args) < 2 {
		error_message := []byte(fmt.Sprintf("USAGE: %v <name>\n", os.Args[0]))
		os.Stderr.Write(error_message)
		os.Exit(-1)
	}

	pkmn := pokepull.Pull(os.Args[1])
	json, err := pkmn.ToJson()
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		return
	}
	os.Stdout.Write(json)
	os.Exit(0)
}
