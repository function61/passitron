package main

import (
	"encoding/json"
	"github.com/function61/pi-security-module/state"
	"github.com/gorilla/mux"
	_ "github.com/wader/disable_sendfile_vbox_linux"
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

func defineApi(router *mux.Router) {
	router.HandleFunc("/command/{commandName}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commandName := mux.Vars(r)["commandName"]

		// commandHandlers is generated
		if handler, ok := commandHandlers[commandName]; ok {
			handler(w, r)
		} else {
			http.Error(w, "Unsupported command: "+commandName, http.StatusBadRequest)
		}
	}))

	router.HandleFunc("/folder/{folderId}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		search := strings.ToLower(r.URL.Query().Get("search"))

		w.Header().Set("Content-Type", "application/json")

		// no filter
		if search == "" {
			// w.WriteHeader(http.StatusOK)
			// w.Write([]byte("hello world"))
			json.NewEncoder(w).Encode(state.Data.Secrets)
		} else {
			matches := []state.Secret{}

			for _, s := range state.Data.Secrets {
				if !strings.Contains(strings.ToLower(s.Title), search) {
					continue
				}

				matches = append(matches, s.ToSecureSecret())
			}

			json.NewEncoder(w).Encode(matches)
		}
	}))

	router.HandleFunc("/secrets/{secretId}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := state.SecretById(mux.Vars(r)["secretId"])

		if secret == nil {
			http.Error(w, "Secret not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secret)
	}))

	router.HandleFunc("/secrets/{secretId}/expose", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorized, err := askAuthorization()
		if err != nil { // technical error in the authorization process
			panic(err)
		}

		if !authorized {
			http.Error(w, "Did not receive authorization", http.StatusForbidden)
			return
		}

		secret := state.SecretById(mux.Vars(r)["secretId"])

		if secret == nil {
			http.Error(w, "Secret not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secret.GetPassword())
	}))
}

func main() {
	state.Initialize()

	router := mux.NewRouter()

	defineApi(router)

	// this most generic one has to be introduced last
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	log.Println("Starting in port 80")

	log.Fatal(http.ListenAndServe(":80", router))
}
