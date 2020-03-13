package state

import (
	"crypto/rsa"
	"errors"
	"github.com/function61/gokit/cryptoutil"
	"github.com/function61/pi-security-module/pkg/envelopeenc"
)

const (
	hardcodedTemporaryKek = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCqGKukO1De7zhZj6+H0qtjTkVxwTCpvKe4eCZ0FPqri0cb2JZfXJ/DgYSF6vUp
wmJG8wVQZKjeGcjDOL5UlsuusFncCzWBQ7RKNUSesmQRMSGkVb1/3j+skZ6UtW+5u09lHNsj6tQ5
1s1SPrCBkedbNf0Tp0GbMJDyR4e9T04ZZwIDAQABAoGAFijko56+qGyN8M0RVyaRAXz++xTqHBLh
3tx4VgMtrQ+WEgCjhoTwo23KMBAuJGSYnRmoBZM3lMfTKevIkAidPExvYCdm5dYq3XToLkkLv5L2
pIIVOFMDG+KESnAFV7l2c+cnzRMW0+b6f8mR1CJzZuxVLL6Q02fvLi55/mbSYxECQQDeAw6fiIQX
GukBI4eMZZt4nscy2o12KyYner3VpoeE+Np2q+Z3pvAMd/aNzQ/W9WaI+NRfcxUJrmfPwIGm63il
AkEAxCL5HQb2bQr4ByorcMWm/hEP2MZzROV73yF41hPsRC9m66KrheO9HPTJuo3/9s5p+sqGxOlF
L0NDt4SkosjgGwJAFklyR1uZ/wPJjj611cdBcztlPdqoxssQGnh85BzCj/u3WqBpE2vjvyyvyI5k
X6zk7S0ljKtt2jny2+00VsBerQJBAJGC1Mg5Oydo5NwD6BiROrPxGo2bpTbu/fhrT8ebHkTz2epl
U9VQQSQzY1oZMVX8i1m5WUTLPz2yLJIBQVdXqhMCQBGoiuSoSjafUhV7i1cEGpb88h5NBYZzWXGZ
37sJ5QsW+sJyoNde3xH8vdXhzU7eT82D6X/scw9RZz+/6rCJ4p0=
-----END RSA PRIVATE KEY-----`
)

type cryptoThingie struct {
	unlocked   bool
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func newCryptoThingie() (*cryptoThingie, error) {
	privKey, err := cryptoutil.ParsePemPkcs1EncodedRsaPrivateKey([]byte(hardcodedTemporaryKek))
	if err != nil {
		return nil, err
	}

	return &cryptoThingie{
		privateKey: privKey,
		publicKey:  &privKey.PublicKey,
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
	if pwd == "opensesame" {
		c.unlocked = true

		return nil
	} else {
		return errors.New("UnlockDecryptionKey failed")
	}
}

// this will be a network hop or done in a browser
func (c *cryptoThingie) Decrypt(envelopeBytes []byte) ([]byte, error) {
	if !c.unlocked {
		return nil, errors.New("decryption key not unlocked")
	}

	env, err := envelopeenc.Unmarshal(envelopeBytes)
	if err != nil {
		return nil, err
	}

	return env.Decrypt(c.privateKey)
}
