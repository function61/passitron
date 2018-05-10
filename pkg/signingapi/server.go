package signingapi

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/httputil"
	"github.com/function61/pi-security-module/pkg/signingapi/signingapitypes"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/ssh"
	"log"
	"net/http"
	"time"
)

func errorIfSealed(unsealed bool, w http.ResponseWriter) bool {
	if !unsealed {
		httputil.RespondHttpJson(httputil.GenericError("database_is_sealed", nil), http.StatusForbidden, w)
		return true
	}

	return false
}

func lookupSignerByPubKey(pubKeyMarshaled []byte, st *state.State) (ssh.Signer, *state.WrappedAccount, error) {
	for _, wacc := range st.State.WrappedAccounts {
		for _, secret := range wacc.Secrets {
			if secret.SshPrivateKey == "" {
				continue
			}

			signer, err := ssh.ParsePrivateKey([]byte(secret.SshPrivateKey))
			if err != nil {
				panic(err)
			}

			publicKey := signer.PublicKey()

			// TODO: is there better way to compare than marshal result?
			if bytes.Equal(pubKeyMarshaled, publicKey.Marshal()) {
				return signer, &wacc, nil
			}
		}
	}

	return nil, nil, errors.New("privkey not found by pubkey")
}

func expectedAuthHeader(st *state.State) string {
	bearerHash := sha256.Sum256([]byte(st.GetMasterPassword() + ":sshagent"))

	return "Bearer " + hex.EncodeToString(bearerHash[:])
}

func Setup(router *mux.Router, st *state.State) {
	log.Printf("signingapi expected auth: %s", expectedAuthHeader(st))

	router.HandleFunc("/_api/signer/publickeys", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfSealed(st.IsUnsealed(), w) {
			return
		}

		if r.Header.Get("Authorization") != expectedAuthHeader(st) {
			httputil.RespondHttpJson(httputil.GenericError("invalid_auth_header", errors.New("Authorization failed")), http.StatusForbidden, w)
			return
		}

		resp := signingapitypes.NewPublicKeysResponse()

		for _, wacc := range st.State.WrappedAccounts {
			for _, secret := range wacc.Secrets {
				if secret.SshPrivateKey == "" {
					continue
				}

				signer, err := ssh.ParsePrivateKey([]byte(secret.SshPrivateKey))
				if err != nil {
					panic(err)
				}

				publicKey := signer.PublicKey()

				pitem := signingapitypes.PublicKeyResponseItem{
					Format:  publicKey.Type(),
					Blob:    publicKey.Marshal(),
					Comment: wacc.Account.Title,
				}

				resp.PublicKeys = append(resp.PublicKeys, pitem)
			}
		}

		httputil.RespondHttpJson(resp, http.StatusOK, w)
	}))

	router.HandleFunc("/_api/signer/sign", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if errorIfSealed(st.IsUnsealed(), w) {
			return
		}

		if r.Header.Get("Authorization") != expectedAuthHeader(st) {
			httputil.RespondHttpJson(httputil.GenericError("invalid_auth_header", errors.New("Authorization failed")), http.StatusForbidden, w)
			return
		}

		var input signingapitypes.SignRequestInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			httputil.RespondHttpJson(httputil.GenericError("unable_to_parse_json", err), http.StatusBadRequest, w)
			return
		}

		signer, wacc, err := lookupSignerByPubKey(input.PublicKey, st)
		if err != nil {
			httputil.RespondHttpJson(httputil.GenericError("privkey_for_pubkey_not_found", err), http.StatusBadRequest, w)
			return
		}

		signature, err := signer.Sign(rand.Reader, input.Data)
		if err != nil {
			httputil.RespondHttpJson(httputil.GenericError("signing_failed", err), http.StatusInternalServerError, w)
			return
		}

		st.EventLog.Append(
			domain.NewAccountSecretUsed(
				wacc.Account.Id,
				domain.SecretUsedTypeSshSigning,
				domain.Meta(time.Now(), domain.DefaultUserIdTODO)))

		httputil.RespondHttpJson(signingapitypes.SignResponse{
			Signature: signature,
		}, http.StatusOK, w)
	}))
}
