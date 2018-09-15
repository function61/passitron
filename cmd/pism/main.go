package main

import (
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

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <run>", os.Args[0])
	} else if os.Args[1] == "keepassimport" {
		keepassimport.Run(os.Args[2:])
		return
	} else if os.Args[1] == "agent" {
		sshagent.Run(os.Args[2:])
		return
	} else if os.Args[1] == "install" {
		errInstall := systemdinstaller.InstallSystemdServiceFile(
			"pi-security-module",
			[]string{"run"},
			"Pi security module")

		if errInstall != nil {
			log.Fatalf("Installation failed: %s", errInstall)
		}
		return
	} else if os.Args[1] == "install-ssh-agent" && len(os.Args) == 4 {
		baseurl := os.Args[2]
		token := os.Args[3]

		errInstall := systemdinstaller.InstallSystemdServiceFile(
			"pi-security-module-ssh-agent",
			[]string{
				"agent",
				baseurl,
				token},
			"Pi security module SSH-agent")

		if errInstall != nil {
			log.Fatalf("Installation failed: %s", errInstall)
		}
		return
	} else if os.Args[1] == "run" {
		runMain()
		return
	}

	log.Fatalf("Invalid command: %v", os.Args[1])
}
