package main

import (
	"flag"
	"fmt"

	"epoch/git"
)

func showDiff(staged bool) error {
	repo, err := git.FindRepo()
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	index := git.NewIndex()
	if err := index.Read(repo); err != nil {
		return fmt.Errorf("failed to read index: %w", err)
	}

	diff, err := git.GenerateDiff(repo, index, staged)
	if err != nil {
		return fmt.Errorf("failed to generate diff: %w", err)
	}

	if len(diff.Files) == 0 {
		fmt.Println("No changes")
		return nil
	}

	fmt.Print(diff.String())
	return nil
}

var (
	stagedFlag = flag.Bool("staged", false, "Show staged changes")
)
