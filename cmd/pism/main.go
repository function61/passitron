package main

import (
	"github.com/function61/pi-security-module/pkg/extractpublicfiles"
	"github.com/function61/pi-security-module/pkg/keepassimport"
	"github.com/function61/pi-security-module/pkg/restapi"
	"github.com/function61/pi-security-module/pkg/systemdinstaller"
	"github.com/function61/pi-security-module/pkg/version"
	"github.com/function61/pi-security-module/signingapi"
	"github.com/function61/pi-security-module/sshagent"
	"github.com/function61/pi-security-module/state"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

//go:generate go run gen/main.go gen/version.go gen/commands.go gen/events.go

func runMain() {
	if err := extractpublicfiles.Run(); err != nil {
		panic(err)
	}

	st := state.New()
	defer st.Close()

	router := mux.NewRouter()

	restapi.Define(router, st)

	signingapi.Setup(router, st)

	// this most generic one has to be introduced last
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	log.Printf("Version %s listening in port 80", version.Version)

	log.Fatal(http.ListenAndServe(":80", router))
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
	} else if os.Args[1] == "run" {
		runMain()
		return
	}

	log.Fatalf("Invalid command: %v", os.Args[1])
}
