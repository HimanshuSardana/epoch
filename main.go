package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func main() {
	root := "./test/"

	err := os.Mkdir(root+".git/", 0o755)
	err = os.Mkdir(root+".git/objects/", 0o755)
	err = os.Mkdir(root+".git/refs/", 0o755)
	err = os.WriteFile(root+".git/HEAD", nil, 0o755)
	err = os.WriteFile(root+".git/config", nil, 0o755)
	if err != nil {
		log.Fatalf(err.Error())
	}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			fmt.Println(path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("error walking the path %q: %v\n", root, err)
	}
}
