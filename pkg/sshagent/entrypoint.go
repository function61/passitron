package sshagent

import (
	"fmt"
	"github.com/function61/gokit/systemdinstaller"
	"github.com/spf13/cobra"
	"log"
)

func Entrypoint() *cobra.Command {
	sshAgent := &cobra.Command{
		Use:   "ssh-agent-proxy [baseurl] [token]",
		Short: "Starts the SSH agent proxy, which will forward SSH signing requests to pi-security-module",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			baseurl := args[0]
			token := args[1]

			Run(baseurl, token)
		},
	}

	sshAgent.AddCommand(&cobra.Command{
		Use:   "install [baseurl] [token]",
		Short: "Installs systemd unit file to make ssh-agent-proxy start on system boot",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			baseurl := args[0]
			token := args[1]

			hints, err := systemdinstaller.InstallSystemdServiceFile(
				"pi-security-module-ssh-agent",
				[]string{
					"ssh-agent-proxy",
					baseurl,
					token},
				"Pi security module SSH-agent")

			if err != nil {
				log.Fatalf("Installation failed: %s", err)
			} else {
				fmt.Println(hints)
			}
		},
	})

	return sshAgent
}
