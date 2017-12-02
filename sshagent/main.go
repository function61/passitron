package sshagent

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/signingapi/signingapitypes"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	sourceSocket = "/tmp/ssh-agent.sock"
)

/*	Linux / Mac: use ENV variable

	Windows: use https://github.com/131/pageantbridge
*/

// SSH agent RFC:
// 	https://tools.ietf.org/html/rfc4253#section-6.6
// alternative via virtual SmartCard crypto:
// 	https://github.com/frankmorgner/vsmartcard

var errNotImplemented = errors.New("not implemented")

/*	OpenSSH client will:

	1) List() to get list of public keys
	2) Sign(pkey, dataToSign) if server accepts any of keys returned previously
*/

// implements interface
type AgentServer struct{}

func (a AgentServer) List() ([]*agent.Key, error) {
	log.Printf("SshAgentServer: List()")

	knownKeys := []*agent.Key{}

	resp, err := http.Get("http://localhost:8080/_api/signer/publickeys")
	if err != nil {
		return knownKeys, errors.New("public keys list request failed")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return knownKeys, errors.New("failed reading body")
	}

	var output signingapitypes.PublicKeysResponse
	if err := json.Unmarshal(body, &output); err != nil {
		return knownKeys, errors.New("failed to parse JSON response")
	}

	for _, key := range output.PublicKeys {
		knownKey := &agent.Key{
			Format:  key.Format,
			Blob:    key.Blob,
			Comment: key.Comment,
		}

		knownKeys = append(knownKeys, knownKey)
	}

	return knownKeys, nil
}

func (a AgentServer) Sign(key ssh.PublicKey, data []byte) (*ssh.Signature, error) {
	log.Printf("SshAgentServer: Sign()")

	reqJson, _ := json.Marshal(signingapitypes.SignRequestInput{
		PublicKey: key.Marshal(),
		Data:      data,
	})

	resp, err := http.Post(
		"http://localhost:8080/_api/signer/sign",
		"application/json",
		bytes.NewReader(reqJson))
	if err != nil {
		return nil, errors.New("request failed")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed reading body")
	}

	var output signingapitypes.SignResponse
	if err := json.Unmarshal(body, &output); err != nil {
		return nil, errors.New("failed to parse JSON response")
	}

	return output.Signature, nil
}

func (a AgentServer) Add(key agent.AddedKey) error {
	log.Printf("SshAgentServer: Add()")

	return errNotImplemented
}

func (a AgentServer) Remove(key ssh.PublicKey) error {
	log.Printf("SshAgentServer: Remove()")

	return errNotImplemented
}

func (a AgentServer) RemoveAll() error {
	log.Printf("SshAgentServer: RemoveAll()")

	return errNotImplemented
}

func (a AgentServer) Lock(passphrase []byte) error {
	log.Printf("SshAgentServer: Lock()")

	return errNotImplemented
}

func (a AgentServer) Unlock(passphrase []byte) error {
	log.Printf("SshAgentServer: Unlock()")

	return errNotImplemented
}

func (a AgentServer) Signers() ([]ssh.Signer, error) {
	log.Printf("SshAgentServer: Signers()")

	return []ssh.Signer{}, errNotImplemented
}

var agentServer = AgentServer{}

func checkSocketExistence() {
	_, err := os.Stat(sourceSocket)
	if err == nil { // socket exists
		if err := os.Remove(sourceSocket); err != nil {
			log.Fatalf("sshagent: remove error: %s", err.Error())
		}
	} else if !os.IsNotExist(err) { // some other error than not exists
		log.Fatalf("sshagent: unexpected Stat() error: %s", err.Error())
	}
}

func handleOneClient(client net.Conn) {
	log.Printf("sshagent: client connected")

	agent.ServeAgent(agentServer, client)
}

func Run() {
	checkSocketExistence()

	log.Printf("sshagent: listening at %s", sourceSocket)
	log.Printf("sshagent: pro tip $ export SSH_AUTH_SOCK=\"%s\"", sourceSocket)

	socketListener, err := net.Listen("unix", sourceSocket)
	if err != nil {
		log.Fatalf("sshagent: sock listen error: %s", err.Error())
	}

	for {
		// intentionally only supporting sequential connections for now
		client, err := socketListener.Accept()
		if err != nil {
			log.Printf("sshagent: Accept() error: %s", err.Error())
			continue
		}

		go handleOneClient(client)
	}
}
