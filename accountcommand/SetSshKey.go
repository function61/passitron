package accountcommand

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"github.com/function61/pi-security-module/accountevent"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
	"golang.org/x/crypto/ssh"
	"net/http"
)

type SetSshKeyRequest struct {
	Id            string
	SshPrivateKey string
	// public key in OpenSSH authorized_keys format
	sshPublicKeyAuthorized string
}

func (f *SetSshKeyRequest) Validate() error {
	if f.Id == "" {
		return errors.New("Id missing")
	}
	if f.SshPrivateKey == "" {
		return errors.New("SshPrivateKey missing")
	}

	// validate and re-format SSH key
	block, rest := pem.Decode([]byte(f.SshPrivateKey))
	if block == nil {
		return errors.New("Failed to parse PEM block")
	}

	if len(rest) > 0 {
		return errors.New("Extra data included in PEM content")
	}

	if block.Type != "RSA PRIVATE KEY" {
		return errors.New("Currently we only support RSA format keys")
	}

	if x509.IsEncryptedPEMBlock(block) {
		// TODO: maybe implement here in import phase
		return errors.New("We do not support encypted PEM blocks yet")
	}

	f.SshPrivateKey = string(pem.EncodeToMemory(block))

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	// convert to SSH public key
	publicKeySsh, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	f.sshPublicKeyAuthorized = string(ssh.MarshalAuthorizedKey(publicKeySsh))

	return nil
}

func HandleSetSshKeyRequest(w http.ResponseWriter, r *http.Request) {
	var req SetSshKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		util.CommandValidationError(w, r, err)
		return
	}

	state.Inst.EventLog.Append(accountevent.SshKeyAdded{
		Event:                  eventbase.NewEvent(),
		Account:                req.Id,
		Id:                     eventbase.RandomId(),
		SshPrivateKey:          req.SshPrivateKey,
		SshPublicKeyAuthorized: req.sshPublicKeyAuthorized,
	})

	util.CommandGenericSuccess(w, r)
}
