package slowcrypto

import (
	"bytes"
	"encoding/hex"
	"github.com/function61/gokit/assert"
	"io"
	"testing"
)

func TestPbkdf2Sha256100kDerive(t *testing.T) {
	assert.EqualString(t, hex.EncodeToString(
		Pbkdf2Sha256100kDerive([]byte("hunter2"), []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF})),
		"fc970d4cd9541ea4520daaea54dcd0dde0f5c4dcb4f70aabf7625df8e012da79")

	// changing salt changes result
	assert.EqualString(t, hex.EncodeToString(
		Pbkdf2Sha256100kDerive([]byte("hunter2"), []byte{0xAB, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF})),
		"f0196fb714d753071904eb4128c59704b631e8e00f7f1799ca454ac7c4a34437")

	// changing key changes result
	assert.EqualString(t, hex.EncodeToString(
		Pbkdf2Sha256100kDerive([]byte("hunter1"), []byte{0xAB, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF})),
		"3876c39131f03ab1dfdc0b3f3af06a3477436a65cd2b7e8cc6d396b9ee0e33cb")
}

func TestEncryptAndDecrypt(t *testing.T) {
	ciphertext, err := WithPassword("hunter2").encryptWithRandom(
		[]byte("the germans are coming"),
		constRandom())
	assert.Ok(t, err)

	_, err = WithPassword("incorrect").Decrypt(ciphertext)
	assert.EqualString(t, err.Error(), "decryption error. wrong password?")

	plaintext, err := WithPassword("hunter2").Decrypt(ciphertext)
	assert.Ok(t, err)

	assert.EqualString(t, string(plaintext), "the germans are coming")
}

func constRandom() io.Reader {
	return bytes.NewBuffer([]byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		0x10, 0x11, 0x12, 0x13, 0x24, 0x25, 0x26, 0x27,
	})
}
