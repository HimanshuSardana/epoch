package git

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
)

type TreeEntry struct {
	Mode string
	Name string
	SHA  string
}

type Tree struct {
	Entries []TreeEntry
}

func NewTree() *Tree {
	return &Tree{Entries: make([]TreeEntry, 0)}
}

func (t *Tree) Type() ObjectType {
	return ObjectTree
}

func (t *Tree) Serialize() []byte {
	var buf bytes.Buffer
	for _, e := range t.Entries {
		buf.WriteString(fmt.Sprintf("%s %s\x00", e.Mode, e.Name))
		// SHA is hex string, convert to bytes for binary storage
		shaBytes, _ := decodeHexString(e.SHA)
		buf.Write(shaBytes)
	}
	return buf.Bytes()
}

func (t *Tree) Parse(data []byte) error {
	t.Entries = nil
	reader := bytes.NewReader(data)

	for reader.Len() > 0 {
		// Read until null byte
		var lineBuf bytes.Buffer
		b, err := reader.ReadByte()
		if err != nil {
			return err
		}
		for b != 0 {
			lineBuf.WriteByte(b)
			b, err = reader.ReadByte()
			if err != nil {
				return err
			}
		}

		// Parse mode and name
		parts := strings.SplitN(lineBuf.String(), " ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid tree entry")
		}

		// Read SHA (20 bytes)
		shaBytes := make([]byte, 20)
		n, err := reader.Read(shaBytes)
		if err != nil {
			return err
		}
		if n != 20 {
			return fmt.Errorf("incomplete SHA")
		}

		t.Entries = append(t.Entries, TreeEntry{
			Mode: parts[0],
			Name: parts[1],
			SHA:  encodeHexString(shaBytes),
		})
	}

	return nil
}

func (t *Tree) String() string {
	return fmt.Sprintf("Tree{%d entries}", len(t.Entries))
}

func decodeHexString(s string) ([]byte, error) {
	hexBytes, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return hexBytes, nil
}

func encodeHexString(b []byte) string {
	return hex.EncodeToString(b)
}
