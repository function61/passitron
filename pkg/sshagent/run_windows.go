// +build windows

package sshagent

import (
	"context"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/taskrunner"
	"github.com/microsoft/go-winio"
	"log"
)

const (
	// we're not OpenSSH but we need to use the name the SSH binary expects to connect to
	pipeName = `\\.\pipe\openssh-ssh-agent`
)

func run(ctx context.Context, agentServer *AgentServer, logger *log.Logger) error {
	logl := logex.Levels(logger)

	clientHandlerLogger := logex.Levels(logex.Prefix("handleOneClient", logger))

	listener, err := winio.ListenPipe(pipeName, nil)
	if err != nil {
		return err
	}

	tasks := taskrunner.New(ctx, logger)

	tasks.Start("listener "+pipeName, func(_ context.Context, _ string) error {
		logl.Info.Printf("listening at %s", pipeName)

		for {
			client, err := listener.Accept()
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

		return listener.Close()
	})

	return tasks.Wait()
}
