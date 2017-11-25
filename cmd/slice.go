package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const (
	ChunkSize = 1024 * 1024
)

var (
	fIn  = flag.String("in", "", "input file")
	fOut = flag.String("out", "", "output prefix")
)

type K struct {
	P string
	C string
}

func main() {
	flag.Parse()
	infh, err := os.Open(*fIn)
	if err != nil {
		log.Fatal(err)
	}

	b64 := base64.RawURLEncoding
	i := 0
	b := make([]byte, ChunkSize)
	e := make([]byte, ChunkSize)
	keys := []K{}
	for {
		n, err := infh.Read(b)
		if n > 0 {
			b = b[:n]
			sum := sha512.Sum384(b)
			iv := sum[:aes.BlockSize]
			key := sum[aes.BlockSize:]
			block, err := aes.NewCipher(key)
			if err != nil {
				log.Fatal(err)
			}
			stream := cipher.NewOFB(block, iv)
			e = e[:len(b)]
			stream.XORKeyStream(e, b)
			cSum := sha256.Sum256(e)
			cName := make([]byte, b64.EncodedLen(len(cSum)))
			b64.Encode(cName, cSum[:])
			pName := make([]byte, b64.EncodedLen(len(sum)))
			b64.Encode(pName, sum[:])

			outfh, err := os.Create(filepath.Join("/tmp", string(cName)))
			if err != nil {
				log.Fatal(err)
			}
			if _, err = outfh.Write(e); err != nil {
				log.Fatal(err)
			}
			outfh.Close()
			k := K{
				P: string(pName),
				C: string(cName),
			}
			keys = append(keys, k)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		i += 1
	}
	kj, err := json.Marshal(keys)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(kj))
}
