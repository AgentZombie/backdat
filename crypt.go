package backdat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"

	"github.com/pkg/errors"
)

const (
	ChunkSize = 1024 * 1024
)

var (
	b64 = base64.RawURLEncoding
)

func Encrypt(pt []byte) ([]byte, *K, error) {
	sum := sha512.Sum384(pt)
	iv := sum[:aes.BlockSize]
	key := sum[aes.BlockSize:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, errors.Wrap(err, "creating aes context")
	}
	stream := cipher.NewOFB(block, iv)
	ct := make([]byte, len(pt))
	stream.XORKeyStream(ct, pt)
	cSum := sha256.Sum256(ct)
	cName := make([]byte, b64.EncodedLen(len(cSum)))
	b64.Encode(cName, cSum[:])
	pName := make([]byte, b64.EncodedLen(len(sum)))
	b64.Encode(pName, sum[:])
	k := &K{
		P: string(pName),
		C: string(cName),
	}
	return ct, k, nil
}

func Decrypt(ct []byte, k *K) ([]byte, error) {
	cSumExpected := make([]byte, b64.DecodedLen(len([]byte(k.C))))
	b64.Decode(cSumExpected, []byte(k.C))
	cSum := sha256.Sum256(ct)
	if !bytes.Equal(cSumExpected, cSum[:]) {
		return nil, errors.New("ciphertext hash mismatch")
	}
	pSum := make([]byte, b64.DecodedLen(len([]byte(k.P))))
	b64.Decode(pSum, []byte(k.P))
	iv := pSum[:aes.BlockSize]
	key := pSum[aes.BlockSize:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "creating aes context")
	}
	stream := cipher.NewOFB(block, iv)
	pt := make([]byte, len(ct))
	stream.XORKeyStream(pt, ct)
	outSum := sha512.Sum384(pt)
	if !bytes.Equal(pSum, outSum[:]) {
		return nil, errors.New("plaintext hash mismatch")
	}
	return pt, nil
}
