package httputil

import (
	"encoding/json"
	"log"
	"net/http"
)

func RespondHttpJson(out interface{}, httpCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	// https://stackoverflow.com/a/2068407
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.WriteHeader(httpCode)

	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.Printf("respondHttpJson: failed to encode JSON: %s", err.Error())
	}
}
