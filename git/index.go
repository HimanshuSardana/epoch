package git

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type IndexEntry struct {
	Path  string
	SHA   string
	Mode  string
	Stage int
	Size  int64
	Mtime time.Time
}

type Index struct {
	Version int
	Entries []IndexEntry
}

func NewIndex() *Index {
	return &Index{
		Version: 2,
		Entries: make([]IndexEntry, 0),
	}
}

func (i *Index) Add(path, sha, mode string, size int64) {
	for idx, e := range i.Entries {
		if e.Path == path {
			i.Entries[idx].SHA = sha
			i.Entries[idx].Mode = mode
			i.Entries[idx].Size = size
			i.Entries[idx].Mtime = time.Now()
			return
		}
	}
	i.Entries = append(i.Entries, IndexEntry{
		Path:  path,
		SHA:   sha,
		Mode:  mode,
		Stage: 0,
		Size:  size,
		Mtime: time.Now(),
	})
}

func (i *Index) Remove(path string) {
	newEntries := make([]IndexEntry, 0)
	for _, e := range i.Entries {
		if e.Path != path {
			newEntries = append(newEntries, e)
		}
	}
	i.Entries = newEntries
}

func (i *Index) Get(path string) *IndexEntry {
	for _, e := range i.Entries {
		if e.Path == path {
			return &e
		}
	}
	return nil
}

func (i *Index) Write(repo *Repo) error {
	buf := new(bytes.Buffer)

	// Write header
	buf.Write([]byte("DIRC"))
	binary.Write(buf, binary.BigEndian, uint32(i.Version))
	binary.Write(buf, binary.BigEndian, uint32(len(i.Entries)))

	// Write entries
	for _, e := range i.Entries {
		// Ctime (4 bytes)
		binary.Write(buf, binary.BigEndian, uint32(e.Mtime.Unix()))
		// Mtime (4 bytes)
		binary.Write(buf, binary.BigEndian, uint32(e.Mtime.Unix()))
		// Dev (4 bytes)
		binary.Write(buf, binary.BigEndian, uint32(0))
		// Inode (4 bytes)
		binary.Write(buf, binary.BigEndian, uint32(0))
		// Mode (4 bytes)
		mode, _ := parseMode(e.Mode)
		binary.Write(buf, binary.BigEndian, mode)
		// Object type (4 bytes) - assume blob
		binary.Write(buf, binary.BigEndian, uint32(3)) // blob
		// Size (4 bytes)
		binary.Write(buf, binary.BigEndian, uint32(e.Size))
		// SHA (20 bytes)
		shaBytes, _ := decodeHexString(e.SHA)
		buf.Write(shaBytes)
		// Flags (2 bytes)
		flags := uint16(len(e.Path))
		if len(e.Path) > 0xFFF {
			flags = 0xFFF
		}
		binary.Write(buf, binary.BigEndian, flags)
		// Extended flag (2 bytes)
		binary.Write(buf, binary.BigEndian, uint16(0))
		// Path (null-terminated)
		buf.Write([]byte(e.Path))
		buf.WriteByte(0)
		// Pad to 8-byte alignment
		pad := (8 - (buf.Len() % 8)) % 8
		for j := 0; j < pad; j++ {
			buf.WriteByte(0)
		}
	}

	return os.WriteFile(repo.IndexFile(), buf.Bytes(), 0644)
}

func parseMode(mode string) (uint32, error) {
	switch mode {
	case "100644":
		return 0x81A4, nil // regular file
	case "100755":
		return 0x81ED, nil // executable
	case "040000":
		return 0x4000, nil // directory
	case "120000":
		return 0xA000, nil // symlink
	default:
		return 0x81A4, nil
	}
}

func (i *Index) Read(repo *Repo) error {
	data, err := os.ReadFile(repo.IndexFile())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read index: %w", err)
	}

	if len(data) < 12 {
		return fmt.Errorf("invalid index file")
	}

	// Verify signature
	if string(data[:4]) != "DIRC" {
		return fmt.Errorf("invalid index signature")
	}

	buf := bytes.NewReader(data[4:])
	var version, numEntries uint32
	binary.Read(buf, binary.BigEndian, &version)
	binary.Read(buf, binary.BigEndian, &numEntries)

	i.Version = int(version)
	i.Entries = make([]IndexEntry, 0, numEntries)

	for idx := uint32(0); idx < numEntries; idx++ {
		entry, err := readIndexEntry(buf)
		if err != nil {
			return err
		}
		i.Entries = append(i.Entries, entry)
	}

	return nil
}

func readIndexEntry(buf *bytes.Reader) (IndexEntry, error) {
	// Read fixed-size fields (62 bytes total)
	var ctime, mtime, dev, inode, mode, objType, size uint32
	var sha [20]byte
	var flags, extended uint16

	binary.Read(buf, binary.BigEndian, &ctime)
	binary.Read(buf, binary.BigEndian, &mtime)
	binary.Read(buf, binary.BigEndian, &dev)
	binary.Read(buf, binary.BigEndian, &inode)
	binary.Read(buf, binary.BigEndian, &mode)
	binary.Read(buf, binary.BigEndian, &objType)
	binary.Read(buf, binary.BigEndian, &size)
	buf.Read(sha[:])
	binary.Read(buf, binary.BigEndian, &flags)
	binary.Read(buf, binary.BigEndian, &extended)

	// Read path until null
	pathBuf := make([]byte, 0)
	for {
		b, err := buf.ReadByte()
		if err != nil {
			return IndexEntry{}, err
		}
		if b == 0 {
			break
		}
		pathBuf = append(pathBuf, b)
	}
	path := string(pathBuf)

	// Skip padding to 8-byte boundary
	pad := (8 - (int(buf.Size()-int64(buf.Len())) % 8)) % 8
	if pad > 0 {
		buf.Seek(int64(pad), 1)
	}

	return IndexEntry{
		Path:  path,
		SHA:   encodeHexString(sha[:]),
		Mode:  formatMode(mode),
		Stage: int(flags >> 12),
		Size:  int64(size),
		Mtime: time.Unix(int64(mtime), 0),
	}, nil
}

func formatMode(mode uint32) string {
	switch mode {
	case 0x81A4:
		return "100644"
	case 0x81ED:
		return "100755"
	case 0x4000:
		return "040000"
	case 0xA000:
		return "120000"
	default:
		return "100644"
	}
}

func (i *Index) WorkingDirChanged(repo *Repo) (bool, error) {
	for _, e := range i.Entries {
		fullPath := filepath.Join(repo.Root, e.Path)
		info, err := os.Stat(fullPath)
		if err != nil {
			return true, nil // File missing = changed
		}
		if info.Size() != e.Size {
			return true, nil // Size changed = modified
		}
		if info.ModTime().After(e.Mtime) {
			return true, nil // Modified after staging = changed
		}
	}
	return false, nil
}
