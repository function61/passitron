package restapi

import (
	"encoding/json"
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/command"
	"github.com/function61/pi-security-module/pkg/commandhandlers"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/httputil"
	"github.com/function61/pi-security-module/pkg/physicalauth"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
	"time"
)

func errorIfSealed(unsealed bool, w http.ResponseWriter) bool {
	if !unsealed {
		httputil.RespondHttpJson(httputil.GenericError("database_is_sealed", nil), http.StatusForbidden, w)
		return true
	}

	return false
}

func Define(router *mux.Router, st *state.State) {
	router.HandleFunc("/auditlog", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httputil.RespondHttpJson(st.State.AuditLog, http.StatusOK, w)
	}))

	router.HandleFunc("/command/{commandName}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commandName := mux.Vars(r)["commandName"]

		// only command able to be invoked anonymously is the Unseal command
		commandNeedsAuthorization := commandName != "database.Unseal"

		if commandNeedsAuthorization && errorIfSealed(st.IsUnsealed(), w) {
			return
		}

		cmdStructBuilder, commandExists := commandhandlers.StructBuilders[commandName]
		if !commandExists {
			httputil.RespondHttpJson(httputil.GenericError("unsupported_command", nil), http.StatusBadRequest, w)
			return
		}

		ctx := &command.Ctx{
			State: st,
			Meta:  domain.Meta(time.Now(), domain.DefaultUserIdTODO),
		}

		cmdStruct := cmdStructBuilder()

		// FIXME: assert application/json
		if errJson := json.NewDecoder(r.Body).Decode(cmdStruct); errJson != nil {
			httputil.RespondHttpJson(httputil.GenericError("json_parsing_failed", errJson), http.StatusBadRequest, w)
			return
		}

		if errValidate := cmdStruct.Validate(); errValidate != nil {
			httputil.RespondHttpJson(httputil.GenericError("command_validation_failed", errValidate), http.StatusBadRequest, w)
			return
		}

		if errInvoke := cmdStruct.Invoke(ctx); errInvoke != nil {
			httputil.RespondHttpJson(httputil.GenericError("command_failed", errInvoke), http.StatusBadRequest, w)
			return
		}

		raisedEvents := ctx.GetRaisedEvents()

		log.Printf("Command %s raised %d event(s)", commandName, len(raisedEvents))

		st.EventLog.AppendBatch(raisedEvents)

		httputil.RespondHttpJson(httputil.GenericSuccess(), http.StatusOK, w)
	}))

	router.HandleFunc("/folder/{folderId}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfSealed(st.IsUnsealed(), w) {
			return
		}

		folder := st.FolderById(mux.Vars(r)["folderId"])

		accounts := state.UnwrapAccounts(st.WrappedAccountsByFolder(folder.Id))
		subFolders := st.SubfoldersById(folder.Id)
		parentFolders := []apitypes.Folder{}

		parentId := folder.ParentId
		for parentId != "" {
			parent := st.FolderById(parentId)

			parentFolders = append(parentFolders, *parent)

			parentId = parent.ParentId
		}

		httputil.RespondHttpJson(apitypes.FolderResponse{
			Folder:        folder,
			SubFolders:    subFolders,
			ParentFolders: parentFolders,
			Accounts:      accounts,
		}, http.StatusOK, w)
	}))

	router.HandleFunc("/accounts", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfSealed(st.IsUnsealed(), w) {
			return
		}

		search := strings.ToLower(r.URL.Query().Get("search"))
		sshkey := strings.ToLower(r.URL.Query().Get("sshkey"))

		matches := []apitypes.Account{}

		if sshkey == "y" {
			for _, wacc := range st.State.WrappedAccounts {
				for _, secret := range wacc.Secrets {
					if secret.Secret.SshPublicKeyAuthorized == "" {
						continue
					}

					matches = append(matches, wacc.Account)
				}
			}
		} else if search == "" { // no filter => return all
			for _, wacc := range st.State.WrappedAccounts {
				matches = append(matches, wacc.Account)
			}
		} else { // search filter
			for _, wacc := range st.State.WrappedAccounts {
				if !strings.Contains(strings.ToLower(wacc.Account.Title), search) {
					continue
				}

				matches = append(matches, wacc.Account)
			}
		}

		httputil.RespondHttpJson(matches, http.StatusOK, w)
	}))

	router.HandleFunc("/accounts/{accountId}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfSealed(st.IsUnsealed(), w) {
			return
		}

		wacc := st.WrappedAccountById(mux.Vars(r)["accountId"])

		if wacc == nil {
			httputil.RespondHttpJson(httputil.GenericError("account_not_found", nil), http.StatusNotFound, w)
			return
		}

		httputil.RespondHttpJson(wacc.Account, http.StatusOK, w)
	}))

	router.HandleFunc("/accounts/{accountId}/secrets", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfSealed(st.IsUnsealed(), w) {
			return
		}

		wacc := st.WrappedAccountById(mux.Vars(r)["accountId"])

		if wacc == nil {
			httputil.RespondHttpJson(httputil.GenericError("account_not_found", nil), http.StatusNotFound, w)
			return
		}

		authorized, err := physicalauth.Dummy()
		if err != nil {
			httputil.RespondHttpJson(httputil.GenericError("technical_error_in_physical_authorization", err), http.StatusInternalServerError, w)
			return
		}

		if !authorized {
			httputil.RespondHttpJson(httputil.GenericError("did_not_receive_physical_authorization", nil), http.StatusForbidden, w)
			return
		}

		st.EventLog.Append(domain.NewAccountSecretUsed(
			wacc.Account.Id,
			domain.SecretUsedTypePasswordExposed,
			domain.Meta(time.Now(), domain.DefaultUserIdTODO)))

		httputil.RespondHttpJson(state.UnwrapSecrets(wacc.Secrets), http.StatusOK, w)
	}))
}
