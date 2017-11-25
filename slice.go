package backdat

import (
	"io"

	"github.com/pkg/errors"
)

type Chunk struct {
	Data []byte
	K    *K
	Err  error
}

func Slice(r io.Reader, out chan Chunk) {
	defer close(out)
	b := make([]byte, ChunkSize)
	for {
		n, err := r.Read(b)
		if err != nil {
			if err == io.EOF {
				return
			}
			out <- Chunk{Err: errors.Wrap(err, "reading plaintext")}
			return
		}
		ct, k, err := Encrypt(b[:n])
		if err != nil {
			out <- Chunk{Err: errors.Wrap(err, "encrypting chunk")}
			return
		}
		out <- Chunk{Data: ct, K: k}
	}
}
