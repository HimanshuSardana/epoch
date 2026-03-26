package main

import (
	"fmt"
	"os"
	"path/filepath"

	"epoch/git"
)

func addFiles(paths []string) error {
	repo, err := git.FindRepo()
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	store := git.NewObjectStore(repo)
	index := git.NewIndex()
	if err := index.Read(repo); err != nil {
		return fmt.Errorf("failed to read index: %w", err)
	}

	for _, path := range paths {
		if err := addFile(repo, store, index, path); err != nil {
			return err
		}
	}

	if err := index.Write(repo); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	fmt.Printf("Added %d file(s)\n", len(paths))
	return nil
}

func addFile(repo *git.Repo, store *git.ObjectStore, index *git.Index, path string) error {
	fullPath := filepath.Join(repo.Root, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("cannot stat %s: %w", path, err)
	}

	if info.IsDir() {
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return fmt.Errorf("cannot read directory %s: %w", path, err)
		}
		for _, e := range entries {
			subPath := filepath.Join(path, e.Name())
			if err := addFile(repo, store, index, subPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			}
		}
		return nil
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}

	blob := git.NewBlob(data)
	sha, err := store.Write(blob)
	if err != nil {
		return fmt.Errorf("failed to write blob: %w", err)
	}

	mode := "100644"
	if info.Mode()&0111 != 0 {
		mode = "100755"
	}

	index.Add(path, sha, mode, info.Size())
	return nil
}
