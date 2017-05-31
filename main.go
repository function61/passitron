package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	_ "github.com/wader/disable_sendfile_vbox_linux"
	"log"
	"net/http"
	"strings"
	"time"
)

//go:generate go run gen/main.go

var state *Statefile

func secretById(id string) *Secret {
	for _, s := range state.Secrets {
		if s.Id == id {
			secret := s.ToSecureSecret()
			return &secret
		}
	}

	return nil
}

func folderById(id string) *Folder {
	for _, f := range state.Folders {
		if f.Id == id {
			return &f
		}
	}

	return nil
}

func askAuthorization() (bool, error) {
	time.Sleep(2 * time.Second)

	return true, nil
}

func oneSecret(w http.ResponseWriter, r *http.Request) {
	secret := secretById(mux.Vars(r)["secretId"])

	if secret == nil {
		http.Error(w, "Secret not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secret)
}

func expose(w http.ResponseWriter, r *http.Request) {
	authorized, err := askAuthorization()
	if err != nil { // technical error in the authorization process
		panic(err)
	}

	if !authorized {
		http.Error(w, "Did not receive authorization", http.StatusForbidden)
		return
	}

	secret := secretById(mux.Vars(r)["secretId"])

	if secret == nil {
		http.Error(w, "Secret not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secret.GetPassword())
}

func getSecrets(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	w.Header().Set("Content-Type", "application/json")

	// no filter
	if search == "" {
		// w.WriteHeader(http.StatusOK)
		// w.Write([]byte("hello world"))
		json.NewEncoder(w).Encode(state.Secrets)
	} else {
		matches := []Secret{}

		for _, s := range state.Secrets {
			if !strings.Contains(s.Title, search) {
				continue
			}

			matches = append(matches, s.ToSecureSecret())
		}

		json.NewEncoder(w).Encode(matches)
	}
}

type FolderResponse struct {
	Folder        *Folder
	SubFolders    []Folder
	ParentFolders []Folder
	Secrets       []Secret
}

func subfoldersById(id string) []Folder {
	subFolders := []Folder{}

	for _, f := range state.Folders {
		if f.ParentId != id {
			continue
		}

		subFolders = append(subFolders, f)
	}

	return subFolders
}

func secretsByFolder(id string) []Secret {
	secrets := []Secret{}

	for _, s := range state.Secrets {
		if s.FolderId != id {
			continue
		}

		secrets = append(secrets, s.ToSecureSecret())
	}

	return secrets
}

func restFolder(w http.ResponseWriter, r *http.Request) {
	folder := folderById(mux.Vars(r)["folderId"])

	secrets := secretsByFolder(folder.Id)
	subFolders := subfoldersById(folder.Id)
	parentFolders := []Folder{}

	parentId := folder.ParentId
	for parentId != "" {
		parent := folderById(parentId)

		parentFolders = append(parentFolders, *parent)

		parentId = parent.ParentId
	}

	resp := FolderResponse{folder, subFolders, parentFolders, secrets}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	state, _ = ReadStatefile()

	router := mux.NewRouter()

	router.HandleFunc("/command/{commandName}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commandName := mux.Vars(r)["commandName"]

		// commandHandlers is generated
		if handler, ok := commandHandlers[commandName]; ok {
			handler(w, r)
		} else {
			http.Error(w, "Unsupported command: "+commandName, http.StatusBadRequest)
		}
	}))

	router.HandleFunc("/folder/{folderId}", http.HandlerFunc(restFolder))
	router.HandleFunc("/secrets", http.HandlerFunc(getSecrets))
	router.HandleFunc("/secrets/{secretId}", http.HandlerFunc(oneSecret))
	router.HandleFunc("/secrets/{secretId}/expose", http.HandlerFunc(expose))

	// this most generic one has to be introduced last
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	log.Println("Starting in port 80")

	log.Fatal(http.ListenAndServe(":80", router))
}

func ApplyEvents(events []interface{}) {
	for _, e := range events {
		if !ApplyOneEvent(e) {
			panic("Unknown event")
		}
	}
}
