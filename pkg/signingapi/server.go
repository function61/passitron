package signingapi

import (
	"bytes"
	"crypto/rand"
	"errors"
	"github.com/function61/gokit/httpauth"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/eventkit/event"
	"github.com/function61/pi-security-module/pkg/httpserver/muxregistrator"
	"github.com/function61/pi-security-module/pkg/httputil"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/ssh"
	"net/http"
	"time"
)

func lookupSignerByPubKey(pubKeyMarshaled []byte, userStorage *state.UserStorage) (ssh.Signer, *state.WrappedAccount, string, error) {
	for _, wacc := range userStorage.WrappedAccounts {
		for _, secret := range wacc.Secrets {
			if secret.SshPrivateKey == "" {
				continue
			}

			signer, err := ssh.ParsePrivateKey([]byte(secret.SshPrivateKey))
			if err != nil { // shouldn't happen
				return nil, nil, "", err
			}

			publicKey := signer.PublicKey()

			// apparently identities can only be compared by Marshal(), this is is done
			// the same way in SSH package
			if bytes.Equal(pubKeyMarshaled, publicKey.Marshal()) {
				return signer, &wacc, secret.Secret.Id, nil
			}
		}
	}

	return nil, nil, "", errors.New("privkey not found by pubkey")
}

type handlers struct {
	st *state.AppState
}

func (h *handlers) GetPublicKeys(rctx *httpauth.RequestContext, w http.ResponseWriter, r *http.Request) *PublicKeysOutput {
	keys := PublicKeysOutput{}

	for _, wacc := range h.st.DB.UserScope[rctx.User.Id].WrappedAccounts {
		for _, secret := range wacc.Secrets {
			if secret.SshPrivateKey == "" {
				continue
			}

			signer, err := ssh.ParsePrivateKey([]byte(secret.SshPrivateKey))
			if err != nil {
				httputil.RespondHttpJson(
					httputil.GenericError("private_key_parse_failed", err),
					http.StatusInternalServerError,
					w)
				return nil
			}

			publicKey := signer.PublicKey()

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

	signer, wacc, secretId, err := lookupSignerByPubKey(input.PublicKey, h.st.DB.UserScope[uid])
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
		event.Meta(time.Now(), uid))

	if err := h.st.EventLog.Append([]event.Event{secretUsedEvent}); err != nil {
		panic(err)
	}

	sig := Signature(*signature) // structs are type-compatible
	return &sig
}

func Setup(router *mux.Router, mwares httpauth.MiddlewareChainMap, st *state.AppState) {
	RegisterRoutes(&handlers{st}, mwares, muxregistrator.New(router))
}
