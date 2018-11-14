package restqueryapi

import (
	"bytes"
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
	apitypes.RegisterRoutes(&queryHandlers{
		st: st,
	}, func(path string, fn http.HandlerFunc) {
		router.HandleFunc(path, fn)
	})
}

type queryHandlers struct {
	st *state.State
}

func (a *queryHandlers) GetFolder(w http.ResponseWriter, r *http.Request) *apitypes.FolderResponse {
	if errorIfSealed(a.st.IsUnsealed(), w) {
		return nil
	}

	folder := a.st.FolderById(mux.Vars(r)["folderId"])

	accounts := state.UnwrapAccounts(a.st.WrappedAccountsByFolder(folder.Id))
	subFolders := a.st.SubfoldersByParentId(folder.Id)
	parentFolders := []apitypes.Folder{}

	parentId := folder.ParentId
	for parentId != "" {
		parent := a.st.FolderById(parentId)

		parentFolders = append(parentFolders, *parent)

		parentId = parent.ParentId
	}

	return &apitypes.FolderResponse{
		Folder:        folder,
		SubFolders:    subFolders,
		ParentFolders: parentFolders,
		Accounts:      accounts,
	}
}

func (a *queryHandlers) GetKeylistKey(w http.ResponseWriter, r *http.Request) *apitypes.SecretKeylistKey {
	key := mux.Vars(r)["key"]

	if errorIfSealed(a.st.IsUnsealed(), w) {
		return nil
	}

	if !runPhysicalAuth(w) {
		return nil // error handled internally
	}

	accountId := mux.Vars(r)["accountId"]

	wsecret := a.st.WrappedSecretById(accountId, mux.Vars(r)["secretId"])
	if wsecret == nil {
		httputil.RespondHttpJson(httputil.GenericError("keylist_key_not_found", nil), http.StatusNotFound, w)
		return nil
	}

	for _, keyEntry := range wsecret.KeylistKeys {
		if keyEntry.Key == key {
			a.st.EventLog.Append(domain.NewAccountSecretUsed(
				accountId,
				[]string{wsecret.Secret.Id},
				domain.SecretUsedTypeKeylistKeyExposed,
				keyEntry.Key,
				domain.Meta(time.Now(), domain.DefaultUserIdTODO)))

			return &keyEntry
		}
	}

	httputil.RespondHttpJson(httputil.GenericError("keylist_key_not_found", nil), http.StatusNotFound, w)
	return nil
}

func (a *queryHandlers) GetSecrets(input apitypes.GetSecretsInput, w http.ResponseWriter, r *http.Request) *[]apitypes.ExposedSecret {
	if errorIfSealed(a.st.IsUnsealed(), w) {
		return nil
	}

	wacc := a.st.WrappedAccountById(mux.Vars(r)["accountId"])

	if wacc == nil {
		httputil.RespondHttpJson(httputil.GenericError("account_not_found", nil), http.StatusNotFound, w)
		return nil
	}

	if err := exposeSecretsChallengeResponseOk(input.Challenge, input.SignResult, wacc.Account, a.st); err != nil {
		httputil.RespondHttpJson(httputil.GenericError("challenge_failed", err), http.StatusForbidden, w)
		return nil
	}

	secrets := state.UnwrapSecrets(wacc.Secrets)

	secretIdsForAudit := []string{}
	for _, secret := range secrets {
		secretIdsForAudit = append(secretIdsForAudit, secret.Secret.Id)
	}

	a.st.EventLog.Append(domain.NewAccountSecretUsed(
		wacc.Account.Id,
		secretIdsForAudit,
		domain.SecretUsedTypePasswordExposed,
		"",
		domain.Meta(time.Now(), domain.DefaultUserIdTODO)))

	return &secrets
}

func (a *queryHandlers) AuditLogEntries(w http.ResponseWriter, r *http.Request) *[]apitypes.AuditlogEntry {
	return &a.st.State.AuditLog
}

