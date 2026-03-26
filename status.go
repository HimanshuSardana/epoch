package main

import (
	"fmt"
	"os"
	"path/filepath"

	"epoch/git"
)

type StatusOptions struct {
	ShowBranch bool
	Verbose    bool
}

func showStatus(opts StatusOptions) error {
	repo, err := git.FindRepo()
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	index := git.NewIndex()
	if err := index.Read(repo); err != nil {
		return fmt.Errorf("failed to read index: %w", err)
	}

	refs := git.NewRefs(repo)
	branch, err := refs.CurrentBranch()
	if err != nil {
		fmt.Printf("HEAD detached at %s\n", err.Error()[22:29])
	} else if opts.ShowBranch {
		fmt.Printf("On branch %s\n\n", branch)
	}

	// Build index map
	indexMap := make(map[string]*git.IndexEntry)
	for i := range index.Entries {
		idx := &index.Entries[i]
		indexMap[idx.Path] = idx
	}

	// Build HEAD map
	headTree := getHeadTree(repo)
	headMap := make(map[string]git.TreeEntry)
	if headTree != nil {
		for _, e := range headTree.Entries {
			headMap[e.Name] = e
		}
	}

	staged := []string{}
	unstaged := []string{}
	untracked := []string{}

	// Check staged changes (index vs HEAD)
	for _, entry := range index.Entries {
		headEntry, inHead := headMap[entry.Path]
		if !inHead {
			staged = append(staged, entry.Path)
		} else if headEntry.SHA != entry.SHA {
			staged = append(staged, entry.Path)
		}
	}

	// Check for deleted from index (in HEAD but not in index)
	for path := range headMap {
		if indexMap[path] == nil {
			staged = append(staged, path+" (deleted)")
		}
	}

	// Check unstaged changes (working dir vs index)
	for _, entry := range index.Entries {
		fullPath := filepath.Join(repo.Root, entry.Path)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if info.Size() != entry.Size {
			unstaged = append(unstaged, entry.Path)
			continue
		}

		data, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Calculate SHA
		header := fmt.Sprintf("blob %d\x00", len(data))
		content := header + string(data)
		currentSHA := git.HashObject([]byte(content))

		if currentSHA != entry.SHA {
			unstaged = append(unstaged, entry.Path)
		}
	}

	// Check untracked files
	entries, _ := os.ReadDir(repo.Root)
	for _, e := range entries {
		if e.Name() == ".git" || e.Name() == ".epoch.toml" {
			continue
		}
		if _, ok := indexMap[e.Name()]; !ok && !e.IsDir() {
			untracked = append(untracked, e.Name())
		}
	}

	// Print status
	if len(staged) > 0 {
		fmt.Println("Changes to be committed:")
		for _, p := range staged {
			fmt.Printf("  (staged) %s\n", p)
		}
		fmt.Println()
	}

	if len(unstaged) > 0 {
		fmt.Println("Changes not staged for commit:")
		for _, p := range unstaged {
			fmt.Printf("  (modified) %s\n", p)
		}
		fmt.Println()
	}

	if len(untracked) > 0 {
		fmt.Println("Untracked files:")
		for _, p := range untracked {
			fmt.Printf("  %s\n", p)
		}
		fmt.Println()
	}

	if len(staged) == 0 && len(unstaged) == 0 && len(untracked) == 0 {
		fmt.Println("No changes")
	}

	return nil
}

func getHeadTree(repo *git.Repo) *git.Tree {
	refs := git.NewRefs(repo)
	headRef, isSymref, err := refs.ReadHead()
	if err != nil || !isSymref {
		return nil
	}

	sha, err := refs.ReadRef(headRef)
	if err != nil {
		return nil
	}

	obj, err := git.NewObjectStore(repo).Read(sha)
	if err != nil {
		return nil
	}

	commit, ok := obj.(*git.Commit)
	if !ok {
		return nil
	}

	treeObj, err := git.NewObjectStore(repo).Read(commit.Tree)
	if err != nil {
		return nil
	}

	return treeObj.(*git.Tree)
}
