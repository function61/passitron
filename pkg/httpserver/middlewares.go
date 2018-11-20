package httpserver

import (
	"errors"
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/auth"
	"github.com/function61/pi-security-module/pkg/httputil"
	"github.com/function61/pi-security-module/pkg/state"
	"net/http"
)

func createMiddlewares(st *state.State) (apitypes.MiddlewareChainMap, error) {
	jwtAuth, err := auth.NewJwtAuthenticator(st.GetJwtValidationKey())
	if err != nil {
		return nil, err
	}

	sealedCheck := func(w http.ResponseWriter) bool {
		if st.IsUnsealed() {
			return true
		}

		httputil.RespondHttpJson(
			httputil.GenericError(
				"database_is_sealed",
				nil),
			http.StatusForbidden,
			w)

		return false
	}

	csrfCheck := func(w http.ResponseWriter, r *http.Request) bool {
		if r.Header.Get("x-csrf-token") == st.GetCsrfToken() {
			return true
		}

		httputil.RespondHttpJson(
			httputil.GenericError(
				"invalid_csrf_token",
				errors.New("CSRF token is invalid or missing. Do you happen to be wearing a hoodie?")),
			http.StatusForbidden,
			w)

		return false
	}

	authCheck := func(w http.ResponseWriter, r *http.Request) *auth.UserDetails {
		authDetails := jwtAuth.Authenticate(r)
		if authDetails != nil {
			return authDetails
		}

		httputil.RespondHttpJson(
			httputil.GenericError(
				"not_signed_in",
				errors.New("You must sign in before accessing this resource")),
			http.StatusForbidden,
			w)

		return nil
	}

	/*
		public: no checks whatsoever
		authdWrite: sealed check, CSRF check and auth check
		authdQuery: same as authdWrite but no CSRF check
	*/
	return apitypes.MiddlewareChainMap{
		"public": func(w http.ResponseWriter, r *http.Request) *apitypes.RequestContext {
			return &apitypes.RequestContext{}
		},
		"authdQuery": func(w http.ResponseWriter, r *http.Request) *apitypes.RequestContext {
			if !sealedCheck(w) {
				return nil
			}

			authDetails := authCheck(w, r)
			if authDetails == nil {
				return nil
			}

			return &apitypes.RequestContext{
				User: authDetails,
			}
		},
		"authdWrite": func(w http.ResponseWriter, r *http.Request) *apitypes.RequestContext {
			if !sealedCheck(w) {
				return nil
			}
			if !csrfCheck(w, r) {
				return nil
			}

			authDetails := authCheck(w, r)
			if authDetails == nil {
				return nil
			}

			return &apitypes.RequestContext{
				User: authDetails,
			}
		},
	}, nil
}
