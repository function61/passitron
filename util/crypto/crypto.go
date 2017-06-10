package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"log"
	"time"
)

func passwordTo256BitEncryptionKey(pwd string, salt []byte) [32]byte {
	started := time.Now()

	// 1.4sec @ 100k on Raspberry Pi 2
	// https://github.com/borgbackup/borg/issues/77#issuecomment-130459726
	iterationCount := 100 * 1000

	encryptionKey := pbkdf2.Key(
		[]byte(pwd),
		salt,
		iterationCount,
		32,
		sha256.New)

	if len(encryptionKey) != 32 {
		panic("returned pbkdf2 key not 32 bytes")
	}

	var ret [32]byte
	copy(ret[:], encryptionKey)

	log.Printf("Encryption key derived from password in %s", time.Since(started))

	return ret
}

// envelope = <24 bytes of nonce> <ciphertext>
func Encrypt(plaintext []byte, password string) ([]byte, error) {
	// You must use a different nonce for each message you encrypt with the
	// same key. Since the nonce here is 192 bits long, a random value
	// provides a sufficiently small probability of repeats.
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	// using Seal() nonce as PBKDF2 salt
	encryptionKey := passwordTo256BitEncryptionKey(password, nonce[:])

	nonceAndCiphertextEnvelope := []byte{}
	nonceAndCiphertextEnvelope = secretbox.Seal(nonce[:], plaintext, &nonce, &encryptionKey)

	return nonceAndCiphertextEnvelope, nil
}

func Decrypt(nonceAndCiphertextEnvelope []byte, password string) ([]byte, error) {
	// When you decrypt, you must use the same nonce and key you used to
	// encrypt the message. One way to achieve this is to store the nonce
	// alongside the encrypted message. Above, we stored the nonce in the first
	// 24 bytes of the encrypted text.
	// 24 bytes of nonce seems fine https://security.stackexchange.com/a/112592
	var decryptNonce [24]byte
	copy(decryptNonce[:], nonceAndCiphertextEnvelope[:24])

	// using Seal() nonce as PBKDF2 salt
	encryptionKey := passwordTo256BitEncryptionKey(password, decryptNonce[:])

	plaintextBytes, ok := secretbox.Open(nil, nonceAndCiphertextEnvelope[24:], &decryptNonce, &encryptionKey)
	if !ok {
		return nil, errors.New("decryption error. wrong password?")
	}

	return plaintextBytes, nil
}
