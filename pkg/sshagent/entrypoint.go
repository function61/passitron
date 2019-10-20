package sshagent

import (
	"fmt"
	"github.com/function61/gokit/systemdinstaller"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func Entrypoint() *cobra.Command {
	sshAgent := &cobra.Command{
		Use:   "ssh-agent-proxy [baseurl] [token]",
		Short: "Starts the SSH agent proxy, which will forward SSH signing requests to pi-security-module",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			baseurl := args[0]
			token := args[1]

			rootLogger := log.New(os.Stderr, "", log.LstdFlags)

			if err := Run(baseurl, token, rootLogger); err != nil {
				panic(err)
			}
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
				"pi-security-module-ssh-agent",
				"Pi security module SSH-agent",
				systemdinstaller.Args("ssh-agent-proxy", baseurl, token))

			if err := systemdinstaller.Install(service); err != nil {
				log.Fatalf("Installation failed: %s", err)
			} else {
				fmt.Println(systemdinstaller.GetHints(service))
			}
		},
	})

	return sshAgent
}
