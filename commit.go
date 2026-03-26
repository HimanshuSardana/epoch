package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"epoch/config"
	"epoch/git"
	"epoch/llm"
)

type CommitOptions struct {
	Message string
	Auto    bool
	Yes     bool
}

func runCommit(opts CommitOptions) error {
	repo, err := git.FindRepo()
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	index := git.NewIndex()
	if err := index.Read(repo); err != nil {
		return fmt.Errorf("failed to read index: %w", err)
	}

	if len(index.Entries) == 0 {
		return fmt.Errorf("nothing to commit")
	}

	msg := opts.Message
	if opts.Auto {
		cfg, err := config.Load(filepath.Join(repo.Root, ".epoch.toml"))
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		diff, err := git.GenerateDiff(repo, index, true)
		if err != nil {
			return fmt.Errorf("failed to generate diff: %w", err)
		}

		generatedMsg, err := generateCommitMessage(diff.String(), cfg)
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
		}

		fmt.Println("Generated commit message:")
		fmt.Println(generatedMsg)

		if !opts.Yes {
			fmt.Print("\nUse this message? [y/n/e]: ")
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = input[:len(input)-1]

			switch input {
			case "y", "yes":
				msg = generatedMsg
			case "e", "edit":
				fmt.Println("Enter new message:")
				reader := bufio.NewReader(os.Stdin)
				newMsg, _ := reader.ReadString('\n')
				msg = newMsg[:len(newMsg)-1]
			default:
				return fmt.Errorf("aborted")
			}
		} else {
			msg = generatedMsg
		}
	}

	if msg == "" {
		return fmt.Errorf("commit message required")
	}

	// Validate conventional commits format
	cfg, _ := config.Load(filepath.Join(repo.Root, ".epoch.toml"))
	if cfg.Commits.EnforceConventional {
		if !isValidConventionalCommit(msg, cfg.Commits.AllowedTypes) {
			return fmt.Errorf("commit message must follow Conventional Commits format (type: description)")
		}
	}

	// Build tree from index
	tree := buildTree(repo, index)

	// Write tree
	store := git.NewObjectStore(repo)
	treeSHA, err := store.Write(tree)
	if err != nil {
		return fmt.Errorf("failed to write tree: %w", err)
	}

	// Get parent commit
	refs := git.NewRefs(repo)
	parentSHA := ""
	headRef, isSymref, _ := refs.ReadHead()
	if isSymref {
		parentSHA, _ = refs.ReadRef(headRef)
	}

	// Create commit
	author := getAuthor()
	now := time.Now()
	commit := git.NewCommit(treeSHA, []string{parentSHA}, author, author, now, now, msg)

	commitSHA, err := store.Write(commit)
	if err != nil {
		return fmt.Errorf("failed to write commit: %w", err)
	}

	// Update HEAD
	if isSymref {
		if err := refs.WriteRef(headRef, commitSHA); err != nil {
			return fmt.Errorf("failed to update ref: %w", err)
		}
	} else {
		refs.WriteHead(commitSHA, false)
	}

	fmt.Printf("Created commit: %s\n", commitSHA[:7])
	fmt.Println(msg)

	return nil
}

func buildTree(repo *git.Repo, index *git.Index) *git.Tree {
	tree := git.NewTree()

	for _, entry := range index.Entries {
		parts := splitPath(entry.Path)
		addTreeEntry(tree, parts, 0, entry.SHA, entry.Mode)
	}

	return tree
}

func addTreeEntry(tree *git.Tree, parts []string, idx int, sha, mode string) {
	if idx >= len(parts)-1 {
		tree.Entries = append(tree.Entries, git.TreeEntry{
			Mode: mode,
			Name: parts[idx],
			SHA:  sha,
		})
		return
	}

	// Find or create subdirectory
	subtree := findOrCreateSubtree(tree, parts[idx])
	subtree.Entries = append(subtree.Entries, git.TreeEntry{
		Mode: "040000",
		Name: parts[idx+1],
		SHA:  "", // Will be filled after building
	})
}

func findOrCreateSubtree(parent *git.Tree, name string) *git.Tree {
	for i, e := range parent.Entries {
		if e.Name == name && e.Mode == "040000" {
			// This is a placeholder, we'd need to look up the actual subtree
			// For now, create a new subtree
			subtree := git.NewTree()
			parent.Entries[i].SHA = "placeholder"
			return subtree
		}
	}

	subtree := git.NewTree()
	parent.Entries = append(parent.Entries, git.TreeEntry{
		Mode: "040000",
		Name: name,
		SHA:  "placeholder",
	})
	return subtree
}

func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, c := range path {
		if c == '/' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	parts = append(parts, current)
	return parts
}

func getAuthor() string {
	// Check git config
	home := os.Getenv("HOME")
	if home != "" {
		configPath := filepath.Join(home, ".gitconfig")
		if data, err := os.ReadFile(configPath); err == nil {
			// Simple parsing - look for name/email
			// This is a simplified version
			for _, line := range splitLines(string(data)) {
				if len(line) > 5 && line[:5] == "name=" {
					name := line[5:]
					email := "user@example.com"
					return fmt.Sprintf("%s <%s>", name, email)
				}
			}
		}
	}
	return "User <user@example.com>"
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if len(current) > 0 {
		lines = append(lines, current)
	}
	return lines
}

func isValidConventionalCommit(msg string, allowedTypes []string) bool {
	msg = trim(msg)

	for _, t := range allowedTypes {
		if len(msg) > len(t)+1 && msg[:len(t)+1] == t+":" {
			return true
		}
	}
	return false
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}

func generateCommitMessage(diff string, cfg *config.Config) (string, error) {
	client, err := llm.NewLLMClient(cfg.AI.Model, cfg.AI.Temperature)
	if err != nil {
		return "", err
	}

	prompt := llm.BuildCommitMessagePrompt(diff)
	msg, err := client.Generate(prompt)
	if err != nil {
		return "", err
	}

	return trim(msg), nil
}
