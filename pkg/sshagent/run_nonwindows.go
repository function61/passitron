// +build !windows

package sshagent

import (
	"context"
	"fmt"
	"github.com/function61/gokit/fileexists"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/taskrunner"
	"log"
	"net"
	"os"
)

const (
	sourceSocket = "/tmp/ssh-agent.sock"
)

func run(ctx context.Context, agentServer *AgentServer, logger *log.Logger) error {
	logl := logex.Levels(logger)

	if err := removeFileIfExists(sourceSocket); err != nil {
		return fmt.Errorf("removeFileIfExists: %s", err.Error())
	}

	socketListener, err := net.Listen("unix", sourceSocket)
	if err != nil {
		return fmt.Errorf("Listen(): %s", err.Error())
	}

	tasks := taskrunner.New(ctx, logger)

	tasks.Start("listener "+sourceSocket, func(ctx context.Context, _ string) error {
		logl.Info.Printf("listening at %s", sourceSocket)
		logl.Info.Printf("pro tip $ export SSH_AUTH_SOCK=\"%s\"", sourceSocket)

		clientHandlerLogger := logex.Levels(logex.Prefix("handleOneClient", logger))

		for {
			client, err := socketListener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return nil // expected Accept() error
				default:
					return err
				}
			}

			go agentServer.handleOneClient(client, clientHandlerLogger)
		}
	})

	tasks.Start("listenershutdowner", func(ctx context.Context, _ string) error {
		<-ctx.Done()

		return socketListener.Close()
	})

	return tasks.Wait()
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
