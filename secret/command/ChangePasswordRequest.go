package command

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/secret/event"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
	"net/http"
)

type ChangePasswordRequest struct {
	Id             string
	Password       string
	PasswordRepeat string
}

func (f *ChangePasswordRequest) Validate() error {
	if f.Id == "" {
		return errors.New("Id missing")
	}
	if f.Password == "" {
		return errors.New("Password missing")
	}
	if f.Password != f.PasswordRepeat {
		return errors.New("PasswordRepeat different than Password")
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

	util.ApplyEvents([]interface{}{
		event.PasswordChanged{
			Event:    eventbase.NewEvent(),
			Id:       req.Id,
			Password: req.Password,
		},
	})

	util.CommandGenericSuccess(w, r)
}
