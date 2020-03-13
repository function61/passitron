package httpserver

import (
	"errors"
	"github.com/function61/gokit/httpauth"
	"github.com/function61/pi-security-module/pkg/httputil"
	"github.com/function61/pi-security-module/pkg/state"
	"net/http"
	"strings"
)

func createMiddlewares(appState *state.AppState) (httpauth.MiddlewareChainMap, error) {
	jwtAuth, err := httpauth.NewEcJwtAuthenticator(appState.GetJwtValidationKey())
	if err != nil {
		return nil, err
	}

	sealedCheck := func(w http.ResponseWriter) bool {
		if appState.IsUnsealed() {
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

	authCheck := func(w http.ResponseWriter, r *http.Request) *httpauth.UserDetails {
		authDetails, err := jwtAuth.AuthenticateWithCsrfProtection(r)
		if err != nil {
			httputil.RespondHttpJson(
				httputil.GenericError(
					"not_signed_in",
					err),
				http.StatusForbidden,
				w)

			return nil
		}

		return authDetails
	}

	resolveUidByAccessToken := func(r *http.Request) (string, bool) {
		bearerPrefix := "Bearer "
		authHeader := r.Header.Get("Authorization")

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			return "", false
		}

		token := authHeader[len(bearerPrefix):]
		if token == "" {
			return "", false
		}

		for _, userScope := range appState.DB.UserScope {
			if userScope.SensitiveUser().AccessToken == token {
				return userScope.SensitiveUser().User.Id, true
			}
		}

		return "", false
	}

	/*
		       public: no checks whatsoever
		authenticated: sealed check and auth check (itself contains CSRF check)
		       bearer: sealed check and bearer token check
	*/
	return httpauth.MiddlewareChainMap{
		"public": func(w http.ResponseWriter, r *http.Request) *httpauth.RequestContext {
			return &httpauth.RequestContext{}
		},
		"bearer": func(w http.ResponseWriter, r *http.Request) *httpauth.RequestContext {
			if !sealedCheck(w) {
				return nil
			}

			uid, ok := resolveUidByAccessToken(r)
			if !ok {
				httputil.RespondHttpJson(
					httputil.GenericError(
						"not_signed_in",
						errors.New("You must sign in before accessing this resource")),
					http.StatusForbidden,
					w)

				return nil
			}

			return &httpauth.RequestContext{
				User: &httpauth.UserDetails{
					Id: uid,
				},
			}
		},
		"authenticated": func(w http.ResponseWriter, r *http.Request) *httpauth.RequestContext {
			if !sealedCheck(w) {
				return nil
			}

			authDetails := authCheck(w, r)
			if authDetails == nil {
				return nil
			}

			return &httpauth.RequestContext{
				User: authDetails,
			}
		},
	}, nil
}
