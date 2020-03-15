package state

import (
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/gokit/cryptoutil"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/envelopeenc"
	"github.com/function61/pi-security-module/pkg/slowcrypto"
)

var (
	ErrDecryptionKeyLocked = errors.New("decryption key locked")
)

type cryptoThingie struct {
	privateKeyEncrypted []byte          // slowcrypto(pem(pkcs1(rsaPrivateKey)))
	privateKey          *rsa.PrivateKey // gets decrypted here from privateKeyEncrypted
	publicKey           *rsa.PublicKey
}

func newCryptoThingie(rsaPublicKeyPemPkcs string, decryptionKeyEncrypted []byte) (*cryptoThingie, error) {
	pubKey, err := cryptoutil.ParsePemPkcs1EncodedRsaPublicKey([]byte(rsaPublicKeyPemPkcs))
	if err != nil {
		return nil, err
	}

	return &cryptoThingie{
		privateKeyEncrypted: decryptionKeyEncrypted,
		publicKey:           pubKey,
	}, nil
}

func (c *cryptoThingie) Encrypt(secret []byte) ([]byte, error) {
	env, err := envelopeenc.Encrypt(secret, []*rsa.PublicKey{c.publicKey})
	if err != nil {
		return nil, err
	}

	return env.Marshal()
}

func (c *cryptoThingie) UnlockDecryptionKey(pwd string) error {
	if c.privateKey != nil {
		return errors.New("UnlockDecryptionKey: already unlocked")
	}

	decryptionKey, err := slowcrypto.WithPassword(pwd).Decrypt(c.privateKeyEncrypted)
	if err != nil {
		return fmt.Errorf("UnlockDecryptionKey: %w", err)
	}

	privKey, err := cryptoutil.ParsePemPkcs1EncodedRsaPrivateKey(decryptionKey)
	if err != nil {
		return err
	}

	c.privateKey = privKey

	return nil
}

// this will be a network hop or done in a browser
func (c *cryptoThingie) Decrypt(envelopeBytes []byte) ([]byte, error) {
	if c.privateKey == nil {
		return nil, ErrDecryptionKeyLocked
	}

	env, err := envelopeenc.Unmarshal(envelopeBytes)
	if err != nil {
		return nil, err
	}

	return env.Decrypt(c.privateKey)
}

func (c *cryptoThingie) ChangeDecryptionKeyPassword(
	newPassword string,
	meta ehevent.EventMeta,
) (*domain.UserDecryptionKeyPasswordChanged, error) {
	if c.privateKey == nil {
		return nil, ErrDecryptionKeyLocked
	}

	return ExportPrivateKeyWithPassword(c.privateKey, newPassword, meta)
}

func ExportPrivateKeyWithPassword(
	privKey *rsa.PrivateKey,
	password string,
	meta ehevent.EventMeta,
) (*domain.UserDecryptionKeyPasswordChanged, error) {
	privateKeyPem := cryptoutil.MarshalPemBytes(
		x509.MarshalPKCS1PrivateKey(privKey),
		cryptoutil.PemTypeRsaPrivateKey)

	privateKeyEncrypted, err := slowcrypto.WithPassword(password).Encrypt(privateKeyPem)
	if err != nil {
		return nil, err
	}

	pubKeyPem := cryptoutil.MarshalPemBytes(
		x509.MarshalPKCS1PublicKey(&privKey.PublicKey),
		cryptoutil.PemTypeRsaPublicKey)

	return domain.NewUserDecryptionKeyPasswordChanged(
		string(pubKeyPem),
		privateKeyEncrypted,
		meta), nil
}
