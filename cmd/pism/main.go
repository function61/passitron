package main

import (
	"fmt"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/ossignal"
	"github.com/function61/gokit/stopper"
	"github.com/function61/gokit/systemdinstaller"
	"github.com/function61/pi-security-module/pkg/httpserver"
	"github.com/function61/pi-security-module/pkg/keepassimport"
	"github.com/function61/pi-security-module/pkg/signingapi"
	"github.com/function61/pi-security-module/pkg/sshagent"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/function61/pi-security-module/pkg/version"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func serverEntrypoint() *cobra.Command {
	server := &cobra.Command{
		Use:   "server",
		Short: "Starts the server",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			rootLogger := log.New(os.Stderr, "", log.LstdFlags)

			logl := logex.Levels(logex.Prefix("serverEntrypoint", rootLogger))

			logl.Info.Printf("%s starting", version.Version)
			defer logl.Info.Printf("Stopped")

			workers := stopper.NewManager()

			go func() {
				logl.Info.Printf("Received signal %s; stopping", <-ossignal.InterruptOrTerminate())

				workers.StopAllWorkersAndWait()
			}()

			if err := httpserver.Run(workers.Stopper(), logex.Prefix("httpserver", rootLogger)); err != nil {
				panic(err)
			}
		},
	}

	server.AddCommand(&cobra.Command{
		Use:   "init-config",
		Short: "Initializes configuration file",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := state.InitConfig(); err != nil {
				panic(err)
			}
		},
	})

	server.AddCommand(&cobra.Command{
		Use:   "print-signingapi-auth-token",
		Short: "Displays the auth token required to use the signing API",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			st := state.New()
			defer st.Close()

			fmt.Printf("%s\n", signingapi.ExpectedAuthHeader(st))
		},
	})

	server.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Installs systemd unit file to make pi-security-module start on system boot",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			hints, err := systemdinstaller.InstallSystemdServiceFile(
				"pi-security-module",
				[]string{"server"},
				"Pi security module")

			if err != nil {
				panic(err)
			} else {
				fmt.Println(hints)
			}
		},
	})

	return server
}

func main() {
	rootCmd := &cobra.Command{
		Use:     os.Args[0],
		Short:   "Software for a hardware security module",
		Version: version.Version,
	}

	rootCmd.AddCommand(serverEntrypoint())

	rootCmd.AddCommand(sshagent.Entrypoint())

	rootCmd.AddCommand(keepassimport.Entrypoint())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
