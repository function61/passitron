package main

import (
	"encoding/json"
	"net/http"
)

type WriteKeepassRequest struct {
}

func HandleWriteKeepassRequest(w http.ResponseWriter, r *http.Request) {
	var req WriteKeepassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	keepassExport("supersecret")

	w.Write([]byte("OK"))
}
