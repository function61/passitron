package main

import (
	"fmt"
	"github.com/function61/gokit/dynversion"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/ossignal"
	"github.com/function61/gokit/systemdinstaller"
	"github.com/function61/passitron/pkg/httpserver"
	"github.com/function61/passitron/pkg/keepassimport"
	"github.com/function61/passitron/pkg/sshagent"
	"github.com/function61/passitron/pkg/state"
	"github.com/spf13/cobra"
	"os"
)

func serverEntrypoint() *cobra.Command {
	server := &cobra.Command{
		Use:   "server",
		Short: "Starts the server",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			rootLogger := logex.StandardLogger()

			exitIfError(httpserver.Run(
				ossignal.InterruptOrTerminateBackgroundCtx(rootLogger),
				rootLogger))
		},
	}

	server.AddCommand(&cobra.Command{
		Use:   "init-config [adminUsername] [adminPassword]",
		Short: "Initializes configuration file",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			exitIfError(state.InitConfig(args[0], args[1]))
		},
	})

	server.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Installs systemd unit file to make Passitron start on system boot",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			service := systemdinstaller.SystemdServiceFile(
				"passitron",
				"Passitron",
				systemdinstaller.Args("server"))

			exitIfError(systemdinstaller.Install(service))

			fmt.Println(systemdinstaller.GetHints(service))
		},
	})

	return server
}

func main() {
	rootCmd := &cobra.Command{
		Use:     os.Args[0],
		Short:   "Keeps your secrets as secure as possible",
		Version: dynversion.Version,
	}

	rootCmd.AddCommand(serverEntrypoint())

	rootCmd.AddCommand(sshagent.Entrypoint())

	rootCmd.AddCommand(keepassimport.Entrypoint())

	exitIfError(rootCmd.Execute())
}

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
