package backdat

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"time"
)

const (
	EntryTypeDir = iota
	EntryTypeChunked
	EntryTypeComplete
)

type K struct {
	P string
	C string
}

type Stat struct {
	Path    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	IsDir   bool
}

func (s Stat) Fingerprint() string {
	hash := sha256.New()
	fmt.Fprint(hash, s.Path)
	fmt.Fprint(hash, s.Size)
	fmt.Fprint(hash, s.Mode)
	fmt.Fprint(hash, s.ModTime.UnixNano())
	fmt.Fprint(hash, s.IsDir)
	out := hash.Sum(nil)
	fp := make([]byte, base64.StdEncoding.EncodedLen(len(out)))
	base64.StdEncoding.Encode(fp, out)
	return string(fp)
}

type FileEntry struct {
	Type int
	Stat Stat
	K    *K    `json:',omitempty'`
	ID   int64 `json:',omitempty'`
}

type Snapshot struct {
	Paths map[string]*FileEntry
	IDs   map[uint64][]*K
}

func NewEntry(s Stat) *FileEntry {
	fe := &FileEntry{
		Stat: s,
	}
	if s.IsDir {
		fe.Type = EntryTypeDir
	} else {
		if s.Size < ChunkSize {
			fe.K = &K{P: "blarg"}
			fe.Type = EntryTypeComplete
		} else {
			fe.ID = -1
			fe.Type = EntryTypeChunked
		}
	}
	return fe
}
