package backdat

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/pkg/errors"
)

type Backup struct {
	Chunks    ChunkStore
	FP        FPStore
	Snapshots SnapshotStore
}

func (b Backup) Backup(path string) error {
	fps, err := b.FP.ListFingerprintSets()
	if err != nil {
		return errors.Wrap(err, "listing fingerprint sets")
	}
	var lastfpset FP
	if len(fps) == 0 {
		lastfpset = &NilFP{}
	} else {
		sort.Slice(fps, func(i, j int) bool { return fps[i].Before(fps[j]) })
		lastfpset, err = b.FP.OpenFingerprintSet(fps[len(fps)-1])
		if err != nil {
			return errors.Wrap(err, "opening fingerprint set")
		}
	}
	ss, start, err := b.Snapshots.NewSnapshot()
	if err != nil {
		return errors.Wrap(err, "initializing snapshot")
	}
	newfpset, err := b.FP.NewFingerprintSet(start)
	if err != nil {
		return errors.Wrap(err, "initializing fingerprint set")
	}

	recurseOut := make(chan Stat)
	var recurseErr error
	go func(errOut *error) {
		defer close(recurseOut)
		*errOut = Recurse(path, recurseOut)
	}(&recurseErr)

	for s := range recurseOut {
		fp := s.Fingerprint()
		haveFP, err := lastfpset.HaveFingerprint(fp)
		if err != nil {
			return errors.Wrap(err, "checking for fingerprint")
		}
		if haveFP {
			if err := newfpset.AddFingerprint(fp); err != nil {
				return errors.Wrap(err, "adding fingerprint")
			}
			continue
		}
		fe, err := NewEntry(s)
		if err != nil {
			return errors.Wrap(err, "creating file entry")
		}
		ss.AddPath(s.Path, fe)
		switch fe.Type {
		case EntryTypeDir:
			continue
		case EntryTypeSymlink:
			continue
		case EntryTypeComplete:
			data, err := ioutil.ReadFile(s.Path)
			if err != nil {
				return errors.Wrap(err, "opening infile: "+s.Path)
			}
			ct, k, err := Encrypt(data)
			if err != nil {
				return errors.Wrap(err, "encrypting chunk")
			}
			if err := b.Chunks.Store(k.C, ct); err != nil {
				return errors.Wrap(err, "storing chunk")
			}
			fe.K = k
		case EntryTypeChunked:
			infh, err := os.Open(s.Path)
			if err != nil {
				return errors.Wrap(err, "opening plaintext file")
			}
			out := make(chan Chunk)
			go Slice(infh, out)
			s256 := sha256.New()
			ks := []*K{}
			for chunk := range out {
				if chunk.Err != nil {
					return errors.Wrap(chunk.Err, "handling slice")
				}
				has, err := b.Chunks.Has(chunk.K.C)
				if err != nil {
					return errors.Wrap(err, "looking for chunk")
				}
				if !has {
					if err := b.Chunks.Store(chunk.K.C, chunk.Data); err != nil {
						return errors.Wrap(err, "storing chunk")
					}
				}
				fmt.Fprint(s256, chunk.K.C)
				ks = append(ks, chunk.K)
			}
			id := binary.BigEndian.Uint64(s256.Sum(nil)[:8])
			ss.AddID(id, ks)
			infh.Close()
		default:
			return errors.Errorf("invalid entry type: ", fe.Type)
		}
		if err := newfpset.AddFingerprint(fp); err != nil {
			return errors.Wrap(err, "adding fingerprint")
		}
	}
	if err := newfpset.Close(); err != nil {
		return errors.Wrap(err, "closing fingerprint set")
	}
	return errors.Wrap(ss.Close(), "closing snapshot")
}
