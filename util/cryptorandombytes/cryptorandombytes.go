package cryptorandombytes

import (
	"crypto/rand"
	"encoding/hex"
)

func Hex(bytesLen int) string {
	randBytes := make([]byte, bytesLen)

	if _, err := rand.Read(randBytes); err != nil {
		panic(err)
	}

	return hex.EncodeToString(randBytes)
}
