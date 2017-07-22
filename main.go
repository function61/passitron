package main

import (
	"encoding/json"
	"github.com/function61/pi-security-module/secret/event"
	"github.com/function61/pi-security-module/sshagent"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
	"github.com/gorilla/mux"
	"log"
	"net/http"
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
	Secrets       []state.Secret
}

func errorIfUnsealed(w http.ResponseWriter, r *http.Request) bool {
	if !state.Inst.IsUnsealed() {
		util.CommandCustomError(w, r, "database_is_sealed", nil, http.StatusForbidden)
		return true
	}

	return false
}

func defineApi(router *mux.Router) {
	router.HandleFunc("/command/{commandName}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commandName := mux.Vars(r)["commandName"]

		// only command able to be invoked unsealed is the Unseal command
		if commandName != "UnsealRequest" && errorIfUnsealed(w, r) {
			return
		}

		// commandHandlers is generated
		if handler, ok := commandHandlers[commandName]; ok {
			handler(w, r)
		} else {
			util.CommandCustomError(w, r, "unsupported_command", nil, http.StatusBadRequest)
		}
	}))

	router.HandleFunc("/folder/{folderId}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfUnsealed(w, r) {
			return
		}

		folder := state.FolderById(mux.Vars(r)["folderId"])

		secrets := state.SecretsByFolder(folder.Id)
		subFolders := state.SubfoldersById(folder.Id)
		parentFolders := []state.Folder{}

		parentId := folder.ParentId
		for parentId != "" {
			parent := state.FolderById(parentId)

			parentFolders = append(parentFolders, *parent)

			parentId = parent.ParentId
		}

		resp := FolderResponse{folder, subFolders, parentFolders, secrets}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))

	router.HandleFunc("/secrets", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfUnsealed(w, r) {
			return
		}

		search := strings.ToLower(r.URL.Query().Get("search"))
		sshkey := strings.ToLower(r.URL.Query().Get("sshkey"))

		w.Header().Set("Content-Type", "application/json")

		matches := []state.Secret{}

		if sshkey == "y" {
			for _, s := range state.Inst.State.Secrets {
				if s.SshPublicKeyAuthorized == "" {
					continue
				}

				matches = append(matches, s.ToSecureSecret())
			}
		} else if search == "" { // no filter => return all
			for _, s := range state.Inst.State.Secrets {
				matches = append(matches, s.ToSecureSecret())
			}
		} else { // search filter
			for _, s := range state.Inst.State.Secrets {
				if !strings.Contains(strings.ToLower(s.Title), search) {
					continue
				}

				matches = append(matches, s.ToSecureSecret())
			}
		}

		json.NewEncoder(w).Encode(matches)
	}))

	router.HandleFunc("/secrets/{secretId}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfUnsealed(w, r) {
			return
		}

		secret := state.SecretById(mux.Vars(r)["secretId"])

		if secret == nil {
			util.CommandCustomError(w, r, "secret_not_found", nil, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secret)
	}))

	router.HandleFunc("/secrets/{secretId}/expose", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfUnsealed(w, r) {
			return
		}

		secret := state.SecretById(mux.Vars(r)["secretId"])

		if secret == nil {
			util.CommandCustomError(w, r, "secret_not_found", nil, http.StatusNotFound)
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

		util.ApplyEvents([]interface{}{
			event.SecretUsed{
				Event:  eventbase.NewEvent(),
				Secret: secret.Id,
				Type:   event.SecretUsedTypePasswordExposed,
			},
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secret.GetPassword())
	}))
}

func main() {
	extractPublicFiles()

	state.Initialize()

	go sshagent.Start()

	router := mux.NewRouter()

	defineApi(router)

	// this most generic one has to be introduced last
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	log.Println("Starting in port 80")

	log.Fatal(http.ListenAndServe(":80", router))
}
