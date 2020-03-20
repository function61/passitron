package restqueryapi

import (
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/gokit/httpauth"
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
	if folder == nil {
		http.NotFound(w, r)
		return nil
	}

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

	userData := a.userData(rctx)

	u2fChallengeHash := u2futil.ChallengeHashForKeylistKey(
		accountId,
		secretId,
		key)

	isecret := userData.InternalSecretById(accountId, secretId)
	if isecret == nil {
		httputil.RespondHttpJson(httputil.GenericError("keylist_key_not_found", nil), http.StatusNotFound, w)
		return nil
	}

	u2fTokenUsedEvent, err := u2futil.SignatureOk(u2fResponse, u2fChallengeHash, userData)
	if err != nil {
		httputil.RespondHttpJson(httputil.GenericError("u2f_challenge_response_failed", err), http.StatusForbidden, w)
		return nil
	}
	if err := a.state.EventLog.Append([]ehevent.Event{u2fTokenUsedEvent}); err != nil {
		httputil.RespondHttpJson(httputil.GenericError("u2f_audit_failed", err), http.StatusInternalServerError, w)
		return nil
	}

	keys, err := userData.DecryptKeylist(*isecret)
	if err != nil {
		respondSecretDecryptionFailed(w, err)
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
	challengeBundle, err := u2futil.MakeChallengeBundle(u2futil.ChallengeHashForKeylistKey(
		mux.Vars(r)["accountId"],
		mux.Vars(r)["secretId"],
		mux.Vars(r)["key"]), a.userData(rctx))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	return challengeBundle
}

func (a *queryHandlers) GetSecrets(rctx *httpauth.RequestContext, u2fResponse apitypes.U2FResponseBundle, w http.ResponseWriter, r *http.Request) *[]apitypes.ExposedSecret {
	userData := a.userData(rctx)

	acc := userData.WrappedAccountById(mux.Vars(r)["accountId"])

	if acc == nil {
		httputil.RespondHttpJson(httputil.GenericError("account_not_found", nil), http.StatusNotFound, w)
		return nil
	}

	u2fTokenUsedEvent, err := u2futil.SignatureOk(u2fResponse, u2futil.ChallengeHashForAccountSecrets(acc.Account), userData)
	if err != nil {
		httputil.RespondHttpJson(httputil.GenericError("u2f_challenge_response_failed", err), http.StatusForbidden, w)
		return nil
	}
	if err := a.state.EventLog.Append([]ehevent.Event{u2fTokenUsedEvent}); err != nil {
		httputil.RespondHttpJson(httputil.GenericError("u2f_audit_failed", err), http.StatusInternalServerError, w)
		return nil
	}

	secrets, err := userData.DecryptSecrets(acc.Secrets)
	if err != nil {
		respondSecretDecryptionFailed(w, err)
		return nil
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
	userData := a.userData(rctx)

	acc := userData.WrappedAccountById(mux.Vars(r)["id"])

	if acc == nil {
		httputil.RespondHttpJson(httputil.GenericError("account_not_found", nil), http.StatusNotFound, w)
		return nil
	}

	challengeBundle, err := u2futil.MakeChallengeBundle(
		u2futil.ChallengeHashForAccountSecrets(acc.Account),
		userData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	return &apitypes.WrappedAccount{
		ChallengeBundle: *challengeBundle,
		Account:         acc.Account,
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
	c, err := u2f.NewChallenge(u2futil.GetAppIdHostname(), u2futil.MakeTrustedFacets())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	req := u2f.NewWebRegisterRequest(c, u2futil.GrabUsersU2FTokens(a.userData(rctx)))

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

func (a *queryHandlers) GetSignInChallenge(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *apitypes.U2FChallengeBundle {
	userData := a.state.User(mux.Vars(r)["userId"])

	// list of keyhandles for user is not exactly the most sensitive data, but still better
	// have a mac proving that user knew username/password combo before exposing this data
	mac := r.URL.Query().Get("mac")

	if err := userData.SignInGetU2fChallengeMac().Authenticate(mac); err != nil {
		httputil.RespondHttpJson(httputil.GenericError("invalid_mac", err), http.StatusForbidden, w)
		return nil
	}

	challengeBundle, err := u2futil.MakeChallengeBundle(u2futil.ChallengeHashForSignIn(userData.UserId()), userData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	return challengeBundle
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

	if err := userData.OtpKeyExportMac(secret).Authenticate(r.URL.Query().Get("mac")); err != nil {
		httputil.RespondHttpJson(httputil.GenericError("invalid_mac", err), http.StatusForbidden, w)
		return
	}

	// also validates secret kind
	otpProvisioningUrl, err := userData.DecryptOtpProvisioningUrl(*secret)
	if err != nil {
		respondSecretDecryptionFailed(w, err)
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

func respondSecretDecryptionFailed(w http.ResponseWriter, err error) {
	if err == state.ErrDecryptionKeyLocked {
		httputil.RespondHttpJson(
			httputil.GenericError(
				"database_is_sealed",
				nil),
			http.StatusForbidden,
			w)
	} else {
		httputil.RespondHttpJson(httputil.GenericError("secret_decryption_failed", err), http.StatusInternalServerError, w)
	}
}
