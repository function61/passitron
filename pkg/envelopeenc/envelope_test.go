package envelopeenc

import (
	"bytes"
	"crypto/rsa"
	"encoding/hex"
	"github.com/function61/gokit/assert"
	"github.com/function61/gokit/cryptoutil"
	"golang.org/x/crypto/nacl/secretbox"
	"io"
	"testing"
)

func TestEncryptAndDecrypt(t *testing.T) {
	kek1, err := cryptoutil.ParsePemPkcs1EncodedRsaPrivateKey([]byte(testKek1))
	assert.Ok(t, err)

	oneKey := []*rsa.PublicKey{&kek1.PublicKey}

	// we can observe from expected outputs that nonce is at front of EncryptedContent
	tcs := []struct {
		encryptionKey  byte
		nonce          byte
		expectedOutput string
	}{
		{
			0x00,
			0x01,
			"0101010101010101010101010101010101010101010101018a7339270718de7fb3ab5bed387b75fc3824d11162466d",
		},
		{
			0xcc, // change encryption key
			0x01,
			"010101010101010101010101010101010101010101010101336d698a0b1d33381ca943b2edd78acc9b5dc1b1e80623",
		},
		{
			0xcc,
			0x21, // change nonce
			"21212121212121212121212121212121212121212121212139eb1f77b8f42d1cdc7b75254f115678cb130ffc5cf247",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.expectedOutput, func(t *testing.T) {
			pwdEnvelope, err := encryptWithRand(
				[]byte("hunter2"),
				oneKey,
				deterministicRand(tc.encryptionKey, tc.nonce))
			assert.Ok(t, err)

			assert.EqualString(
				t,
				hex.EncodeToString(pwdEnvelope.EncryptedContent),
				tc.expectedOutput)

			nonceLen := 24

			assert.Assert(t, len(pwdEnvelope.EncryptedContent)-nonceLen == len("hunter2")+secretbox.Overhead)

			decrypted, err := pwdEnvelope.Decrypt(kek1)
			assert.Ok(t, err)

			assert.EqualString(t, string(decrypted), "hunter2")
		})
	}
}

func deterministicRand(encryptionKey byte, nonce byte) io.Reader {
	return bytes.NewBuffer(append(
		bytes.Repeat([]byte{encryptionKey}, 32),
		bytes.Repeat([]byte{nonce}, 24)...))
}

const (
	testKek1 = `-----BEGIN RSA PRIVATE KEY-----
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
