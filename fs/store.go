package fs

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"backdat"
)

type Store string

func (s Store) Store(ch string, ct []byte) error {
	outPath := filepath.Join(string(s), ch)
	outfh, err := os.Create(outPath)
	if err != nil {
		return errors.Wrap(err, "creating chunk file")
	}
	defer outfh.Close()
	_, err = outfh.Write(ct)
	return errors.Wrap(err, "writing chunk file")
}

func (s Store) Read(ch string) ([]byte, error) {
	inPath := filepath.Join(string(s), ch)
	infh, err := os.Open(inPath)
	if err != nil {
		return nil, errors.Wrap(err, "opening chunk file")
	}
	defer infh.Close()
	ct := make([]byte, backdat.ChunkSize)
	_, err = infh.Read(ct)
	return ct, errors.Wrap(err, "reading chunk file")
}

func (s Store) Has(ch string) (bool, error) {
	inPath := filepath.Join(string(s), ch)
	_, err := os.Stat(inPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "checking for chunk")
	}
	return true, nil
}

func (s Store) List(out chan string, errs chan error, abort chan interface{}) {
	defer close(out)
	defer close(errs)
	infh, err := os.Open(string(s))
	if err != nil {
		errs <- errors.Wrap(err, "opening store directory")
		return
	}
	defer infh.Close()
	names, err := infh.Readdirnames(0)
	if err != nil {
		errs <- errors.Wrap(err, "reading store directory")
		return
	}
	for _, name := range names {
		select {
		case out <- name:
		case <-abort:
			return
		}
	}
}

func (s Store) Delete(ch string) error {
	delPath := filepath.Join(string(s), ch)
	return errors.Wrap(os.Remove(delPath), "deleting chunk")
}
