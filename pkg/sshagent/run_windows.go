// +build windows

package sshagent

import (
	"github.com/function61/gokit/logex"
	"github.com/microsoft/go-winio"
	"log"
)

const (
	// we're not OpenSSH but we need to use the name the SSH binary expects to connect to
	pipeName = `\\.\pipe\openssh-ssh-agent`
)

func run(agentServer *AgentServer, logger *log.Logger) error {
	logl := logex.Levels(logger)

	listener, err := winio.ListenPipe(pipeName, nil)
	if err != nil {
		return err
	}

	logl.Info.Printf("listening at %s", pipeName)

	clientHandlerLogger := logex.Levels(logex.Prefix("handleOneClient", logger))

	for {
		client, err := listener.Accept()
		if err != nil {
			return err
		}

		go agentServer.handleOneClient(client, clientHandlerLogger)
	}
}
