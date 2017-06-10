package command

import (
	"encoding/json"
	"github.com/function61/pi-security-module/state"
	"errors"
	"net/http"
)

type UnsealRequest struct {
	MasterPassword string
}

func (f *UnsealRequest) Validate() error {
	if f.MasterPassword == "" {
		return errors.New("MasterPassword missing")
	}

	return nil
}

func HandleUnsealRequest(w http.ResponseWriter, r *http.Request) {
	var req UnsealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := state.Inst.Unseal(req.MasterPassword); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Write([]byte("OK"))
}
