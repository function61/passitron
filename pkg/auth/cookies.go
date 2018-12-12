package auth

import (
	"net/http"
)

const (
	loginCookieName = "login"
)

func ToCookie(tokenString string) *http.Cookie {
	return &http.Cookie{
		Name:     loginCookieName,
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true, // = not visible to JavaScript
		// Secure: true, // FIXME
	}
}

func DeleteLoginCookie() *http.Cookie {
	// NOTE: keep cookie attributes in sync with ToCookie(), since the cookies may be
	//       considered separate cookies, unless components like "Path" (might be more) match
	return &http.Cookie{
		Name:     loginCookieName,
		Value:    "del",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // => delete
		// Secure: true, // FIXME
	}
}
