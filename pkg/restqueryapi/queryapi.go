package restqueryapi

import (
	"bytes"
	"errors"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/function61/gokit/httpauth"
	"github.com/function61/gokit/mac"
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/eventkit/event"
	"github.com/function61/pi-security-module/pkg/httputil"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/function61/pi-security-module/pkg/u2futil"
	"github.com/gorilla/mux"
	"github.com/tstranex/u2f"
	"image/png"
	"log"
	"net/http"
	"strings"
	"time"
)

func Register(router *mux.Router, mwares httpauth.MiddlewareChainMap, st *state.State) {
	apitypes.RegisterRoutes(&queryHandlers{
		st: st,
	}, mwares, func(method string, path string, fn http.HandlerFunc) {
		router.HandleFunc(path, fn).Methods(method)
	})
}

type queryHandlers struct {
	st *state.State
}

func (a *queryHandlers) GetFolder(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *apitypes.FolderResponse {
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

func (a *queryHandlers) GetKeylistItem(rctx *httpauth.RequestContext, u2fResponse apitypes.U2FResponseBundle, w http.ResponseWriter, r *http.Request) *apitypes.SecretKeylistKey {
	accountId := mux.Vars(r)["accountId"]
	secretId := mux.Vars(r)["secretId"]
	key := mux.Vars(r)["key"]

	u2fChallengeHash := u2futil.ChallengeHashForKeylistKey(
		accountId,
		secretId,
		key)

	wsecret := a.st.WrappedSecretById(accountId, secretId)
	if wsecret == nil {
		httputil.RespondHttpJson(httputil.GenericError("keylist_key_not_found", nil), http.StatusNotFound, w)
		return nil
	}

	if err := u2fSignatureOk(rctx, u2fResponse, u2fChallengeHash, a.st); err != nil {
		httputil.RespondHttpJson(httputil.GenericError("u2f_challenge_response_failed", err), http.StatusForbidden, w)
		return nil
	}

	for _, keyEntry := range wsecret.KeylistKeys {
		if keyEntry.Key == key {
			secretUsedEvent := domain.NewAccountSecretUsed(
				accountId,
				[]string{wsecret.Secret.Id},
				domain.SecretUsedTypeKeylistKeyExposed,
				keyEntry.Key,
				event.Meta(time.Now(), rctx.User.Id))

			if err := a.st.EventLog.Append([]event.Event{secretUsedEvent}); err != nil {
				panic(err)
			}

			return &keyEntry
		}
	}

	httputil.RespondHttpJson(httputil.GenericError("keylist_key_not_found", nil), http.StatusNotFound, w)
	return nil
}

func (a *queryHandlers) GetKeylistItemChallenge(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *apitypes.U2FChallengeBundle {
	challengeHash := u2futil.ChallengeHashForKeylistKey(
		mux.Vars(r)["accountId"],
		mux.Vars(r)["secretId"],
		mux.Vars(r)["key"])

	u2fTokens := u2futil.GrabUsersU2FTokens(a.st)

	if len(u2fTokens) == 0 {
		http.Error(w, "no registered U2F tokens", http.StatusBadRequest)
		return nil
	}

	challenge, err := u2futil.NewU2FCustomChallenge(
		u2futil.GetAppIdHostname(),
		u2futil.GetTrustedFacets(),
		challengeHash)
	if err != nil {
		log.Printf("u2f.NewChallenge error: %v", err)
		http.Error(w, "error", http.StatusInternalServerError)
		return nil
	}

	return &apitypes.U2FChallengeBundle{
		Challenge:   u2futil.ChallengeToApiType(*challenge),
		SignRequest: u2futil.SignRequestToApiType(*challenge.SignRequest(u2fTokens)),
	}
}

func (a *queryHandlers) GetSecrets(rctx *httpauth.RequestContext, u2fResponse apitypes.U2FResponseBundle, w http.ResponseWriter, r *http.Request) *[]apitypes.ExposedSecret {
	wacc := a.st.WrappedAccountById(mux.Vars(r)["accountId"])

	if wacc == nil {
		httputil.RespondHttpJson(httputil.GenericError("account_not_found", nil), http.StatusNotFound, w)
		return nil
	}

	if err := u2fSignatureOk(rctx, u2fResponse, u2futil.ChallengeHashForAccountSecrets(wacc.Account), a.st); err != nil {
		httputil.RespondHttpJson(httputil.GenericError("u2f_challenge_response_failed", err), http.StatusForbidden, w)
		return nil
	}

	secrets := state.UnwrapSecrets(wacc.Secrets, a.st)

	secretIdsForAudit := []string{}
	for _, secret := range secrets {
		secretIdsForAudit = append(secretIdsForAudit, secret.Secret.Id)
	}

	secretUsedEvent := domain.NewAccountSecretUsed(
		wacc.Account.Id,
		secretIdsForAudit,
		domain.SecretUsedTypePasswordExposed,
		"",
		event.Meta(time.Now(), rctx.User.Id))

	if err := a.st.EventLog.Append([]event.Event{secretUsedEvent}); err != nil {
		panic(err)
	}

	return &secrets
}

func (a *queryHandlers) AuditLogEntries(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *[]apitypes.AuditlogEntry {
	return &a.st.State.AuditLog
}

func (a *queryHandlers) GetAccount(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *apitypes.WrappedAccount {
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
		ChallengeBundle: apitypes.U2FChallengeBundle{
			Challenge:   u2futil.ChallengeToApiType(*challenge),
			SignRequest: u2futil.SignRequestToApiType(*signRequest),
		},
		Account: wacc.Account,
	}
}

func (a *queryHandlers) Search(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *apitypes.FolderResponse {
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

func (a *queryHandlers) U2fEnrollmentChallenge(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *apitypes.U2FEnrollmentChallenge {
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

func (a *queryHandlers) U2fEnrolledTokens(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *[]apitypes.U2FEnrolledToken {
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

func (a *queryHandlers) TotpBarcodeExport(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) {
	accountId := mux.Vars(r)["accountId"]
	secretId := mux.Vars(r)["secretId"]

	account := a.st.WrappedAccountById(accountId)
	if account == nil {
		httputil.RespondHttpJson(httputil.GenericError("account_not_found", nil), http.StatusNotFound, w)
		return
	}

	var secret *state.WrappedSecret = nil
	for _, secretItem := range account.Secrets {
		if secretItem.Secret.Id == secretId {
			secret = &secretItem
			break
		}
	}

	if secret == nil {
		httputil.RespondHttpJson(httputil.GenericError("secret_not_found", nil), http.StatusNotFound, w)
		return
	}

	exportMac := mac.New(a.st.GetMacSigningKey(), secret.Secret.Id)

	if err := exportMac.Authenticate(r.URL.Query().Get("mac")); err != nil {
		httputil.RespondHttpJson(httputil.GenericError("invalid_mac", err), http.StatusForbidden, w)
		return
	}

	qrCode, err := qr.Encode(secret.OtpProvisioningUrl, qr.M, qr.Auto)
	if err != nil {
		httputil.RespondHttpJson(httputil.GenericError("qr_encode", err), http.StatusInternalServerError, w)
		return
	}

	qrCode, err = barcode.Scale(qrCode, 200, 200)
	if err != nil {
		httputil.RespondHttpJson(httputil.GenericError("barcode_scale", err), http.StatusInternalServerError, w)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	if err := png.Encode(w, qrCode); err != nil {
		httputil.RespondHttpJson(httputil.GenericError("png_encode", err), http.StatusInternalServerError, w)
		return
	}
}

func u2fSignatureOk(
	rctx *httpauth.RequestContext,
	response apitypes.U2FResponseBundle,
	expectedHash [32]byte,
	st *state.State,
) error {
	nativeChallenge := u2futil.ChallengeFromApiType(response.Challenge)

	if bytes.Compare(nativeChallenge.Challenge, expectedHash[:]) != 0 {
		return errors.New("invalid challenge hash")
	}

	u2ftoken := u2futil.GrabUsersU2FTokenByKeyHandle(st, response.SignResult.KeyHandle)
	if u2ftoken == nil {
		return errors.New("U2F token not found by KeyHandle")
	}

	reg := u2futil.U2ftokenToRegistration(u2ftoken)

	newCounter, authErr := reg.Authenticate(
		u2futil.SignResponseFromApiType(response.SignResult),
		nativeChallenge,
		u2ftoken.Counter)
	if authErr != nil {
		return authErr
	}

	u2fTokenUsedEvent := domain.NewUserU2FTokenUsed(
		response.SignResult.KeyHandle,
		int(newCounter),
		event.Meta(time.Now(), rctx.User.Id))

	return st.EventLog.Append([]event.Event{u2fTokenUsedEvent})
}
