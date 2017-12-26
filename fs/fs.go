package fs

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"backdat"
)

const (
	ChunksPath      = "chunks"
	FPSetPath       = "fps"
	SnapshotSetPath = "snap"
)

func New(path string) (backdat.ChunkStore, backdat.FPStore, backdat.SnapshotStore, error) {
	fpbase := filepath.Join(path, FPSetPath)
	if err := dirCheck(fpbase); err != nil {
		return nil, nil, nil, errors.Wrap(err, "checking Fingerprint Set path")
	}

	ssbase := filepath.Join(path, SnapshotSetPath)
	if err := dirCheck(ssbase); err != nil {
		return nil, nil, nil, errors.Wrap(err, "checking Snapshot Set path")
	}

	chunkbase := filepath.Join(path, ChunksPath)
	if err := dirCheck(ssbase); err != nil {
		return nil, nil, nil, errors.Wrap(err, "checking Chunks path")
	}

	return ChunkStore(chunkbase), FPStore(fpbase), SnapshotStore(ssbase), nil
}

func Init(path string) error {
	for _, p := range []string{
		ChunksPath,
		FPSetPath,
		SnapshotSetPath,
	} {
		sp := filepath.Join(path, p)
		if err := os.Mkdir(sp, 0700); err != nil {
			return errors.Wrapf(err, "creating store path %q", sp)
		}
	}
	return nil
}

func dirCheck(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return errors.Errorf("not a diretory: %q", path)
	}
	return nil
}
