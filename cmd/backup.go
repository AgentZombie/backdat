package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"backdat"
	"backdat/fs"
)

func fatalIfError(err error, msg string) {
	if err != nil {
		log.Fatal("error: ", msg, err.Error())
	}
}

func main() {
	store := fs.Store("/tmp/blarg/chunks")

	out := make(chan backdat.Stat)
	go func() {
		defer close(out)
		fatalIfError(backdat.Recurse("/home/jason/go", out), "recursing")
	}()

	ss, _, err := fs.SnapshotStore("/tmp/blarg/snap").NewSnapshot()
	fatalIfError(err, "creating snapshot")
	fpset := fs.FPStore("/tmp/blarg/fps")
	fps, err := fpset.New(time.Now())
	fatalIfError(err, "creating fingerprint set")

	for s := range out {
		fe, err := backdat.NewEntry(s)
		fatalIfError(err, "making file entry")
		fps.AddFingerprint(s.Fingerprint())
		ss.AddPath(s.Path, fe)
		switch fe.Type {
		case backdat.EntryTypeDir:
			continue
		case backdat.EntryTypeSymlink:
			continue
		case backdat.EntryTypeComplete:
			b, err := ioutil.ReadFile(s.Path)
			fatalIfError(err, "opening infile: "+s.Path)
			ct, k, err := backdat.Encrypt(b)
			fatalIfError(err, "encrypting chunk")
			fatalIfError(store.Store(k.C, ct), "storing chunk")
			fe.K = k
		case backdat.EntryTypeChunked:
			infh, err := os.Open(s.Path)
			fatalIfError(err, "opening plaintext file")
			out := make(chan backdat.Chunk)
			go backdat.Slice(infh, out)
			s256 := sha256.New()
			ks := []*backdat.K{}
			for chunk := range out {
				fatalIfError(chunk.Err, "handling slice")
				has, err := store.Has(chunk.K.C)
				fatalIfError(err, "looking for chunk")
				if !has {
					fatalIfError(store.Store(chunk.K.C, chunk.Data), "storing chunk")
				}
				fmt.Fprint(s256, chunk.K.C)
				ks = append(ks, chunk.K)
			}
			id := binary.BigEndian.Uint64(s256.Sum(nil)[:8])
			ss.AddID(id, ks)
			infh.Close()
		default:
			panic(fmt.Sprint("invalid entry type: ", fe.Type))
		}
	}

	fatalIfError(ss.Close(), "closing snapshot file")
	fatalIfError(fps.Close(), "closing fingerprint file")
}
