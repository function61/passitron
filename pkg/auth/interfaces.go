package auth

import (
	"net/http"
)

type RequestContext struct {
	User *UserDetails
}

type UserDetails struct {
	Id string
}

type HttpRequestAuthenticator interface {
	Authenticate(*http.Request) *UserDetails
}

type Signer interface {
	Sign(userDetails UserDetails) string
}

// if returns nul, request handling is aborted.
// in return=nil case middleware is responsible for error response.
type MiddlewareChain func(w http.ResponseWriter, r *http.Request) *RequestContext

type MiddlewareChainMap map[string]MiddlewareChain
