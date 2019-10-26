// +build !windows

package sshagent

import (
	"fmt"
	"github.com/function61/gokit/fileexists"
	"github.com/function61/gokit/logex"
	"log"
	"net"
	"os"
)

const (
	sourceSocket = "/tmp/ssh-agent.sock"
)

func run(agentServer *AgentServer, logger *log.Logger) error {
	logl := logex.Levels(logger)

	if err := removeFileIfExists(sourceSocket); err != nil {
		return fmt.Errorf("removeFileIfExists: %s", err.Error())
	}

	socketListener, err := net.Listen("unix", sourceSocket)
	if err != nil {
		return fmt.Errorf("Listen(): %s", err.Error())
	}

	logl.Info.Printf("listening at %s", sourceSocket)
	logl.Info.Printf("pro tip $ export SSH_AUTH_SOCK=\"%s\"", sourceSocket)

	clientHandlerLogger := logex.Levels(logex.Prefix("handleOneClient", logger))

	for {
		client, err := socketListener.Accept()
		if err != nil {
			return fmt.Errorf("Accept(): %s", err.Error())
		}

		go agentServer.handleOneClient(client, clientHandlerLogger)
	}
}

func removeFileIfExists(path string) error {
	exists, err := fileexists.Exists(path)
	if err != nil {
		return err
	}

	if exists {
		return os.Remove(path)
	} else {
		return nil
	}
}
