package main

import (
	"fmt"
	"github.com/function61/pi-security-module/pkg/extractpublicfiles"
	"github.com/function61/pi-security-module/pkg/keepassimport"
	"github.com/function61/pi-security-module/pkg/osinterrupt"
	"github.com/function61/pi-security-module/pkg/restapi"
	"github.com/function61/pi-security-module/pkg/signingapi"
	"github.com/function61/pi-security-module/pkg/sshagent"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/function61/pi-security-module/pkg/systemdinstaller"
	"github.com/function61/pi-security-module/pkg/u2futil"
	"github.com/function61/pi-security-module/pkg/version"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	certFile = "cert.pem"
	keyFile  = "key.pem"
)

func runMain() {
	downloadUrl := extractpublicfiles.PublicFilesDownloadUrl(version.Version)
	if version.IsDevVersion() {
		downloadUrl = ""
	}

	if err := extractpublicfiles.Run(downloadUrl); err != nil {
		panic(err)
	}

	st := state.New()
	defer st.Close()

	router := mux.NewRouter()

	// FIXME: remove this crap bubblegum (uses global state)
	certBytes, errReadCertBytes := ioutil.ReadFile(certFile)
	if errReadCertBytes != nil {
		panic(errReadCertBytes)
	}

	if err := u2futil.InjectCommonNameFromSslCertificate(certBytes); err != nil {
		panic(err)
	}

	restapi.Define(router, st)

	signingapi.Setup(router, st)

	// this most generic one has to be introduced last
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	log.Printf("Version %s listening in port 443", version.Version)

	srv := &http.Server{
		Addr:    ":443",
		Handler: router,
	}

	httpStopped := make(chan bool)

	go func() {
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
			log.Printf("ListenAndServe() returned: %s", err.Error())
		}

		httpStopped <- true
	}()

	log.Printf("Received signal %s; shutting down", osinterrupt.WaitForIntOrTerm())

	if err := srv.Shutdown(nil); err != nil {
		log.Printf("Error shutting down HTTP server: %s", err.Error())
	}

	<-httpStopped

	log.Printf("Bye")
}

func serverEntrypoint() *cobra.Command {
	server := &cobra.Command{
		Use:   "server",
		Short: "Starts the server",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			runMain()
		},
	}

	server.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "Installs systemd unit file to make pi-security-module start on system boot",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			errInstall := systemdinstaller.InstallSystemdServiceFile(
				"pi-security-module",
				[]string{"server"},
				"Pi security module")

			if errInstall != nil {
				log.Fatalf("Installation failed: %s", errInstall)
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
