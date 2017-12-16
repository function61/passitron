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

	// TODO: predictable comparison time
	if state.Inst.GetMasterPassword() != f.MasterPassword {
		return errors.New("invalid password")
	}

	if state.Inst.IsUnsealed() {
		return errors.New("state already unsealed")
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

	state.Inst.SetSealed(false)

	state.Inst.EventLog.Append(event.DatabaseUnsealed{
		Event: eventbase.NewEvent(),
	})

	util.CommandGenericSuccess(w, r)
}
