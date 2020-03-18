package sshagent

import (
	"context"
	"errors"
	"github.com/function61/gokit/ezhttp"
	"github.com/function61/gokit/logex"
	"github.com/function61/pi-security-module/pkg/signingapi"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"log"
	"net"
)

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
	endpoints   *signingapi.RestClientUrlBuilder
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
		a.endpoints.GetPublicKeys(),
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
		a.endpoints.Sign(),
		ezhttp.AuthBearer(a.bearerToken),
		ezhttp.SendJson(&req),
		ezhttp.RespondsJson(&res, false)); err != nil {
		return nil, err
	}

	return &ssh.Signature{
		Format: res.Format,
		Blob:   res.Blob,
	}, nil
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

// not part of agent API
func (a *AgentServer) handleOneClient(client net.Conn, logl *logex.Leveled) {
	logl.Info.Printf("connected")
	defer logl.Info.Printf("disconnected")

	if err := agent.ServeAgent(a, client); err != nil {
		logl.Error.Println(err)
	}
}

func Run(
	ctx context.Context,
	baseurl string,
	token string,
	logger *log.Logger,
) error {
	agentServer := &AgentServer{
		endpoints:   signingapi.NewRestClientUrlBuilder(baseurl),
		bearerToken: token,
		logl:        logex.Levels(logex.Prefix("AgentServer", logger)),
	}

	return run(ctx, agentServer, logger)
}
