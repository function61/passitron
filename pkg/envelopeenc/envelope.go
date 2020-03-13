// Envelope encryption - envelope contains secret content encrypted with NaCl secretbox
// symmetric key, and that key is separately encrypted for each RSA public key recipient.
package envelopeenc

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/function61/gokit/cryptoutil"
	"golang.org/x/crypto/nacl/secretbox"
	"io"
)

type Envelope struct {
	KeySlots         []envelopeKeySlot `json:"key_slots"`
	EncryptedContent []byte            `json:"content"` // nonce || secretbox_ciphertext
}

type envelopeKeySlot struct {
	KekId        string `json:"kek_id"`        // SHA256-fingerprint of RSA public key
	DekEncrypted []byte `json:"dek_encrypted"` // RSA_OAEP_SHA256(kekPub, secretboxSecretKey)
}

func Encrypt(plaintext []byte, keks []*rsa.PublicKey) (*Envelope, error) {
	return encryptWithRand(plaintext, keks, rand.Reader)
}

func encryptWithRand(plaintext []byte, keks []*rsa.PublicKey, cryptoRandReader io.Reader) (*Envelope, error) {
	var secretKey [32]byte
	if _, err := io.ReadFull(cryptoRandReader, secretKey[:]); err != nil {
		return nil, err
	}

	var nonce [24]byte
	if _, err := io.ReadFull(cryptoRandReader, nonce[:]); err != nil {
		return nil, err
	}

	keySlots := []envelopeKeySlot{}

	for _, kek := range keks {
		keySlot, err := makeKeySlot(secretKey[:], kek)
		if err != nil {
			return nil, err
		}

		keySlots = append(keySlots, *keySlot)
	}

	// return is basically append(nonce, ciphertext...)
	nonceAndCiphertext := secretbox.Seal(nonce[:], plaintext, &nonce, &secretKey)

	return &Envelope{
		KeySlots:         keySlots,
		EncryptedContent: nonceAndCiphertext,
	}, nil
}

func (e *Envelope) Decrypt(privKey *rsa.PrivateKey) ([]byte, error) {
	kekId, err := cryptoutil.Sha256FingerprintForPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, err
	}

	for _, slot := range e.KeySlots {
		if slot.KekId == kekId {
			return e.decryptWithSlot(&slot, privKey)
		}
	}

	return nil, fmt.Errorf("no slot found for %s", kekId)
}

func (e *Envelope) decryptWithSlot(slot *envelopeKeySlot, privKey *rsa.PrivateKey) ([]byte, error) {
	dek, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privKey, slot.DekEncrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("decryptWithSlot DecryptOAEP: %v", err)
	}

	var nonce [24]byte
	copy(nonce[:], e.EncryptedContent[:24])

	var dekStatic [32]byte
	copy(dekStatic[:], dek)

	plaintext := []byte{}
	plaintext, ok := secretbox.Open(plaintext, e.EncryptedContent[24:], &nonce, &dekStatic)
	if !ok {
		return nil, errors.New("secretbox.Open failed")
	}

	return plaintext, nil
}

func makeKeySlot(dek []byte, kekPub *rsa.PublicKey) (*envelopeKeySlot, error) {
	kekId, err := cryptoutil.Sha256FingerprintForPublicKey(kekPub)
	if err != nil {
		return nil, err
	}

	dekCiphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, kekPub, dek, nil)
	if err != nil {
		return nil, err
	}

	return &envelopeKeySlot{
		KekId:        kekId,
		DekEncrypted: dekCiphertext,
	}, nil
}
