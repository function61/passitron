package sshagent

import (
	"bytes"
	"crypto/rand"
	"errors"
	"github.com/function61/pi-security-module/accountevent"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
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

const (
	tcpListenAddr = "0.0.0.0:8096"
)

/*	OpenSSH client will:

	1) List() to get list of public keys
	2) Sign(pkey, dataToSign) if server accepts any of keys returned previously
*/

// implements interface
type AgentProxy struct{}

func (a AgentProxy) List() ([]*agent.Key, error) {
	log.Printf("SshAgentServer: List()")

	knownKeys := []*agent.Key{}

	if !state.Inst.IsUnsealed() {
		log.Printf("SshAgentServer: returned empty list because state is sealed")

		return knownKeys, nil
	}

	for _, account := range state.Inst.State.Accounts {
		for _, secret := range account.Secrets {
			if secret.SshPrivateKey == "" {
				continue
			}

			log.Printf("SshAgentServer: List() candidate %s", account.Title)

			signer, err := ssh.ParsePrivateKey([]byte(secret.SshPrivateKey))
			if err != nil {
				panic(err)
			}

			publicKey := signer.PublicKey()

			knownKey := &agent.Key{
				Format:  publicKey.Type(),
				Blob:    publicKey.Marshal(),
				Comment: account.Title,
			}

			knownKeys = append(knownKeys, knownKey)
		}
	}

	return knownKeys, nil
}

func (a AgentProxy) Sign(key ssh.PublicKey, data []byte) (*ssh.Signature, error) {
	log.Printf("SshAgentServer: Sign()")

	if !state.Inst.IsUnsealed() {
		log.Printf("SshAgentServer: Sign() return because state is sealed")

		return nil, errors.New("state is sealed")
	}

	keyMarshaled := key.Marshal()

	for _, account := range state.Inst.State.Accounts {
		for _, secret := range account.Secrets {
			if secret.SshPrivateKey == "" {
				continue
			}

			log.Printf("SshAgentServer: Sign() candidate %s", account.Title)

			signer, err := ssh.ParsePrivateKey([]byte(secret.SshPrivateKey))
			if err != nil {
				panic(err)
			}

			publicKey := signer.PublicKey()

			// TODO: is there better way to compare than marshal result?
			if !bytes.Equal(keyMarshaled, publicKey.Marshal()) {
				log.Printf("SshAgentServer: Sign(): skipping candidate")
				continue
			}

			// found it

			sig, err := signer.Sign(rand.Reader, data)
			if err != nil {
				log.Printf("SshAgentServer: Sign() error: %s", err.Error())
				return nil, err
			}

			util.ApplyEvent(accountevent.SecretUsed{
				Event:   eventbase.NewEvent(),
				Account: account.Id,
				Type:    accountevent.SecretUsedTypeSshSigning,
			})

			return sig, nil
		}
	}

	notFoundErr := errors.New("privkey not found by pubkey")

	log.Printf("SshAgentServer: Sign(): %s", notFoundErr.Error())

	return nil, notFoundErr
}

func (a AgentProxy) Add(key agent.AddedKey) error {
	log.Printf("SshAgentServer: Add()")

	return errNotImplemented
}

func (a AgentProxy) Remove(key ssh.PublicKey) error {
	log.Printf("SshAgentServer: Remove()")

	return errNotImplemented
}

func (a AgentProxy) RemoveAll() error {
	log.Printf("SshAgentServer: RemoveAll()")

	return errNotImplemented
}

func (a AgentProxy) Lock(passphrase []byte) error {
	log.Printf("SshAgentServer: Lock()")

	return errNotImplemented
}

func (a AgentProxy) Unlock(passphrase []byte) error {
	log.Printf("SshAgentServer: Unlock()")

	return errNotImplemented
}

func (a AgentProxy) Signers() ([]ssh.Signer, error) {
	log.Printf("SshAgentServer: Signers()")

	return []ssh.Signer{}, errNotImplemented
}

func Start() {
	agentProxy := AgentProxy{}

	log.Printf("SshAgentMain: starting to listen on %s", tcpListenAddr)

	// netListener, err := net.Listen("unix", tcpListenAddr)
	netListener, err := net.Listen("tcp", tcpListenAddr)
	if err != nil {
		log.Fatalf("SshAgentMain: sock listen error: %s", err.Error())
	}

	for {
		// intentionally only supporting sequential connections for now
		fd, err := netListener.Accept()
		if err != nil {
			log.Printf("SshAgentMain: Accept() error: %s", err.Error())
			continue
		}

		log.Printf("SshAgentMain: client connected")

		agent.ServeAgent(agentProxy, fd)

		log.Printf("SshAgentMain: client disconnected")
	}
}
