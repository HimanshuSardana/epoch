package main

import (
	"crypto/sha1"
	"fmt"
	_ "io/fs"
	"log"
	"os"
	"path/filepath"
)

func initializeRepo(DirName string) {
	gitRoot := filepath.Join(DirName, ".git")
	dirs := []string{
		gitRoot,
		filepath.Join(gitRoot, "objects"),
		filepath.Join(gitRoot, "refs"),
	}

	files := []string{
		filepath.Join(gitRoot, "HEAD"),
		filepath.Join(gitRoot, "config"),
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

func hashObject(fileName string) {
	header := fmt.Sprintf("blob %d\x00", len(fileName))
	content := header + fileName
	hash := sha1.New()
	hash.Write([]byte(content))
	result := hash.Sum(nil)
	fmt.Printf("%x\n", result)
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
