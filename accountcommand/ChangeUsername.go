package accountcommand

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/accountevent"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
	"net/http"
)

type ChangeUsernameRequest struct {
	Id       string
	Username string
}

func (f *ChangeUsernameRequest) Validate() error {
	if f.Id == "" {
		return errors.New("Id missing")
	}
	if state.AccountById(f.Id) == nil {
		return errors.New("Secret by Id not found")
	}

	return nil
}

func HandleChangeUsernameRequest(w http.ResponseWriter, r *http.Request) {
	var req ChangeUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		util.CommandValidationError(w, r, err)
		return
	}

	state.Inst.EventLog.Append(accountevent.UsernameChanged{
		Event:    eventbase.NewEvent(),
		Id:       req.Id,
		Username: req.Username,
	})

	util.CommandGenericSuccess(w, r)
}
