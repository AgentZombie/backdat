package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
)

const (
	ChunkSize = 1024 * 1024
)

var (
	fMan = flag.String("man", "", "manifest file")
)

type K struct {
	P string
	C string
}

func main() {
	flag.Parse()
	manfh, err := os.Open(*fMan)
	if err != nil {
		log.Fatal(err)
	}
	b64 := base64.RawURLEncoding
	jd := json.NewDecoder(manfh)
	keys := []K{}
	if err = jd.Decode(&keys); err != nil {
		log.Fatal(err)
	}
	outfh, err := os.Create("/tmp/out")
	if err != nil {
		log.Fatal(err)
	}
	in := make([]byte, ChunkSize)
	out := make([]byte, ChunkSize)
	for _, k := range keys {
		infh, err := os.Open(filepath.Join("/tmp", k.C))
		if err != nil {
			log.Fatal(err)
		}
		n, err := infh.Read(in)
		if n > 0 {
			in = in[:n]
			cSumExpected := make([]byte, b64.DecodedLen(len([]byte(k.C))))
			b64.Decode(cSumExpected, []byte(k.C))
			cSum := sha256.Sum256(in[:])
			if !bytes.Equal(cSumExpected, cSum[:]) {
				log.Fatal("CText hash mismatch on ", k.C)
			}
			pSum := make([]byte, b64.DecodedLen(len([]byte(k.P))))
			b64.Decode(pSum, []byte(k.P))
			iv := pSum[:aes.BlockSize]
			key := pSum[aes.BlockSize:]
			block, err := aes.NewCipher(key)
			if err != nil {
				log.Fatal(err)
			}
			stream := cipher.NewOFB(block, iv)
			out = out[:len(in)]
			stream.XORKeyStream(out, in)
			outSum := sha512.Sum384(out)
			if !bytes.Equal(pSum, outSum[:]) {
				log.Fatal("PText hash mismatch on ", k.C)
			}
			if _, err := outfh.Write(out); err != nil {
				log.Fatal(err)
			}
		}
		infh.Close()
	}
	outfh.Close()
}
