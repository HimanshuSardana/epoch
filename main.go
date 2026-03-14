package main

import (
	"fmt"
	_ "io/fs"
	"log"
	"os"
	"path/filepath"
)

func InitializeRepo(DirName string) {
	root := DirName
	root += "./.git"
	dirs := []string{
		root,
		filepath.Join(root, "objects"),
		filepath.Join(root, "refs"),
	}

	files := []string{
		filepath.Join(root, "HEAD"),
		filepath.Join(root, "config"),
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}

	for _, file := range files {
		err := os.WriteFile(file, nil, 0o755)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}
}

func main() {
	root := "./test/"
	valid_subcommands := []string{
		"init",
	}

	if len(os.Args) < 2 {
		fmt.Println(`Usage: epoch <subcommand>

Available subcommands:
init: 	Initialize a new repository
		`)
	} else {
		subcommand := os.Args[1]
		IsValid := false
		for _, ValidSubcommand := range valid_subcommands {
			if ValidSubcommand == subcommand {
				IsValid = true
			}
		}

		if IsValid == false {
			fmt.Println(`Usage: epoch <subcommand>

Available subcommands:
init: 	Initialize a new repository
		`)

			os.Exit(1)

		} else {
			if subcommand == "init" {
				InitializeRepo(root)
			}
		}
	}

	// err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if !d.IsDir() {
	// 		fmt.Println(path)
	// 	}
	// 	return nil
	// })
	// if err != nil {
	// 	log.Fatalf("error walking the path %q: %v\n", root, err)
	// }
}
