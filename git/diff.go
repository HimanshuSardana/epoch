package git

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DiffHunk struct {
	OldStart, OldLines int
	NewStart, NewLines int
	Lines              []string
}

type DiffFile struct {
	OldPath string
	NewPath string
	Status  string // "added", "deleted", "modified", "renamed"
	Hunks   []DiffHunk
}

type Diff struct {
	Files []DiffFile
}

func GenerateDiff(repo *Repo, index *Index, staged bool) (*Diff, error) {
	diff := &Diff{Files: make([]DiffFile, 0)}

	// Get HEAD commit tree if exists
	var headTree *Tree
	refs := NewRefs(repo)
	headRef, isSymref, err := refs.ReadHead()
	if err == nil && isSymref {
		sha, err := refs.ReadRef(headRef)
		if err == nil {
			obj, err := NewObjectStore(repo).Read(sha)
			if err == nil {
				if commit, ok := obj.(*Commit); ok {
					treeObj, err := NewObjectStore(repo).Read(commit.Tree)
					if err == nil {
						headTree, _ = treeObj.(*Tree)
					}
				}
			}
		}
	}

	// If not staged, also check working directory
	if !staged {
		index.Read(repo)
	}

	// Build path -> entry map from index
	indexMap := make(map[string]*IndexEntry)
	for i := range index.Entries {
		indexMap[index.Entries[i].Path] = &index.Entries[i]
	}

	// Build path -> entry map from HEAD tree
	headMap := make(map[string]TreeEntry)
	if headTree != nil {
		for _, e := range headTree.Entries {
			headMap[e.Name] = e
		}
	}

	store := NewObjectStore(repo)

	// Compare index to HEAD (staged) or working dir to index (unstaged)
	if staged {
		// Staged: compare index to HEAD
		for _, entry := range index.Entries {
			headEntry, inHead := headMap[entry.Path]

			if !inHead {
				// New file in index
				blob, err := store.Read(entry.SHA)
				if err != nil {
					continue
				}
				content := string(blob.(*Blob).Data)
				diff.Files = append(diff.Files, DiffFile{
					OldPath: "/dev/null",
					NewPath: entry.Path,
					Status:  "added",
					Hunks:   generateHunks("", content),
				})
			} else if headEntry.SHA != entry.SHA {
				// Modified
				oldBlob, _ := store.Read(headEntry.SHA)
				newBlob, _ := store.Read(entry.SHA)
				oldContent := ""
				newContent := ""
				if oldBlob != nil {
					oldContent = string(oldBlob.(*Blob).Data)
				}
				if newBlob != nil {
					newContent = string(newBlob.(*Blob).Data)
				}
				diff.Files = append(diff.Files, DiffFile{
					OldPath: entry.Path,
					NewPath: entry.Path,
					Status:  "modified",
					Hunks:   generateHunks(oldContent, newContent),
				})
			}
		}

		// Check for deleted files in index (in HEAD but not in index)
		for path, headEntry := range headMap {
			if indexMap[path] == nil {
				oldBlob, _ := store.Read(headEntry.SHA)
				oldContent := ""
				if oldBlob != nil {
					oldContent = string(oldBlob.(*Blob).Data)
				}
				diff.Files = append(diff.Files, DiffFile{
					OldPath: path,
					NewPath: "/dev/null",
					Status:  "deleted",
					Hunks:   generateHunks(oldContent, ""),
				})
			}
		}
	} else {
		// Unstaged: compare working dir to index
		for _, entry := range index.Entries {
			fullPath := filepath.Join(repo.Root, entry.Path)
			data, err := os.ReadFile(fullPath)
			if err != nil {
				continue // File deleted or unreadable
			}

			currentSHA := HashObject(append([]byte(fmt.Sprintf("blob %d\x00", len(data))), data...))

			if currentSHA != entry.SHA {
				// Working dir differs from index
				oldBlob, _ := store.Read(entry.SHA)
				oldContent := ""
				if oldBlob != nil {
					oldContent = string(oldBlob.(*Blob).Data)
				}
				diff.Files = append(diff.Files, DiffFile{
					OldPath: entry.Path,
					NewPath: entry.Path,
					Status:  "modified",
					Hunks:   generateHunks(oldContent, string(data)),
				})
			}
		}
	}

	return diff, nil
}

func generateHunks(oldContent, newContent string) []DiffHunk {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	var hunks []DiffHunk

	// Simple line-by-line diff
	i, j := 0, 0
	for i < len(oldLines) || j < len(newLines) {
		if i < len(oldLines) && j < len(newLines) && oldLines[i] == newLines[j] {
			i++
			j++
			continue
		}

		// Start of a hunk
		hunk := DiffHunk{
			OldStart: i + 1,
			NewStart: j + 1,
			Lines:    make([]string, 0),
		}

		// Find hunk content
		var oldHunkLines, newHunkLines []string
		for k := 0; k < 20 && (i < len(oldLines) || j < len(newLines)); k++ {
			if i < len(oldLines) && j < len(newLines) && oldLines[i] == newLines[j] {
				break
			}
			if i < len(oldLines) {
				oldHunkLines = append(oldHunkLines, "-"+oldLines[i])
				i++
			}
			if j < len(newLines) {
				newHunkLines = append(newHunkLines, "+"+newLines[j])
				j++
			}
		}

		// Add context line if available
		if i < len(oldLines) {
			oldHunkLines = append(oldHunkLines, " "+oldLines[i-1])
		}

		hunk.Lines = append(hunk.Lines, oldHunkLines...)
		hunk.Lines = append(hunk.Lines, newHunkLines...)
		hunk.OldLines = len(oldHunkLines)
		hunk.NewLines = len(newHunkLines)

		if len(hunk.Lines) > 0 {
			hunks = append(hunks, hunk)
		}
	}

	if len(hunks) == 0 && (oldContent != newContent) {
		// Entire file changed
		hunks = append(hunks, DiffHunk{
			OldStart: 1,
			NewStart: 1,
			OldLines: len(oldLines),
			NewLines: len(newLines),
			Lines:    []string{"@@ -1," + fmt.Sprint(len(oldLines)) + " +1," + fmt.Sprint(len(newLines)) + " @@"},
		})
	}

	return hunks
}

func (d *Diff) String() string {
	var buf bytes.Buffer
	for _, f := range d.Files {
		buf.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", f.NewPath, f.NewPath))
		buf.WriteString(fmt.Sprintf("--- a/%s\n", f.OldPath))
		buf.WriteString(fmt.Sprintf("+++ b/%s\n", f.NewPath))

		for _, hunk := range f.Hunks {
			buf.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", hunk.OldStart, hunk.OldLines, hunk.NewStart, hunk.NewLines))
			for _, line := range hunk.Lines {
				buf.WriteString(line + "\n")
			}
		}
	}
	return buf.String()
}
