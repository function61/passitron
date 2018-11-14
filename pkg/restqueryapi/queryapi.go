package restqueryapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/httputil"
	"github.com/function61/pi-security-module/pkg/physicalauth"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/function61/pi-security-module/pkg/u2futil"
	"github.com/gorilla/mux"
	"github.com/tstranex/u2f"
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

func runPhysicalAuth(w http.ResponseWriter) bool {
	authorized, err := physicalauth.Dummy()
	if err != nil {
		httputil.RespondHttpJson(httputil.GenericError("technical_error_in_physical_authorization", err), http.StatusInternalServerError, w)
		return false
	}

	if !authorized {
		httputil.RespondHttpJson(httputil.GenericError("did_not_receive_physical_authorization", nil), http.StatusForbidden, w)
		return false
	}

	return true
}

func Register(router *mux.Router, st *state.State) {
	router.HandleFunc("/u2f/enrollment/challenge", func(w http.ResponseWriter, r *http.Request) {
		c, err := u2f.NewChallenge(u2futil.GetAppIdHostname(), u2futil.GetTrustedFacets())
		if err != nil {
			log.Printf("u2f.NewChallenge error: %v", err)
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		req := u2f.NewWebRegisterRequest(c, u2futil.GrabUsersU2FTokens(st))

		registerRequests := []apitypes.U2FRegisterRequest{}
		for _, r := range req.RegisterRequests {
			registerRequests = append(registerRequests, apitypes.U2FRegisterRequest{
				Version:   r.Version,
				Challenge: r.Challenge,
			})
		}

		json.NewEncoder(w).Encode(apitypes.U2FEnrollmentChallenge{
			Challenge: u2futil.ChallengeToApiType(*c),
			RegisterRequest: apitypes.U2FWebRegisterRequest{
				AppID:            req.AppID,
				RegisterRequests: registerRequests,
				RegisteredKeys:   u2futil.RegisteredKeysToApiType(req.RegisteredKeys),
			},
		})
	})

	router.HandleFunc("/auditlog", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httputil.RespondHttpJson(st.State.AuditLog, http.StatusOK, w)
	}))

	router.HandleFunc("/folder/{folderId}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfSealed(st.IsUnsealed(), w) {
			return
		}

		folder := st.FolderById(mux.Vars(r)["folderId"])

		accounts := state.UnwrapAccounts(st.WrappedAccountsByFolder(folder.Id))
		subFolders := st.SubfoldersByParentId(folder.Id)
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

	router.HandleFunc("/search", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfSealed(st.IsUnsealed(), w) {
			return
		}

		query := strings.ToLower(r.URL.Query().Get("q"))

		accounts := []apitypes.Account{}
		folders := []apitypes.Folder{}

		for _, folder := range st.State.Folders {
			if !strings.Contains(strings.ToLower(folder.Name), query) {
				continue
			}

			folders = append(folders, folder)
		}

		for _, wacc := range st.State.WrappedAccounts {
			if !strings.Contains(strings.ToLower(wacc.Account.Title), query) {
				continue
			}

			accounts = append(accounts, wacc.Account)
		}

		httputil.RespondHttpJson(apitypes.FolderResponse{
			SubFolders: folders,
			Accounts:   accounts,
		}, http.StatusOK, w)
	}))

	router.HandleFunc("/u2f/enrolled_tokens", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfSealed(st.IsUnsealed(), w) {
			return
		}

		tokens := []apitypes.U2FEnrolledToken{}

		for _, token := range st.State.U2FTokens {
			tokens = append(tokens, apitypes.U2FEnrolledToken{
				Name:       token.Name,
				EnrolledAt: token.EnrolledAt,
				Version:    token.Version,
			})
		}

		httputil.RespondHttpJson(tokens, http.StatusOK, w)
	}))

	router.HandleFunc("/accounts", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfSealed(st.IsUnsealed(), w) {
			return
		}

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
		} else { // return all
			for _, wacc := range st.State.WrappedAccounts {
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

		u2fTokens := u2futil.GrabUsersU2FTokens(st)

		if len(u2fTokens) == 0 {
			http.Error(w, "no registered U2F tokens", http.StatusBadRequest)
			return
		}

		challenge, err := u2futil.NewU2FCustomChallenge(
			u2futil.GetAppIdHostname(),
			u2futil.GetTrustedFacets(),
			u2futil.ChallengeHashForAccountSecrets(wacc.Account))
		if err != nil {
			log.Printf("u2f.NewChallenge error: %v", err)
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		signRequest := challenge.SignRequest(u2fTokens)

		output := apitypes.WrappedAccount{
			Challenge:   u2futil.ChallengeToApiType(*challenge),
			SignRequest: u2futil.SignRequestToApiType(*signRequest),
			Account:     wacc.Account,
		}

		httputil.RespondHttpJson(output, http.StatusOK, w)
	}))

	router.HandleFunc("/accounts/{accountId}/secrets/{secretId}/keylist_keys/{key}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := mux.Vars(r)["key"]

		if errorIfSealed(st.IsUnsealed(), w) {
			return
		}

		if !runPhysicalAuth(w) {
			return // error handled internally
		}

		accountId := mux.Vars(r)["accountId"]

		wsecret := st.WrappedSecretById(accountId, mux.Vars(r)["secretId"])
		if wsecret == nil {
			httputil.RespondHttpJson(httputil.GenericError("keylist_key_not_found", nil), http.StatusNotFound, w)
			return
		}

		for _, keyEntry := range wsecret.KeylistKeys {
			if keyEntry.Key == key {
				st.EventLog.Append(domain.NewAccountSecretUsed(
					accountId,
					[]string{wsecret.Secret.Id},
					domain.SecretUsedTypeKeylistKeyExposed,
					keyEntry.Key,
					domain.Meta(time.Now(), domain.DefaultUserIdTODO)))

				httputil.RespondHttpJson(keyEntry, http.StatusOK, w)
				return
			}
		}

		httputil.RespondHttpJson(httputil.GenericError("keylist_key_not_found", nil), http.StatusNotFound, w)
	}))

	router.HandleFunc("/accounts/{accountId}/secrets", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfSealed(st.IsUnsealed(), w) {
			return
		}

		var input apitypes.GetSecretsInput
		if errJson := json.NewDecoder(r.Body).Decode(&input); errJson != nil {
			panic(errJson)
		}

		wacc := st.WrappedAccountById(mux.Vars(r)["accountId"])

		if wacc == nil {
			httputil.RespondHttpJson(httputil.GenericError("account_not_found", nil), http.StatusNotFound, w)
			return
		}

		if err := exposeSecretsChallengeResponseOk(input.Challenge, input.SignResult, wacc.Account, st); err != nil {
			httputil.RespondHttpJson(httputil.GenericError("challenge_failed", err), http.StatusForbidden, w)
			return
		}

		secrets := state.UnwrapSecrets(wacc.Secrets)

		secretIdsForAudit := []string{}
		for _, secret := range secrets {
			secretIdsForAudit = append(secretIdsForAudit, secret.Secret.Id)
		}

		st.EventLog.Append(domain.NewAccountSecretUsed(
			wacc.Account.Id,
			secretIdsForAudit,
			domain.SecretUsedTypePasswordExposed,
			"",
			domain.Meta(time.Now(), domain.DefaultUserIdTODO)))

		httputil.RespondHttpJson(secrets, http.StatusOK, w)
	}))
}

func exposeSecretsChallengeResponseOk(
	challenge apitypes.U2FChallenge,
	signResult apitypes.U2FSignResult,
	account apitypes.Account,
	st *state.State,
) error {
	expectedHash := u2futil.ChallengeHashForAccountSecrets(account)

	nativeChallenge := u2futil.ChallengeFromApiType(challenge)

	if bytes.Compare(nativeChallenge.Challenge, expectedHash[:]) != 0 {
		return errors.New("invalid challenge hash")
	}

	u2ftoken := u2futil.GrabUsersU2FTokenByKeyHandle(st, signResult.KeyHandle)
	if u2ftoken == nil {
		return errors.New("U2F token not found by KeyHandle")
	}

	reg := u2futil.U2ftokenToRegistration(u2ftoken)

	newCounter, authErr := reg.Authenticate(
		u2futil.SignResponseFromApiType(signResult),
		nativeChallenge,
		u2ftoken.Counter)
	if authErr != nil {
		return authErr
	}

	st.EventLog.Append(domain.NewUserU2FTokenUsed(
		signResult.KeyHandle,
		int(newCounter),
		domain.Meta(time.Now(), domain.DefaultUserIdTODO)))

	return nil
}
