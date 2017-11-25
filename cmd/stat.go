package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	fRoot = flag.String("root", "", "base path to back up")
)

type Stat struct {
	Path    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	IsDir   bool
}

func Recurse(path string, out chan Stat) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	fis, err := f.Readdir(0)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		s := Stat{
			Path:    filepath.Join(path, fi.Name()),
			Size:    fi.Size(),
			Mode:    fi.Mode(),
			ModTime: fi.ModTime(),
			IsDir:   fi.IsDir(),
		}
		out <- s
		if fi.IsDir() {
			if err = Recurse(s.Path, out); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	flag.Parse()

	if *fRoot == "" {
		log.Fatal("must specify -root")
	}

	out := make(chan Stat)
	go func() {
		if err := Recurse(*fRoot, out); err != nil {
			log.Fatal("error in recurse: ", err)
		}
		close(out)
	}()
	for s := range out {
		b, err := json.Marshal(&s)
		if err != nil {
			log.Fatal("error encoding fi: ", err)
		}
		log.Print(string(b))
	}
}
