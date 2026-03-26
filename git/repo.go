package git

import (
	"fmt"
	"os"
	"path/filepath"
)

type Repo struct {
	Root string
}

func FindRepo() (*Repo, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		info, err := os.Stat(gitDir)
		if err == nil && info.IsDir() {
			return &Repo{Root: dir}, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, fmt.Errorf("not in a git repository")
		}
		dir = parent
	}
}

func (r *Repo) GitDir() string {
	return filepath.Join(r.Root, ".git")
}

func (r *Repo) ObjectsDir() string {
	return filepath.Join(r.GitDir(), "objects")
}

func (r *Repo) RefsDir() string {
	return filepath.Join(r.GitDir(), "refs")
}

func (r *Repo) HeadRef() string {
	return filepath.Join(r.GitDir(), "HEAD")
}

func (r *Repo) IndexFile() string {
	return filepath.Join(r.GitDir(), "index")
}
