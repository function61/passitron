package signingapi

import (
	"bytes"
	"crypto/rand"
	"errors"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/gokit/httpauth"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/httpserver/muxregistrator"
	"github.com/function61/pi-security-module/pkg/httputil"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/ssh"
	"net/http"
	"time"
)

func lookupSignerByPubKey(
	pubKeyMarshaled []byte,
	userStorage *state.UserStorage,
) (ssh.Signer, *state.InternalAccount, string, error) {
	for _, wacc := range userStorage.WrappedAccounts() {
		for _, secret := range wacc.Secrets {
			if secret.Kind != domain.SecretKindSshKey {
				continue
			}

			publicKey, err := parseSshPublicKeyFromAuthorizedFormat(secret.SshPublicKeyAuthorized)
			if err != nil { // shouldn't happen
				return nil, nil, "", err
			}

			// apparently identities can only be compared by Marshal(), this is is done
			// the same way in SSH package
			if !bytes.Equal(pubKeyMarshaled, publicKey.Marshal()) {
				continue
			}

			sshKeyDecrypted, err := userStorage.Crypto().Decrypt(secret.Envelope)
			if err != nil {
				return nil, nil, "", err
			}

			signer, err := ssh.ParsePrivateKey(sshKeyDecrypted)
			if err != nil { // shouldn't happen
				return nil, nil, "", err
			}

			return signer, &wacc, secret.Id, nil
		}
	}

	return nil, nil, "", errors.New("privkey not found by pubkey")
}

type handlers struct {
	st *state.AppState
}

func (h *handlers) GetPublicKeys(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *PublicKeysOutput {
	keys := PublicKeysOutput{}

	for _, wacc := range h.st.User(rctx.User.Id).WrappedAccounts() {
		for _, secret := range wacc.Secrets {
			if secret.Kind != domain.SecretKindSshKey {
				continue
			}

			publicKey, err := parseSshPublicKeyFromAuthorizedFormat(secret.SshPublicKeyAuthorized)
			if err != nil {
				httputil.RespondHttpJson(
					httputil.GenericError("parse_authorized_key", err),
					http.StatusInternalServerError,
					w)
				return nil
			}

			keys = append(keys, PublicKey{
				Format:  publicKey.Type(),
				Blob:    publicKey.Marshal(),
				Comment: wacc.Account.Title,
			})
		}
	}

	return &keys
}

func (h *handlers) Sign(rctx *httpauth.RequestContext, input SignRequestInput, w http.ResponseWriter, r *http.Request) *Signature {
	uid := rctx.User.Id

	signer, wacc, secretId, err := lookupSignerByPubKey(input.PublicKey, h.st.User(uid))
	if err != nil {
		httputil.RespondHttpJson(httputil.GenericError("privkey_for_pubkey_not_found", err), http.StatusBadRequest, w)
		return nil
	}

	signature, err := signer.Sign(rand.Reader, input.Data)
	if err != nil {
		httputil.RespondHttpJson(httputil.GenericError("signing_failed", err), http.StatusInternalServerError, w)
		return nil
	}

	secretUsedEvent := domain.NewAccountSecretUsed(
		wacc.Account.Id,
		[]string{secretId},
		domain.SecretUsedTypeSshSigning,
		"",
		ehevent.Meta(time.Now(), uid))

	if err := h.st.EventLog.Append([]ehevent.Event{secretUsedEvent}); err != nil {
		httputil.RespondHttpJson(httputil.GenericError("audit_event_saving_failed", err), http.StatusInternalServerError, w)
		return nil
	}

	return &Signature{
		Format: signature.Format,
		Blob:   signature.Blob,
	}
}

func Setup(router *mux.Router, mwares httpauth.MiddlewareChainMap, st *state.AppState) {
	RegisterRoutes(&handlers{st}, mwares, muxregistrator.New(router))
}

// parses from same format as in "authorized_keys" file
func parseSshPublicKeyFromAuthorizedFormat(pubKeyAuthorizedKeys string) (ssh.PublicKey, error) {
	// parses many, but we know this contains only one
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pubKeyAuthorizedKeys))
	return publicKey, err
}
