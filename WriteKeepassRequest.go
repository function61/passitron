package main

import (
	"encoding/json"
	"github.com/function61/pi-security-module/state"
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

	keepassExport(state.Inst.GetMasterPassword())

	w.Write([]byte("OK"))
}
