package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Refs struct {
	repo *Repo
}

func NewRefs(repo *Repo) *Refs {
	return &Refs{repo: repo}
}

func (r *Refs) ReadHead() (string, bool, error) {
	data, err := os.ReadFile(r.repo.HeadRef())
	if err != nil {
		return "", false, fmt.Errorf("failed to read HEAD: %w", err)
	}

	content := strings.TrimSpace(string(data))
	if strings.HasPrefix(content, "ref: ") {
		return content[5:], true, nil
	}
	return content, false, nil
}

func (r *Refs) WriteHead(ref string, isSymref bool) error {
	if isSymref {
		return os.WriteFile(r.repo.HeadRef(), []byte("ref: "+ref), 0644)
	}
	return os.WriteFile(r.repo.HeadRef(), []byte(ref), 0644)
}

func (r *Refs) ReadRef(ref string) (string, error) {
	// Try as direct SHA
	if len(ref) == 40 {
		return ref, nil
	}

	// Try as ref path
	refPath := r.RefPath(ref)
	data, err := os.ReadFile(refPath)
	if err != nil {
		return "", fmt.Errorf("ref not found: %s", ref)
	}
	return strings.TrimSpace(string(data)), nil
}

func (r *Refs) WriteRef(ref, sha string) error {
	refPath := r.RefPath(ref)
	dir := filepath.Dir(refPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create ref directory: %w", err)
	}
	return os.WriteFile(refPath, []byte(sha), 0644)
}

func (r *Refs) RefPath(ref string) string {
	if strings.HasPrefix(ref, "refs/") {
		return filepath.Join(r.repo.GitDir(), ref)
	}
	if strings.HasPrefix(ref, "heads/") {
		return filepath.Join(r.repo.RefsDir(), "heads", ref[6:])
	}
	if strings.HasPrefix(ref, "tags/") {
		return filepath.Join(r.repo.RefsDir(), "tags", ref[5:])
	}
	// Default to heads/
	return filepath.Join(r.repo.RefsDir(), "heads", ref)
}

func (r *Refs) ListBranches() ([]string, error) {
	headsDir := filepath.Join(r.repo.RefsDir(), "heads")
	entries, err := os.ReadDir(headsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	branches := make([]string, 0)
	for _, e := range entries {
		if !e.IsDir() {
			branches = append(branches, e.Name())
		}
	}
	return branches, nil
}

func (r *Refs) DeleteRef(ref string) error {
	refPath := r.RefPath(ref)
	return os.Remove(refPath)
}

func (r *Refs) IsBranch(ref string) bool {
	refPath := filepath.Join(r.repo.RefsDir(), "heads", ref)
	_, err := os.Stat(refPath)
	return err == nil
}

func (r *Refs) CurrentBranch() (string, error) {
	ref, isSymref, err := r.ReadHead()
	if err != nil {
		return "", err
	}
	if !isSymref {
		return "", fmt.Errorf("detached HEAD at %s", ref[:7])
	}
	// Extract branch name from refs/heads/BRANCH
	if strings.HasPrefix(ref, "refs/heads/") {
		return ref[11:], nil
	}
	return ref, nil
}
