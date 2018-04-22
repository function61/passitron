package main

import (
	"github.com/function61/eventhorizon/util/ass"
	"testing"
)

var workingSshKey = `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQDYHLgXErTPTKGGwY/sQ6b7dl7zVm5B/nfGlqfejVb10gAkcO1N
RCCsKPCYirbc//qQo+skpGh6/zf1OPJTq9c4ZaAdgh33tHlG42/lV1l0ehxMDrra
eSrL1RgutQsNMB+Zem5zuRLZ3v3pEyVY24nycF8r8Bprgf1xpleR1KJtBwIBJQKB
gCMLk3lc+rnVE0ZI5ugK+HvOAY8+cr6X979Ws3AymHrjyKv2owWcWFNE6L7KYte6
zq+r4PEvauODVS6vSeQOBzlHhop49DiYzjIx/Slzmm4FoE4WehBY/5l2xw901HoW
nv1FJkXsEdWvtu8bw5GFJOlzBYCkpNRoEBk5myOl50DNAkEA+UdJR/tXyGN87c+e
9hZcvlHu0VFkE/756z2N/ysmCZMBZmdQp7YuCopavqSftnBVqlJoRr7piXi/Q8Rb
QJy/TwJBAN3wflBAmD3I6cFccW21cUPDJl14vEBY9OK5wWVsS5sNfj7wc+GZW242
ISlKt8Vgp9YVf7INuXbMFtSr2r+eSMkCQQChsbL+QisaMrHmXSjW+b+eC6HT4cRf
/1YAX0dZaBisQ63hjx+PYWn4/8wooiJogDeREtva3LMovQZxJWufiES9AkA1/Dpm
jEC1FTHww3WJY3xqbb05VLgrU+iKLS8K1ScltyyL2Z+llAF7rE1BZTOetqVdlobY
SIcPDwx464g8cpwVAkEAmXAzw61rhkMQDNaAMpNwINyhd1LM0nakzPJN4NeB5qJh
vD2QakbdLBUy2JF2E2GHmGyTXQ6yp4rWgcCVPeeFRw==
-----END RSA PRIVATE KEY-----`

func TestFails(t *testing.T) {
	pemParsingFails := SetSshKeyRequest{
		Id:            "id_1",
		SshPrivateKey: "invalid",
	}

	ass.EqualString(t, pemParsingFails.Validate().Error(), "Failed to parse PEM block")
}

func TestSucceeds(t *testing.T) {
	succeeds := SetSshKeyRequest{
		Id:            "id_1",
		SshPrivateKey: workingSshKey,
	}

	ass.True(t, succeeds.Validate() == nil)

	ass.EqualString(t,
		succeeds.sshPublicKeyAuthorized,
		"ssh-rsa AAAAB3NzaC1yc2EAAAABJQAAAIEA2By4FxK0z0yhhsGP7EOm+3Ze81ZuQf53xpan3o1W9dIAJHDtTUQgrCjwmIq23P/6kKPrJKRoev839TjyU6vXOGWgHYId97R5RuNv5VdZdHocTA662nkqy9UYLrULDTAfmXpuc7kS2d796RMlWNuJ8nBfK/Aaa4H9caZXkdSibQc=\n")
}
