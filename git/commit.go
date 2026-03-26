package git

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Commit struct {
	Tree       string
	Parent     []string
	Author     string
	AuthorTime time.Time
	Committer  string
	CommitTime time.Time
	Message    string
}

func NewCommit(tree string, parents []string, author, committer string, authorTime, commitTime time.Time, message string) *Commit {
	return &Commit{
		Tree:       tree,
		Parent:     parents,
		Author:     author,
		AuthorTime: authorTime,
		Committer:  committer,
		CommitTime: commitTime,
		Message:    message,
	}
}

func (c *Commit) Type() ObjectType {
	return ObjectCommit
}

func (c *Commit) Serialize() []byte {
	var lines []string
	lines = append(lines, fmt.Sprintf("tree %s", c.Tree))
	for _, p := range c.Parent {
		lines = append(lines, fmt.Sprintf("parent %s", p))
	}
	lines = append(lines, fmt.Sprintf("author %s %d %s", c.Author, c.AuthorTime.Unix(), c.AuthorTime.Format("-0700")))
	lines = append(lines, fmt.Sprintf("committer %s %d %s", c.Committer, c.CommitTime.Unix(), c.CommitTime.Format("-0700")))
	lines = append(lines, "")
	lines = append(lines, c.Message)
	return []byte(strings.Join(lines, "\n"))
}

func (c *Commit) Parse(data []byte) error {
	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		if line == "" {
			c.Message = strings.Join(lines[i+1:], "\n")
			break
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "tree":
			c.Tree = parts[1]
		case "parent":
			c.Parent = append(c.Parent, parts[1])
		case "author":
			c.Author, c.AuthorTime = parseIdent(parts[1])
		case "committer":
			c.Committer, c.CommitTime = parseIdent(parts[1])
		}
	}

	return nil
}

func parseIdent(s string) (string, time.Time) {
	parts := strings.Fields(s)
	if len(parts) < 3 {
		return s, time.Now()
	}

	name := strings.Join(parts[:len(parts)-2], " ")
	unix, _ := strconv.ParseInt(parts[len(parts)-2], 10, 64)
	zone := parts[len(parts)-1]

	// Parse timezone
	loc := time.UTC
	if len(zone) >= 5 {
		sign := zone[0]
		h, _ := strconv.Atoi(zone[1:3])
		m, _ := strconv.Atoi(zone[3:5])
		offset := h*3600 + m*60
		if sign == '-' {
			offset = -offset
		}
		loc = time.FixedZone("", offset)
	}

	return name, time.Unix(unix, 0).In(loc)
}

func (c *Commit) String() string {
	return fmt.Sprintf("Commit{%s: %s}", c.Tree[:7], strings.Split(c.Message, "\n")[0])
}
