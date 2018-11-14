package main

import (
	"fmt"
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/ossignal"
	"github.com/function61/gokit/stopper"
	"github.com/function61/gokit/systemdinstaller"
	"github.com/function61/pi-security-module/pkg/extractpublicfiles"
	"github.com/function61/pi-security-module/pkg/keepassimport"
	"github.com/function61/pi-security-module/pkg/restcommandapi"
	"github.com/function61/pi-security-module/pkg/restqueryapi"
	"github.com/function61/pi-security-module/pkg/signingapi"
	"github.com/function61/pi-security-module/pkg/sshagent"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/function61/pi-security-module/pkg/u2futil"
	"github.com/function61/pi-security-module/pkg/version"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	certFile = "cert.pem"
	keyFile  = "key.pem"
)

func startHttp(st *state.State, stop *stopper.Stopper) error {
	log := logger.New("startHttp")

	router := mux.NewRouter()

	// FIXME: remove this crap bubblegum (uses global state)
	certBytes, errReadCertBytes := ioutil.ReadFile(certFile)
	if errReadCertBytes != nil {
		return errReadCertBytes
	}

	if err := u2futil.InjectCommonNameFromSslCertificate(certBytes); err != nil {
		return err
	}

	restqueryapi.Register(router, st)
	restcommandapi.Register(router, st)

	signingapi.Setup(router, st)

	// this most generic one has to be introduced last
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	srv := &http.Server{
		Addr:    ":443",
		Handler: router,
	}

	go func() {
		log.Info(fmt.Sprintf("serving @ %s", srv.Addr))
		defer log.Info("stopped")

		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
			log.Error(fmt.Sprintf("ListenAndServeTLS(): %s", err.Error()))
		}
	}()

	go func() {
		defer stop.Done()
		<-stop.Signal

		if err := srv.Shutdown(nil); err != nil {
			log.Error(fmt.Sprintf("Shutdown(): %s", err.Error()))
		}
	}()

	return nil
}

func server() error {
	log := logger.New("server")
	log.Info(fmt.Sprintf("%s starting", version.Version))
	defer log.Info("stopped")

	downloadUrl := extractpublicfiles.PublicFilesDownloadUrl(version.Version)
	if version.IsDevVersion() {
		downloadUrl = ""
	}

	if err := extractpublicfiles.Run(downloadUrl); err != nil {
		return err
	}

	st := state.New()
	defer st.Close()

	workers := stopper.NewManager()

	if err := startHttp(st, workers.Stopper()); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Received signal %s; stopping", <-ossignal.InterruptOrTerminate()))

	workers.StopAllWorkersAndWait()

	return nil
}

func serverEntrypoint() *cobra.Command {
	server := &cobra.Command{
		Use:   "server",
		Short: "Starts the server",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := server(); err != nil {
				panic(err)
			}
		},
	}

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
