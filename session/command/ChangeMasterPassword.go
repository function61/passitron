package command

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/session/event"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventapplicator"
	"github.com/function61/pi-security-module/util/eventbase"
	"net/http"
)

type ChangeMasterPassword struct {
	NewMasterPassword       string
	NewMasterPasswordRepeat string
}

func (f *ChangeMasterPassword) Validate() error {
	if f.NewMasterPassword == "" {
		return errors.New("NewMasterPassword missing")
	}
	if f.NewMasterPassword != f.NewMasterPasswordRepeat {
		return errors.New("NewMasterPassword not same as NewMasterPasswordRepeat")
	}

	return nil
}

func HandleChangeMasterPassword(w http.ResponseWriter, r *http.Request) {
	var req ChangeMasterPassword
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		util.CommandValidationError(w, r, err)
		return
	}

	if err := state.Inst.ChangePassword(req.NewMasterPassword); err != nil {
		util.CommandCustomError(w, r, "error", err, http.StatusInternalServerError)
		return
	}

	eventapplicator.ApplyEvent(event.MasterPasswordChanged{
		Event: eventbase.NewEvent(),
	})

	util.CommandGenericSuccess(w, r)
}
