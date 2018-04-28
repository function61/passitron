package main

import (
	"encoding/json"
	"github.com/function61/pi-security-module/domain"
	"github.com/function61/pi-security-module/signingapi"
	"github.com/function61/pi-security-module/sshagent"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/extractpublicfiles"
	"github.com/function61/pi-security-module/util/keepassimport"
	"github.com/function61/pi-security-module/util/systemdinstaller"
	"github.com/function61/pi-security-module/util/version"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:generate go run gen/main.go gen/version.go gen/commands.go gen/events.go

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

// https://stackoverflow.com/a/2068407
func disableCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

func defineApi(router *mux.Router) {
	router.HandleFunc("/command/{commandName}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commandName := mux.Vars(r)["commandName"]

		// only command able to be invoked anonymously is the Unseal command
		commandNeedsAuthorization := commandName != "database.Unseal"

		if commandNeedsAuthorization && util.ErrorIfSealed(w, r, state.Inst.IsUnsealed()) {
			return
		}

		cmdBuilder, handlerExists := commandHandlers[commandName]
		if !handlerExists {
			util.CommandCustomError(w, r, "unsupported_command", nil, http.StatusBadRequest)
			return
		}

		ctx := &Ctx{
			State: state.Inst,
			Meta:  domain.Meta(time.Now(), "2"),
		}

		cmdStruct := cmdBuilder()

		// FIXME: assert application/json
		if errJson := json.NewDecoder(r.Body).Decode(cmdStruct); errJson != nil {
			util.CommandCustomError(w, r, "json_parsing_failed", errJson, http.StatusBadRequest)
			return
		}

		if errValidate := cmdStruct.Validate(); errValidate != nil {
			util.CommandCustomError(w, r, "command_validation_failed", errValidate, http.StatusBadRequest)
			return
		}

		if errInvoke := cmdStruct.Invoke(ctx); errInvoke != nil {
			util.CommandCustomError(w, r, "command_failed", errInvoke, http.StatusBadRequest)
			return
		}

		log.Printf("Command %s raised %d event(s)", commandName, len(ctx.raisedEvents))

		state.Inst.EventLog.AppendBatch(ctx.raisedEvents)

		type Output struct {
			Status string
		}

		disableCache(w)
		util.CommandGenericSuccess(w, r)
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

		disableCache(w)
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

		disableCache(w)
		w.Header().Set("Content-Type", "application/json")
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

		disableCache(w)
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

		state.Inst.EventLog.Append(domain.NewAccountSecretUsed(
			account.Id,
			"PasswordExposed",
			domain.Meta(time.Now(), "2")))

		disableCache(w)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(account.GetSecrets())
	}))
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
	} else if os.Args[1] != "run" {
		log.Fatalf("Invalid command: %v", os.Args[1])
	}

	if err := extractpublicfiles.Run(); err != nil {
		panic(err)
	}

	state.Initialize()
	defer state.Inst.Close()

	router := mux.NewRouter()

	defineApi(router)

	signingapi.Setup(router)

	// this most generic one has to be introduced last
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	log.Printf("Version %s listening in port 80", version.Version)

	log.Fatal(http.ListenAndServe(":80", router))
}