func (a *queryHandlers) GetAccount(w http.ResponseWriter, r *http.Request) *apitypes.WrappedAccount {
	if errorIfSealed(a.st.IsUnsealed(), w) {
		return nil
	}

	wacc := a.st.WrappedAccountById(mux.Vars(r)["id"])

	if wacc == nil {
		httputil.RespondHttpJson(httputil.GenericError("account_not_found", nil), http.StatusNotFound, w)
		return nil
	}

	u2fTokens := u2futil.GrabUsersU2FTokens(a.st)

	if len(u2fTokens) == 0 {
		http.Error(w, "no registered U2F tokens", http.StatusBadRequest)
		return nil
	}

	challenge, err := u2futil.NewU2FCustomChallenge(
		u2futil.GetAppIdHostname(),
		u2futil.GetTrustedFacets(),
		u2futil.ChallengeHashForAccountSecrets(wacc.Account))
	if err != nil {
		log.Printf("u2f.NewChallenge error: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return nil
	}

	signRequest := challenge.SignRequest(u2fTokens)

	return &apitypes.WrappedAccount{
		Challenge:   u2futil.ChallengeToApiType(*challenge),
		SignRequest: u2futil.SignRequestToApiType(*signRequest),
		Account:     wacc.Account,
	}
}

func (a *queryHandlers) Search(w http.ResponseWriter, r *http.Request) *apitypes.FolderResponse {
	if errorIfSealed(a.st.IsUnsealed(), w) {
		return nil
	}

	query := strings.ToLower(r.URL.Query().Get("q"))

	accounts := []apitypes.Account{}
	folders := []apitypes.Folder{}

	for _, folder := range a.st.State.Folders {
		if !strings.Contains(strings.ToLower(folder.Name), query) {
			continue
		}

		folders = append(folders, folder)
	}

	for _, wacc := range a.st.State.WrappedAccounts {
		if !strings.Contains(strings.ToLower(wacc.Account.Title), query) {
			continue
		}

		accounts = append(accounts, wacc.Account)
	}

	return &apitypes.FolderResponse{
		SubFolders: folders,
		Accounts:   accounts,
	}
}

func (a *queryHandlers) U2fEnrollmentChallenge(w http.ResponseWriter, r *http.Request) *apitypes.U2FEnrollmentChallenge {
	c, err := u2f.NewChallenge(u2futil.GetAppIdHostname(), u2futil.GetTrustedFacets())
	if err != nil {
		log.Printf("u2f.NewChallenge error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	req := u2f.NewWebRegisterRequest(c, u2futil.GrabUsersU2FTokens(a.st))

	registerRequests := []apitypes.U2FRegisterRequest{}
	for _, r := range req.RegisterRequests {
		registerRequests = append(registerRequests, apitypes.U2FRegisterRequest{
			Version:   r.Version,
			Challenge: r.Challenge,
		})
	}

	return &apitypes.U2FEnrollmentChallenge{
		Challenge: u2futil.ChallengeToApiType(*c),
		RegisterRequest: apitypes.U2FWebRegisterRequest{
			AppID:            req.AppID,
			RegisterRequests: registerRequests,
			RegisteredKeys:   u2futil.RegisteredKeysToApiType(req.RegisteredKeys),
		},
	}
}

func (a *queryHandlers) U2fEnrolledTokens(w http.ResponseWriter, r *http.Request) *[]apitypes.U2FEnrolledToken {
	if errorIfSealed(a.st.IsUnsealed(), w) {
		return nil
	}

	tokens := []apitypes.U2FEnrolledToken{}

	for _, token := range a.st.State.U2FTokens {
		tokens = append(tokens, apitypes.U2FEnrolledToken{
			Name:       token.Name,
			EnrolledAt: token.EnrolledAt,
			Version:    token.Version,
		})
	}

	return &tokens
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
