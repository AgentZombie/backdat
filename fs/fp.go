package fs

import (
	"backdat"
	"bufio"
	"fmt"
	"io/ioutil"
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
	fis, err := ioutil.ReadDir(string(path))
	if err != nil {
		return nil, errors.Wrap(err, "listing fingerprint sets")
	}
	times := []time.Time{}
	for _, fi := range fis {
		if !fi.Mode().IsRegular() {
			continue
		}
		t, err := time.Parse(FPSetFormat, fi.Name())
		if err != nil {
			continue
		}
		times = append(times, t)
	}
	return times, nil
}

func (path FPStore) OpenFingerprintSet(t time.Time) (backdat.FP, error) {
	inpath := filepath.Join(string(path), t.Format(FPSetFormat))
	infh, err := os.Open(inpath)
	if err != nil {
		return nil, errors.Wrap(err, "opening fingerprint set")
	}
	scanner := bufio.NewScanner(infh)
	fp := &FP{
		fps: map[string]bool{},
	}
	for scanner.Scan() {
		line := scanner.Text()
		fp.fps[line] = true
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "reading fingerprint set")
	}
	return fp, nil
}

func (path FPStore) NewFingerprintSet(t time.Time) (backdat.FP, error) {
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
	if f.t.IsZero() || f.path == "" {
		return errors.New("read-only fingerprint set")
	}
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
