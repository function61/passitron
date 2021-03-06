package httpserver

import (
	"errors"
	"github.com/function61/gokit/httpauth"
	"github.com/function61/passitron/pkg/httputil"
	"github.com/function61/passitron/pkg/state"
	"net/http"
	"strings"
)

func createMiddlewares(appState *state.AppState) (httpauth.MiddlewareChainMap, error) {
	jwtAuth, err := httpauth.NewEcJwtAuthenticator(
		[]byte(appState.ValidatedJwtConf().AuthenticatorKey))
	if err != nil {
		return nil, err
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
		if token == "" { // important! user could have empty access token (if not set)
			return "", false
		}

		for _, userId := range appState.UserIds() {
			userScope := appState.User(userId)

			if userScope.SensitiveUser().AccessToken == token {
				return userScope.SensitiveUser().User.Id, true
			}
		}

		return "", false
	}

	/*
		       public: no checks whatsoever
		authenticated: auth check (itself contains CSRF check)
		       bearer: bearer token check
	*/
	return httpauth.MiddlewareChainMap{
		"public": func(w http.ResponseWriter, r *http.Request) *httpauth.RequestContext {
			return &httpauth.RequestContext{}
		},
		"bearer": func(w http.ResponseWriter, r *http.Request) *httpauth.RequestContext {
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
