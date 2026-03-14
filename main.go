package main

import (
	"fmt"
	_ "io/fs"
	"log"
	"os"
	"path/filepath"
)

func initializeRepo(DirName string) {
	GitRoot := filepath.Join(DirName, ".git")
	dirs := []string{
		GitRoot,
		filepath.Join(GitRoot, "objects"),
		filepath.Join(GitRoot, "refs"),
	}

	files := []string{
		filepath.Join(GitRoot, "HEAD"),
		filepath.Join(GitRoot, "config"),
	}

	for _, d := range dirs {
		err := os.MkdirAll(d, 0o755)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	for _, f := range files {
		err := os.WriteFile(f, nil, 0o755)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func printUsage() {
	fmt.Println(`Usage: epoch <subcommand>

Available subcommands:
init: 	Initialize a new repository
		`)
}

func main() {
	root := "./test/"
	validSubcommands := map[string]bool{
		"init": true,
	}

	if len(os.Args) < 2 {
		printUsage()
	} else {
		subcommand := os.Args[1]

		if !validSubcommands[subcommand] {
			printUsage()
			os.Exit(1)
		}

		switch subcommand {
		case "init":
			initializeRepo(root)
		}
	}
}
