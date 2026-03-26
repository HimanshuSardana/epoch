package main

import (
	"fmt"
	"os"
	"path/filepath"

	"epoch/git"
)

func initializeRepo(dir string) error {
	gitRoot := filepath.Join(dir, ".git")
	dirs := []string{
		gitRoot,
		filepath.Join(gitRoot, "objects"),
		filepath.Join(gitRoot, "refs", "heads"),
		filepath.Join(gitRoot, "refs", "tags"),
	}

	files := map[string]string{
		filepath.Join(gitRoot, "HEAD"):        "ref: refs/heads/main\n",
		filepath.Join(gitRoot, "config"):      "[core]\n\trepositoryformatversion = 0\n",
		filepath.Join(gitRoot, "description"): "Unnamed repository\n",
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	index := git.NewIndex()
	repo := &git.Repo{Root: dir}
	if err := index.Write(repo); err != nil {
		return err
	}

	fmt.Printf("Initialized empty epoch repository in %s/.git/\n", dir)
	return nil
}
