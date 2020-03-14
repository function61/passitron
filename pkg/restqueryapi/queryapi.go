package restqueryapi

import (
	"bytes"
	"errors"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/gokit/httpauth"
	"github.com/function61/gokit/mac"
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/httpserver/muxregistrator"
	"github.com/function61/pi-security-module/pkg/httputil"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/function61/pi-security-module/pkg/u2futil"
	"github.com/gorilla/mux"
	"github.com/tstranex/u2f"
	"image/png"
	"net/http"
	"time"
)

func Register(router *mux.Router, mwares httpauth.MiddlewareChainMap, st *state.AppState) {
	apitypes.RegisterRoutes(&queryHandlers{
		state: st,
	}, mwares, muxregistrator.New(router))
}

type queryHandlers struct {
	state *state.AppState
}

func (q *queryHandlers) userData(rctx *httpauth.RequestContext) *state.UserStorage {
	return q.state.User(rctx.User.Id)
}

func (a *queryHandlers) GetFolder(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *apitypes.FolderResponse {
	folder := a.userData(rctx).FolderById(mux.Vars(r)["folderId"])

	accounts := state.UnwrapAccounts(a.userData(rctx).WrappedAccountsByFolder(folder.Id))
	subFolders := a.userData(rctx).SubfoldersByParentId(folder.Id)
	parentFolders := []apitypes.Folder{}

	parentId := folder.ParentId
	for parentId != "" {
		parent := a.userData(rctx).FolderById(parentId)

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

func (q *queryHandlers) UserList(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *[]apitypes.User {
	users := []apitypes.User{}

	for _, userId := range q.state.UserIds() {
		userStorage := q.state.User(userId)

		users = append(users, userStorage.SensitiveUser().User)
	}

	return &users
}

func (a *queryHandlers) GetKeylistItem(rctx *httpauth.RequestContext, u2fResponse apitypes.U2FResponseBundle, w http.ResponseWriter, r *http.Request) *apitypes.SecretKeylistKey {
	accountId := mux.Vars(r)["accountId"]
	secretId := mux.Vars(r)["secretId"]
	key := mux.Vars(r)["key"]

	u2fChallengeHash := u2futil.ChallengeHashForKeylistKey(
		accountId,
		secretId,
		key)

	isecret := a.userData(rctx).InternalSecretById(accountId, secretId)
	if isecret == nil {
		httputil.RespondHttpJson(httputil.GenericError("keylist_key_not_found", nil), http.StatusNotFound, w)
		return nil
	}

	if err := u2fSignatureOk(rctx, u2fResponse, u2fChallengeHash, a.state); err != nil {
		httputil.RespondHttpJson(httputil.GenericError("u2f_challenge_response_failed", err), http.StatusForbidden, w)
		return nil
	}

	keys, err := a.userData(rctx).DecryptKeylist(*isecret)
	if err != nil {
		// TODO: ErrDecryptionKeyLocked
		httputil.RespondHttpJson(httputil.GenericError("keylist_decryption_failed", err), http.StatusForbidden, w)
		return nil
	}

	for _, keyEntry := range keys {
		if keyEntry.Key == key {
			secretUsedEvent := domain.NewAccountSecretUsed(
				accountId,
				[]string{isecret.Id},
				domain.SecretUsedTypeKeylistKeyExposed,
				keyEntry.Key,
				ehevent.Meta(time.Now(), rctx.User.Id))

			if err := a.state.EventLog.Append([]ehevent.Event{secretUsedEvent}); err != nil {
				httputil.RespondHttpJson(httputil.GenericError("audit_event_append_failed", err), http.StatusInternalServerError, w)
				return nil
			}

			return &apitypes.SecretKeylistKey{
				Key:   keyEntry.Key,
				Value: keyEntry.Value,
			}
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

	u2fTokens := u2futil.GrabUsersU2FTokens(a.state, rctx.User.Id)

	if len(u2fTokens) == 0 {
		http.Error(w, "no registered U2F tokens", http.StatusBadRequest)
		return nil
	}

	challenge, err := u2futil.NewU2FCustomChallenge(
		u2futil.GetAppIdHostname(),
		u2futil.GetTrustedFacets(),
		challengeHash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	return &apitypes.U2FChallengeBundle{
		Challenge:   u2futil.ChallengeToApiType(*challenge),
		SignRequest: u2futil.SignRequestToApiType(*challenge.SignRequest(u2fTokens)),
	}
}

func (a *queryHandlers) GetSecrets(rctx *httpauth.RequestContext, u2fResponse apitypes.U2FResponseBundle, w http.ResponseWriter, r *http.Request) *[]apitypes.ExposedSecret {
	userData := a.userData(rctx)

	acc := userData.WrappedAccountById(mux.Vars(r)["accountId"])

	if acc == nil {
		httputil.RespondHttpJson(httputil.GenericError("account_not_found", nil), http.StatusNotFound, w)
		return nil
	}

	if err := u2fSignatureOk(rctx, u2fResponse, u2futil.ChallengeHashForAccountSecrets(acc.Account), a.state); err != nil {
		httputil.RespondHttpJson(httputil.GenericError("u2f_challenge_response_failed", err), http.StatusForbidden, w)
		return nil
	}

	secrets, err := userData.DecryptSecrets(acc.Secrets, a.state)
	if err != nil {
		if err == state.ErrDecryptionKeyLocked {
			httputil.RespondHttpJson(
				httputil.GenericError(
					"database_is_sealed",
					nil),
				http.StatusForbidden,
				w)
			return nil
		} else {
			httputil.RespondHttpJson(httputil.GenericError("unwrap_secrets_failed", err), http.StatusForbidden, w)
			return nil
		}
	}

	secretIdsForAudit := []string{}
	for _, secret := range secrets {
		secretIdsForAudit = append(secretIdsForAudit, secret.Secret.Id)
	}

	secretUsedEvent := domain.NewAccountSecretUsed(
		acc.Account.Id,
		secretIdsForAudit,
		domain.SecretUsedTypePasswordExposed,
		"",
		ehevent.Meta(time.Now(), rctx.User.Id))

	if err := a.state.EventLog.Append([]ehevent.Event{secretUsedEvent}); err != nil {
		httputil.RespondHttpJson(httputil.GenericError("audit_event_append_failed", err), http.StatusInternalServerError, w)
		return nil
	}

	return &secrets
}

func (a *queryHandlers) AuditLogEntries(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *[]apitypes.AuditlogEntry {
	auditLog := a.userData(rctx).AuditLog()
	return &auditLog
}

func (a *queryHandlers) GetAccount(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *apitypes.WrappedAccount {
	acc := a.userData(rctx).WrappedAccountById(mux.Vars(r)["id"])

	if acc == nil {
		httputil.RespondHttpJson(httputil.GenericError("account_not_found", nil), http.StatusNotFound, w)
		return nil
	}

	u2fTokens := u2futil.GrabUsersU2FTokens(a.state, rctx.User.Id)

	if len(u2fTokens) == 0 {
		http.Error(w, "no registered U2F tokens", http.StatusBadRequest)
		return nil
	}

	challenge, err := u2futil.NewU2FCustomChallenge(
		u2futil.GetAppIdHostname(),
		u2futil.GetTrustedFacets(),
		u2futil.ChallengeHashForAccountSecrets(acc.Account))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	signRequest := challenge.SignRequest(u2fTokens)

	return &apitypes.WrappedAccount{
		ChallengeBundle: apitypes.U2FChallengeBundle{
			Challenge:   u2futil.ChallengeToApiType(*challenge),
			SignRequest: u2futil.SignRequestToApiType(*signRequest),
		},
		Account: acc.Account,
	}
}

func (a *queryHandlers) Search(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *apitypes.FolderResponse {
	query := r.URL.Query().Get("q")

	accounts := a.userData(rctx).SearchAccounts(query)
	folders := a.userData(rctx).SearchFolders(query)

	return &apitypes.FolderResponse{
		SubFolders: folders,
		Accounts:   accounts,
	}
}

func (a *queryHandlers) U2fEnrollmentChallenge(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *apitypes.U2FEnrollmentChallenge {
	c, err := u2f.NewChallenge(u2futil.GetAppIdHostname(), u2futil.GetTrustedFacets())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	req := u2f.NewWebRegisterRequest(c, u2futil.GrabUsersU2FTokens(a.state, rctx.User.Id))

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

	for _, token := range a.userData(rctx).U2FTokens() {
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

	userData := a.userData(rctx)

	secret := userData.InternalSecretById(accountId, secretId)
	if secret == nil {
		httputil.RespondHttpJson(httputil.GenericError("account_or_secret_not_found", nil), http.StatusNotFound, w)
		return
	}

	exportMac := mac.New(a.state.GetMacSigningKey(), secret.Id)

	if err := exportMac.Authenticate(r.URL.Query().Get("mac")); err != nil {
		httputil.RespondHttpJson(httputil.GenericError("invalid_mac", err), http.StatusForbidden, w)
		return
	}

	otpProvisioningUrl, err := userData.DecryptOtpProvisioningUrl(*secret)
	if err != nil {
		// TODO: ErrDecryptionKeyLocked
		httputil.RespondHttpJson(httputil.GenericError("decrypt_totp_provisioning_url", err), http.StatusForbidden, w)
		return
	}

	qrCode, err := qr.Encode(otpProvisioningUrl, qr.M, qr.Auto)
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
	st *state.AppState,
) error {
	nativeChallenge := u2futil.ChallengeFromApiType(response.Challenge)

	if !bytes.Equal(nativeChallenge.Challenge, expectedHash[:]) {
		return errors.New("invalid challenge hash")
	}

	u2ftoken := u2futil.GrabUsersU2FTokenByKeyHandle(st, rctx.User.Id, response.SignResult.KeyHandle)
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
		ehevent.Meta(time.Now(), rctx.User.Id))

	return st.EventLog.Append([]ehevent.Event{u2fTokenUsedEvent})
}
