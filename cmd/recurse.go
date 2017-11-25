package main

import (
	"encoding/json"
	"os"

	"backdat"
)

func main() {
	out := make(chan backdat.Stat)
	je := json.NewEncoder(os.Stdout)
	go func() {
		backdat.Recurse("/home/jason/go/", out)
		close(out)
	}()
	for s := range out {
		fe := backdat.FileEntry{
			Stat: s,
		}
		if !s.IsDir {
			if s.Size < backdat.ChunkSize {
				fe.K = &backdat.K{P: "blarg"}
			} else {
				fe.ID = -1
			}
		}
		je.Encode(&fe)
	}
}
