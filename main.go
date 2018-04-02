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

	json, err := pokepull.Pull(os.Args[1]).ToJson()

	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(-1)
	}

	os.Stdout.Write(json)
	os.Exit(0)
}
