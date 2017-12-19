package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

const (
	FPSetFormat = "20060102150405"
)

type FPStore string

func (path FPStore) ListFingerprintSets() ([]time.Time, error) {
	return nil, nil
}

func (path FPStore) OpenFingerprintSet(time.Time) (*FP, error) {
	return nil, nil
}

func (path FPStore) New(t time.Time) (*FP, error) {
	fp := &FP{
		fps:  map[string]bool{},
		t:    t,
		path: string(path),
	}
	return fp, nil
}

type FP struct {
	fps  map[string]bool
	t    time.Time
	path string
}

func (f FP) AddFingerprint(fp string) error {
	f.fps[fp] = true
	return nil
}

func (f FP) HaveFingerprint(fp string) (bool, error) {
	_, ok := f.fps[fp]
	return ok, nil
}

func (f FP) Close() error {
	outpath := filepath.Join(f.path, f.t.UTC().Format(FPSetFormat))
	outfh, err := os.Create(outpath)
	if err != nil {
		return errors.Wrapf(err, "creating FP file %q: %q", outpath, err)
	}
	for k := range f.fps {
		if _, err = fmt.Fprintln(outfh, k); err != nil {
			outfh.Close()
			return errors.Wrapf(err, "writing line to FP file %q: %q", outpath, err)
		}
	}
	return errors.Wrapf(err, "closing FP file %q: %q", outpath, err)
}
