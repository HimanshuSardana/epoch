package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"epoch/config"
	"epoch/git"
)

func runSquash(windowMinutes int, dryRun bool, cfg *config.Config) error {
	repo, err := git.FindRepo()
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	refs := git.NewRefs(repo)
	store := git.NewObjectStore(repo)

	headRef, isSymref, err := refs.ReadHead()
	if err != nil || !isSymref {
		return fmt.Errorf("cannot squash: HEAD is detached")
	}

	sha, err := refs.ReadRef(headRef)
	if err != nil {
		return fmt.Errorf("failed to read HEAD: %w", err)
	}

	commits := []struct {
		sha  string
		time time.Time
		msg  string
	}{}
	current := sha
	for len(commits) < 20 {
		obj, err := store.Read(current)
		if err != nil {
			break
		}
		commit, ok := obj.(*git.Commit)
		if !ok {
			break
		}
		commits = append(commits, struct {
			sha  string
			time time.Time
			msg  string
		}{current, commit.CommitTime, commit.Message})

		if len(commit.Parent) == 0 {
			break
		}
		current = commit.Parent[0]
	}

	if len(commits) < 2 {
		return fmt.Errorf("not enough commits to squash")
	}

	// Group commits by temporal proximity
	groups := [][]int{[]int{0}}
	for i := 1; i < len(commits); i++ {
		lastGroup := groups[len(groups)-1]
		lastIdx := lastGroup[len(lastGroup)-1]
		elapsed := commits[lastIdx].time.Sub(commits[i].time)
		if elapsed.Minutes() < float64(windowMinutes) {
			groups[len(groups)-1] = append(lastGroup, i)
		} else {
			groups = append(groups, []int{i})
		}
	}

	// Show preview
	fmt.Println("Squash groups:")
	for gIdx, group := range groups {
		if len(group) < 2 {
			continue
		}
		fmt.Printf("\nGroup %d (%d commits):\n", gIdx+1, len(group))
		for _, idx := range group {
			fmt.Printf("  %s %s\n", commits[idx].sha[:7], commits[idx].msg)
		}
	}

	// Check if there are any groups to squash
	hasSquashable := false
	for _, group := range groups {
		if len(group) >= 2 {
			hasSquashable = true
			break
		}
	}

	if !hasSquashable {
		fmt.Println("No commits to squash within the time window")
		return nil
	}

	if dryRun {
		fmt.Println("\n(Dry run - no changes made)")
		return nil
	}

	fmt.Print("\nProceed with squash? [y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = input[:len(input)-1]

	if input != "y" && input != "yes" {
		return fmt.Errorf("aborted")
	}

	// For now, just show a message that squash is not fully implemented
	fmt.Println("Squash functionality is under development")
	return nil
}
