package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"epoch/config"
	"epoch/git"
)

var (
	version = "0.1.0"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "init":
		handleInit(args)
	case "add":
		handleAdd(args)
	case "status":
		handleStatus(args)
	case "diff":
		handleDiff(args)
	case "commit":
		handleCommit(args)
	case "reword":
		handleReword(args)
	case "squash":
		handleSquash(args)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func handleInit(args []string) {
	flagset := flag.NewFlagSet("init", flag.ExitOnError)
	root := flagset.String("root", ".", "Repository root directory")
	flagset.Parse(args)

	if err := initializeRepo(*root); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleAdd(args []string) {
	flagset := flag.NewFlagSet("add", flag.ExitOnError)
	flagset.Parse(args)

	if flagset.NArg() == 0 {
		fmt.Println("Usage: epoch add <files>")
		os.Exit(1)
	}

	if err := addFiles(flagset.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleStatus(args []string) {
	flagset := flag.NewFlagSet("status", flag.ExitOnError)
	showBranch := flagset.Bool("branch", false, "Show branch name")
	verbose := flagset.Bool("v", false, "Show verbose status")
	flagset.Parse(args)

	opts := StatusOptions{
		ShowBranch: *showBranch,
		Verbose:    *verbose,
	}
	if err := showStatus(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleDiff(args []string) {
	flagset := flag.NewFlagSet("diff", flag.ExitOnError)
	staged := flagset.Bool("staged", false, "Show staged changes")
	flagset.Parse(args)

	if err := showDiff(*staged); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleCommit(args []string) {
	flagset := flag.NewFlagSet("commit", flag.ExitOnError)
	message := flagset.String("m", "", "Commit message")
	auto := flagset.Bool("auto", false, "Auto-generate commit message using AI")
	yes := flagset.Bool("y", false, "Skip confirmation prompt")
	flagset.Parse(args)

	opts := CommitOptions{
		Message: *message,
		Auto:    *auto,
		Yes:     *yes,
	}

	if err := runCommit(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleReword(args []string) {
	flagset := flag.NewFlagSet("reword", flag.ExitOnError)
	auto := flagset.Bool("auto", false, "Auto-rewrite commit message using AI")
	yes := flagset.Bool("y", false, "Skip confirmation prompt")
	flagset.Parse(args)

	if flagset.NArg() == 0 {
		fmt.Println("Usage: epoch reword [--auto] <commit>")
		os.Exit(1)
	}

	commitRef := flagset.Args()[0]
	repo, err := git.FindRepo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not in a git repository\n")
		os.Exit(1)
	}

	cfg, err := config.Load(filepath.Join(repo.Root, ".epoch.toml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	err = rewordCommit(commitRef, *auto, *yes, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleSquash(args []string) {
	flagset := flag.NewFlagSet("squash", flag.ExitOnError)
	window := flagset.Int("window", 0, "Time window in minutes")
	dryRun := flagset.Bool("dry-run", false, "Show what would be squashed")
	flagset.Parse(args)

	repo, err := git.FindRepo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not in a git repository\n")
		os.Exit(1)
	}

	cfg, err := config.Load(filepath.Join(repo.Root, ".epoch.toml"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	wm := cfg.Squash.WindowMinutes
	if *window > 0 {
		wm = *window
	}

	err = runSquash(wm, *dryRun, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`Epoch v%s - AI-Native Git Client

Usage: epoch <command> [options]

Commands:
  init          Initialize a new repository
  add           Add files to staging area
  status        Show working tree status
  diff          Show changes
  commit        Create a commit
  reword        Rewrite commit message
  squash        Auto-squash commits

Options:
  -h, --help    Show this help message

Use "epoch <command> -h" for more information about a command.
`, version)
}
