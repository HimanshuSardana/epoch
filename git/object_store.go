package git

import (
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type ObjectStore struct {
	repo *Repo
}

func NewObjectStore(repo *Repo) *ObjectStore {
	return &ObjectStore{repo: repo}
}

func (s *ObjectStore) Write(obj Object) (string, error) {
	data, err := WriteObject(obj)
	if err != nil {
		return "", fmt.Errorf("failed to serialize object: %w", err)
	}

	sha := HashObject(data)
	dir := filepath.Join(s.repo.ObjectsDir(), sha[:2])
	file := filepath.Join(dir, sha[2:])

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create object directory: %w", err)
	}

	f, err := os.Create(file)
	if err != nil {
		return "", fmt.Errorf("failed to create object file: %w", err)
	}
	defer f.Close()

	zw := zlib.NewWriter(f)
	if _, err := zw.Write(data); err != nil {
		return "", fmt.Errorf("failed to compress object: %w", err)
	}
	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("failed to close compressor: %w", err)
	}

	return sha, nil
}

func (s *ObjectStore) Read(sha string) (Object, error) {
	if len(sha) < 2 {
		return nil, fmt.Errorf("invalid SHA: %s", sha)
	}

	file := filepath.Join(s.repo.ObjectsDir(), sha[:2], sha[2:])
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open object: %w", err)
	}
	defer f.Close()

	zr, err := zlib.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress object: %w", err)
	}
	defer zr.Close()

	data, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}

	return ReadObject(data)
}

func (s *ObjectStore) Exists(sha string) bool {
	file := filepath.Join(s.repo.ObjectsDir(), sha[:2], sha[2:])
	_, err := os.Stat(file)
	return err == nil
}
