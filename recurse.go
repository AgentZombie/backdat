package backdat

import (
	"os"
	"path/filepath"
)

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
