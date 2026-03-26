package git

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
)

type ObjectType string

const (
	ObjectBlob   ObjectType = "blob"
	ObjectTree   ObjectType = "tree"
	ObjectCommit ObjectType = "commit"
)

type Object interface {
	Type() ObjectType
	Serialize() []byte
	Parse([]byte) error
}

type GitObject struct {
	Type_ ObjectType
	SHA   string
}

func (o *GitObject) Type() ObjectType {
	return o.Type_
}

func HashObject(data []byte) string {
	h := sha1.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func ReadObject(data []byte) (Object, error) {
	nullIdx := bytes.IndexByte(data, 0)
	if nullIdx == -1 {
		return nil, fmt.Errorf("invalid object format")
	}

	header := string(data[:nullIdx])
	var objType ObjectType
	var size int
	_, err := fmt.Sscanf(header, "%s %d", &objType, &size)
	if err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	content := data[nullIdx+1:]
	if len(content) != size {
		return nil, fmt.Errorf("size mismatch: header says %d, got %d", size, len(content))
	}

	switch objType {
	case ObjectBlob:
		b := &Blob{}
		if err := b.Parse(content); err != nil {
			return nil, err
		}
		return b, nil
	case ObjectTree:
		t := &Tree{}
		if err := t.Parse(content); err != nil {
			return nil, err
		}
		return t, nil
	case ObjectCommit:
		c := &Commit{}
		if err := c.Parse(content); err != nil {
			return nil, err
		}
		return c, nil
	default:
		return nil, fmt.Errorf("unknown object type: %s", objType)
	}
}

func WriteObject(obj Object) ([]byte, error) {
	data := obj.Serialize()
	header := fmt.Sprintf("%s %d\x00", obj.Type(), len(data))
	return append([]byte(header), data...), nil
}

func CopyObject(w io.Writer, r io.Reader) (int64, error) {
	return io.Copy(w, r)
}
