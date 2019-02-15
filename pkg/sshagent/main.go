package sshagent

import (
	"context"
	"errors"
	"fmt"
	"github.com/function61/gokit/ezhttp"
	"github.com/function61/gokit/logex"
	"github.com/function61/pi-security-module/pkg/signingapi"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"log"
	"net"
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

// implements golang.org/x/crypto/ssh/agent.Agent
type AgentServer struct {
	baseUrl     string
	bearerToken string
	logl        *logex.Leveled
}

func (a *AgentServer) List() ([]*agent.Key, error) {
	a.logl.Debug.Printf("List()")

	knownKeys := []*agent.Key{}

	ctx, cancel := context.WithTimeout(context.TODO(), ezhttp.DefaultTimeout10s)
	defer cancel()

	output := signingapi.PublicKeysOutput{}
	if _, err := ezhttp.Get(
		ctx,
		a.baseUrl+"/_api/signer/publickeys",
		ezhttp.AuthBearer(a.bearerToken),
		ezhttp.RespondsJson(&output, false)); err != nil {
		return knownKeys, err
	}

	for _, key := range output {
		knownKey := &agent.Key{
			Format:  key.Format,
			Blob:    key.Blob,
			Comment: key.Comment,
		}

		knownKeys = append(knownKeys, knownKey)
	}

	return knownKeys, nil
}

func (a *AgentServer) Sign(key ssh.PublicKey, data []byte) (*ssh.Signature, error) {
	a.logl.Debug.Printf("Sign()")

	req := signingapi.SignRequestInput{
		PublicKey: key.Marshal(),
		Data:      data,
	}
	res := signingapi.Signature{}

	ctx, cancel := context.WithTimeout(context.TODO(), ezhttp.DefaultTimeout10s)
	defer cancel()

	if _, err := ezhttp.Post(
		ctx,
		a.baseUrl+"/_api/signer/sign",
		ezhttp.AuthBearer(a.bearerToken),
		ezhttp.SendJson(&req),
		ezhttp.RespondsJson(&res, false)); err != nil {
		return nil, err
	}

	sshSig := ssh.Signature(res) // structs are type-compatible
	return &sshSig, nil
}

func (a *AgentServer) Add(key agent.AddedKey) error {
	a.logl.Debug.Printf("Add()")

	return errNotImplemented
}

func (a *AgentServer) Remove(key ssh.PublicKey) error {
	a.logl.Debug.Printf("Remove()")

	return errNotImplemented
}

func (a *AgentServer) RemoveAll() error {
	a.logl.Debug.Printf("RemoveAll()")

	return errNotImplemented
}

func (a *AgentServer) Lock(passphrase []byte) error {
	a.logl.Debug.Printf("Lock()")

	return errNotImplemented
}

func (a *AgentServer) Unlock(passphrase []byte) error {
	a.logl.Debug.Printf("Unlock()")

	return errNotImplemented
}

func (a *AgentServer) Signers() ([]ssh.Signer, error) {
	a.logl.Debug.Printf("Signers()")

	return []ssh.Signer{}, errNotImplemented
}

func removeFileIfExists(path string) error {
	_, err := os.Stat(path)
	if err == nil { // socket exists
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("Remove: %s", err.Error())
		}
	} else if !os.IsNotExist(err) { // some other error than not exists
		return fmt.Errorf("Stat(): %s", err.Error())
	}

	return nil
}

func handleOneClient(client net.Conn, server *AgentServer, logl *logex.Leveled) {
	logl.Info.Printf("Connected")
	defer logl.Info.Printf("Disconnected")

	agent.ServeAgent(server, client)
}

func Run(baseurl string, token string, logger *log.Logger) error {
	if err := removeFileIfExists(sourceSocket); err != nil {
		return fmt.Errorf("removeFileIfExists: %s", err.Error())
	}

	logl := logex.Levels(logger)

	agentServer := AgentServer{
		baseUrl:     baseurl,
		bearerToken: token,
		logl:        logex.Levels(logex.Prefix("AgentServer", logger)),
	}

	logl.Info.Printf("Listening at %s", sourceSocket)
	logl.Info.Printf("Pro tip $ export SSH_AUTH_SOCK=\"%s\"", sourceSocket)

	socketListener, err := net.Listen("unix", sourceSocket)
	if err != nil {
		return fmt.Errorf("Listen(): %s", err.Error())
	}

	clientHandlerLogger := logex.Levels(logex.Prefix("handleOneClient", logger))

	for {
		client, err := socketListener.Accept()
		if err != nil {
			return fmt.Errorf("Accept(): %s", err.Error())
		}

		go handleOneClient(client, &agentServer, clientHandlerLogger)
	}
}
