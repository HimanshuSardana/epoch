package main

import (
	"bufio"
	"fmt"
	"os"

	"epoch/config"
	"epoch/git"
	"epoch/llm"
)

func rewordCommit(commitRef string, auto, yes bool, cfg *config.Config) error {
	repo, err := git.FindRepo()
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	store := git.NewObjectStore(repo)
	refs := git.NewRefs(repo)

	sha, err := refs.ReadRef(commitRef)
	if err != nil {
		return fmt.Errorf("failed to find commit: %w", err)
	}

	obj, err := store.Read(sha)
	if err != nil {
		return fmt.Errorf("failed to read commit: %w", err)
	}

	commit, ok := obj.(*git.Commit)
	if !ok {
		return fmt.Errorf("not a commit object")
	}

	fmt.Printf("Current commit message:\n%s\n\n", commit.Message)

	newMsg := commit.Message
	if auto {
		client, err := llm.NewLLMClient(cfg.AI.Model, cfg.AI.Temperature)
		if err != nil {
			return err
		}

		prompt := llm.BuildImproveCommitPrompt(commit.Message)
		newMsg, err = client.Generate(prompt)
		if err != nil {
			return fmt.Errorf("failed to generate improved message: %w", err)
		}

		fmt.Printf("AI suggested message:\n%s\n", newMsg)

		if !yes {
			fmt.Print("\nUse this message? [y/n/e]: ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = input[:len(input)-1]

			switch input {
			case "y", "yes":
				// Use newMsg
			case "e", "edit":
				fmt.Println("Enter new message:")
				reader := bufio.NewReader(os.Stdin)
				newMsg, _ = reader.ReadString('\n')
				newMsg = newMsg[:len(newMsg)-1]
			default:
				return fmt.Errorf("aborted")
			}
		}
	} else {
		fmt.Println("Enter new commit message:")
		reader := bufio.NewReader(os.Stdin)
		newMsg, _ = reader.ReadString('\n')
		newMsg = newMsg[:len(newMsg)-1]
	}

	// Write the new commit (we create a new commit since commit objects are immutable)
	newCommit := git.NewCommit(
		commit.Tree,
		commit.Parent,
		commit.Author,
		commit.Committer,
		commit.AuthorTime,
		commit.CommitTime,
		newMsg,
	)

	newSHA, err := store.Write(newCommit)
	if err != nil {
		return fmt.Errorf("failed to write new commit: %w", err)
	}

	// Update ref
	headRef, isSymref, _ := refs.ReadHead()
	if isSymref {
		if err := refs.WriteRef(headRef, newSHA); err != nil {
			return fmt.Errorf("failed to update ref: %w", err)
		}
	} else {
		refs.WriteHead(newSHA, false)
	}

	fmt.Printf("Reworded commit: %s -> %s\n", sha[:7], newSHA[:7])
	return nil
}
