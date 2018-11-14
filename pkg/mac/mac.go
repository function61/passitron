package mac

import (
	"crypto/sha1"
	"errors"
	"fmt"
)

var ErrMacValidationFailed = errors.New("mac validation failed")

type Mac struct {
	key     string
	message string
}

func New(key string, message string) *Mac {
	return &Mac{key: key, message: message}
}

func (m *Mac) Authenticate(givenMac string) error {
	if m.Sign() != givenMac {
		return ErrMacValidationFailed
	}

	return nil
}

func (m *Mac) Sign() string {
	keyAndMessageCombined := []byte(m.key + ":" + m.message)

	sumHex := fmt.Sprintf("%x", sha1.Sum(keyAndMessageCombined))

	// cap to length of 16 for prettier URLs. we still have 64 bits of entropy
	return sumHex[0:16]
}
