package main

import (
	"encoding/json"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
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

	util.CommandGenericSuccess(w, r)
}
