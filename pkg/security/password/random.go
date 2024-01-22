package password

import (
	"crypto/rand"
	"math/big"
)

const (
	DefaultIdLength   = 25
	DefaultIdAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func RandomString() string {
	b := make([]byte, DefaultIdLength)
	max := big.NewInt(int64(len(DefaultIdAlphabet)))

	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}
		b[i] = DefaultIdAlphabet[n.Int64()]
	}

	return string(b)
}
