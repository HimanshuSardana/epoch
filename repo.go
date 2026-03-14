package main

import (
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
