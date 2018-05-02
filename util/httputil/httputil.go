package httputil

import (
	"net/http"
)

// https://stackoverflow.com/a/2068407
func DisableCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}
