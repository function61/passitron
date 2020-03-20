package sshagent

import (
	"fmt"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/ossignal"
	"github.com/function61/gokit/systemdinstaller"
	"github.com/spf13/cobra"
	"os"
)

func Entrypoint() *cobra.Command {
	sshAgent := &cobra.Command{
		Use:   "ssh-agent-proxy [baseurl] [token]",
		Short: "Starts the SSH agent proxy, which will forward SSH signing requests to Passitron",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			baseurl := args[0]
			token := args[1]

			rootLogger := logex.StandardLogger()

			exitIfError(Run(
				ossignal.InterruptOrTerminateBackgroundCtx(rootLogger),
				baseurl,
				token,
				rootLogger))
		},
	}

	sshAgent.AddCommand(&cobra.Command{
		Use:   "install [baseurl] [token]",
		Short: "Installs systemd unit file to make ssh-agent-proxy start on system boot",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			baseurl := args[0]
			token := args[1]

			service := systemdinstaller.SystemdServiceFile(
				"passitron-ssh-agent",
				"Passitron SSH-agent",
				systemdinstaller.Args("ssh-agent-proxy", baseurl, token))

			exitIfError(systemdinstaller.Install(service))

			fmt.Println(systemdinstaller.GetHints(service))
		},
	})

	return sshAgent
}

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
