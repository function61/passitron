package signingapitypes

import (
	"golang.org/x/crypto/ssh"
)

type PublicKeyResponseItem struct {
	Format  string
	Blob    []byte
	Comment string
}

type PublicKeysResponse struct {
	PublicKeys []PublicKeyResponseItem
}

func NewPublicKeysResponse() PublicKeysResponse {
	return PublicKeysResponse{
		PublicKeys: []PublicKeyResponseItem{},
	}
}

type SignRequestInput struct {
	PublicKey []byte
	Data      []byte
}

type SignResponse struct {
	Signature *ssh.Signature
}
