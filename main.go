package main

import (
	_ "io/fs"
	"os"
)

func main() {
	root := "./test/"
	validSubcommands := map[string]bool{
		"init":        true,
		"hash-object": true,
	}

	if len(os.Args) < 2 {
		printUsage()
	} else {
		cmd := os.Args[1]

		if !validSubcommands[cmd] {
			printUsage()
			os.Exit(1)
		}

		switch cmd {
		case "init":
			initializeRepo(root)
		case "hash-object":
			fileName := os.Args[2]
			hashObject(fileName)
		}
	}
}
