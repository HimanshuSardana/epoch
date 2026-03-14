package main

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
)

func main() {
	root := "./"

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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
