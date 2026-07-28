package password

import (
	"crypto/rand"
	"math/big"
)

const (
	DefaultIDLength   = 25
	DefaultIDAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func RandomString() string {
	b := make([]byte, DefaultIDLength)
	max := big.NewInt(int64(len(DefaultIDAlphabet)))

	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}
		b[i] = DefaultIDAlphabet[n.Int64()]
	}

	return string(b)
}
