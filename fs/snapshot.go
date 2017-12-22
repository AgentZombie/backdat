package fs

import (
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"backdat"

	"github.com/pkg/errors"
)

const (
	SnapshotFormat = FPSetFormat
)

type Snapshot struct {
	basePath string
	ts       time.Time
	Paths    map[string]*backdat.FileEntry
	IDs      map[uint64][]*backdat.K
}

func (s Snapshot) AddPath(p string, fe *backdat.FileEntry) error {
	s.Paths[p] = fe
	return nil
}

func (s Snapshot) AddID(id uint64, k []*backdat.K) error {
	s.IDs[id] = k
	return nil
}

func (s Snapshot) Close() error {
	if s.ts.IsZero() || s.basePath == "" {
		return errors.New("snapshot is read-only")
	}
	outpath := filepath.Join(s.basePath, s.ts.Format(SnapshotFormat))
	outfh, err := os.Create(outpath)
	if err != nil {
		return errors.Wrap(err, "creating snapshot file")
	}
	defer outfh.Close()
	gz, err := gzip.NewWriterLevel(outfh, gzip.BestCompression)
	if err != nil {
		return errors.Wrap(err, "creating compressor")
	}
	je := json.NewEncoder(gz)
	return errors.Wrap(je.Encode(s), "encoding snapshot file")
}

type SnapshotStore string

func (ss SnapshotStore) ListSnapshots() ([]time.Time, error) {
	fis, err := ioutil.ReadDir(string(ss))
	if err != nil {
		return nil, errors.Wrap(err, "listing snapshots")
	}
	times := []time.Time{}
	for _, fi := range fis {
		if !fi.Mode().IsRegular() {
			continue
		}
		t, err := time.Parse(SnapshotFormat, fi.Name())
		if err != nil {
			continue
		}
		times = append(times, t)
	}
	return times, nil
}

func (ss SnapshotStore) NewSnapshot() (backdat.Snapshot, time.Time, error) {
	ts := time.Now().UTC()
	s := &Snapshot{
		basePath: string(ss),
		ts:       ts,
		Paths:    map[string]*backdat.FileEntry{},
		IDs:      map[uint64][]*backdat.K{},
	}
	return s, ts, nil
}

func (ss SnapshotStore) OpenSnapshot(t time.Time) (backdat.Snapshot, error) {
	inPath := filepath.Join(string(ss), t.Format(SnapshotFormat))
	infh, err := os.Open(inPath)
	if err != nil {
		return nil, errors.Wrap(err, "opening snapshot")
	}
	gz, err := gzip.NewReader(infh)
	if err != nil {
		return nil, errors.Wrap(err, "decompression snapshot")
	}
	jd := json.NewDecoder(gz)
	s := &Snapshot{}
	if err = jd.Decode(s); err != nil {
		return nil, errors.Wrap(err, "decoding snapshot")
	}
	return s, nil
}
