package auth

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
)

func GenerateKey() ([]byte, []byte, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	marshaledPrivKey, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return nil, nil, err
	}
	marshaledPubKey, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	pemPrivKey := &bytes.Buffer{}
	pemPubKey := &bytes.Buffer{}

	if err := pem.Encode(pemPrivKey, &pem.Block{Type: "PRIVATE KEY", Bytes: marshaledPrivKey}); err != nil {
		return nil, nil, err
	}

	if err := pem.Encode(pemPubKey, &pem.Block{Type: "PUBLIC KEY", Bytes: marshaledPubKey}); err != nil {
		return nil, nil, err
	}

	return pemPrivKey.Bytes(), pemPubKey.Bytes(), nil
}
