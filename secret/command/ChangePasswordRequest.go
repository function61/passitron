package command

import (
	"encoding/json"
	"errors"
	"net/http"
	"github.com/function61/pi-security-module/secret/event"
	"github.com/function61/pi-security-module/util"
)

type ChangePasswordRequest struct {
	Id       string
	Password string
}

func (f *ChangePasswordRequest) Validate() error {
	if f.Id == "" {
		return errors.New("Id missing")
	}
	if f.Password == "" {
		return errors.New("Password missing")
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.ApplyEvents([]interface{}{
		event.PasswordChanged{
			Id:       req.Id,
			Password: req.Password,
		},
	})

	w.Write([]byte("OK"))
}
