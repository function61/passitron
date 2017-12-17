package randompassword

import (
	"crypto/rand"
	"math/big"
)

const (
	az           = "abcdefghijklmnopqrstuvwxyz"
	AZ           = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers      = "0123456789"
	specialchars = "!_-=:;,."

	// numbers twice and specialchars three times to increase probability
	DefaultAlphabet = az + AZ + numbers + numbers + specialchars + specialchars + specialchars
)

func Build(alphabet string, length int) string {
	alphabetMaxIdx := big.NewInt(int64(len(alphabet) - 1))

	pwd := ""
	for i := 0; i < length; i++ {
		alphabetIdx, err := rand.Int(rand.Reader, alphabetMaxIdx)
		if err != nil {
			panic(err)
		}

		pwd += string(alphabet[alphabetIdx.Int64()])
	}

	return pwd
}
