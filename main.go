package main

import (
	"encoding/json"
	"github.com/function61/pi-security-module/accountevent"
	"github.com/function61/pi-security-module/keepassimport"
	"github.com/function61/pi-security-module/signingapi"
	"github.com/function61/pi-security-module/sshagent"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
	"github.com/function61/pi-security-module/util/eventlog"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:generate go run gen/main.go

func askAuthorization() (bool, error) {
	time.Sleep(2 * time.Second)

	return true, nil
}

type FolderResponse struct {
	Folder        *state.Folder
	SubFolders    []state.Folder
	ParentFolders []state.Folder
	Accounts      []state.SecureAccount
}

func defineApi(router *mux.Router) {
	router.HandleFunc("/command/{commandName}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commandName := mux.Vars(r)["commandName"]

		// only command able to be invoked unsealed is the Unseal command
		if commandName != "UnsealRequest" && util.ErrorIfSealed(w, r, state.Inst.IsUnsealed()) {
			return
		}

		// commandHandlers is generated
		handler, handlerExists := commandHandlers[commandName]
		if !handlerExists {
			util.CommandCustomError(w, r, "unsupported_command", nil, http.StatusBadRequest)
			return
		}

		handler(w, r)
	}))

	router.HandleFunc("/folder/{folderId}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if util.ErrorIfSealed(w, r, state.Inst.IsUnsealed()) {
			return
		}

		folder := state.FolderById(mux.Vars(r)["folderId"])

		accounts := state.AccountsByFolder(folder.Id)
		subFolders := state.SubfoldersById(folder.Id)
		parentFolders := []state.Folder{}

		parentId := folder.ParentId
		for parentId != "" {
			parent := state.FolderById(parentId)

			parentFolders = append(parentFolders, *parent)

			parentId = parent.ParentId
		}

		resp := FolderResponse{folder, subFolders, parentFolders, accounts}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))

	router.HandleFunc("/accounts", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if util.ErrorIfSealed(w, r, state.Inst.IsUnsealed()) {
			return
		}

		search := strings.ToLower(r.URL.Query().Get("search"))
		sshkey := strings.ToLower(r.URL.Query().Get("sshkey"))

		w.Header().Set("Content-Type", "application/json")

		matches := []state.SecureAccount{}

		if sshkey == "y" {
			for _, account := range state.Inst.State.Accounts {
				for _, secret := range account.Secrets {
					if secret.SshPublicKeyAuthorized == "" {
						continue
					}

					matches = append(matches, account.ToSecureAccount())
				}
			}
		} else if search == "" { // no filter => return all
			for _, s := range state.Inst.State.Accounts {
				matches = append(matches, s.ToSecureAccount())
			}
		} else { // search filter
			for _, s := range state.Inst.State.Accounts {
				if !strings.Contains(strings.ToLower(s.Title), search) {
					continue
				}

				matches = append(matches, s.ToSecureAccount())
			}
		}

		json.NewEncoder(w).Encode(matches)
	}))

	router.HandleFunc("/accounts/{accountId}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if util.ErrorIfSealed(w, r, state.Inst.IsUnsealed()) {
			return
		}

		account := state.AccountById(mux.Vars(r)["accountId"])

		if account == nil {
			util.CommandCustomError(w, r, "account_not_found", nil, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(account)
	}))

	router.HandleFunc("/accounts/{accountId}/secrets", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if util.ErrorIfSealed(w, r, state.Inst.IsUnsealed()) {
			return
		}

		account := state.AccountById(mux.Vars(r)["accountId"])

		if account == nil {
			util.CommandCustomError(w, r, "account_not_found", nil, http.StatusNotFound)
			return
		}

		authorized, err := askAuthorization()
		if err != nil {
			util.CommandCustomError(w, r, "technical_error_in_physical_authorization", err, http.StatusInternalServerError)
			return
		}

		if !authorized {
			util.CommandCustomError(w, r, "did_not_receive_physical_authorization", nil, http.StatusForbidden)
			return
		}

		state.Inst.EventLog.Append(accountevent.SecretUsed{
			Event:   eventbase.NewEvent(),
			Account: account.Id,
			Type:    accountevent.SecretUsedTypePasswordExposed,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(account.GetSecrets())
	}))
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <run>", os.Args[0])
	}

	if os.Args[1] == "keepassimport" {
		keepassimport.Run(os.Args[2:])
		return
	}

	if os.Args[1] == "agent" {
		sshagent.Run(os.Args[2:])
		return
	}

	if os.Args[1] != "run" {
		log.Fatalf("Invalid command", os.Args[1])
	}

	extractPublicFiles()

	state.Initialize()
	defer state.Inst.Close()

	if err := eventlog.ReadOldEvents(); err != nil {
		panic(err)
	}

	router := mux.NewRouter()

	defineApi(router)

	signingapi.Setup(router)

	// this most generic one has to be introduced last
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	log.Println("Starting in port 80")

	log.Fatal(http.ListenAndServe(":80", router))

}
