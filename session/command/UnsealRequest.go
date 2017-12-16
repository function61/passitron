package command

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/session/event"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
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
		util.CommandValidationError(w, r, err)
		return
	}

	if err := state.Inst.Unseal(req.MasterPassword); err != nil {
		util.CommandCustomError(w, r, "error", err, http.StatusForbidden)
		return
	}

	state.Inst.EventLog.Append(event.DatabaseUnsealed{
		Event: eventbase.NewEvent(),
	})

	util.CommandGenericSuccess(w, r)
}
