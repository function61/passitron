package accountcommand

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/accountevent"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
	"github.com/function61/pi-security-module/util/randompassword"
	"net/http"
)

type ChangePasswordRequest struct {
	Id             string
	Password       string
	PasswordRepeat string
}

func (f *ChangePasswordRequest) Validate() error {
	// FIXME: validate account presence

	if f.Id == "" {
		return errors.New("Id missing")
	}
	if f.Password == "" {
		return errors.New("Password missing")
	}
	if f.Password != f.PasswordRepeat {
		return errors.New("PasswordRepeat different than Password")
	}

	if f.Password == "_auto" {
		f.Password = randompassword.Build(randompassword.DefaultAlphabet, 16)
	}

	return nil
}

func HandleChangePasswordRequest(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		util.CommandValidationError(w, r, err)
		return
	}

	state.Inst.EventLog.Append(accountevent.PasswordAdded{
		Event:    eventbase.NewEvent(),
		Account:  req.Id,
		Id:       eventbase.RandomId(),
		Password: req.Password,
	})

	util.CommandGenericSuccess(w, r)
}
