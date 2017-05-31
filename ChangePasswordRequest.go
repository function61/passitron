package main

import (
	"encoding/json"
	"errors"
	"net/http"
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

	ApplyEvents([]interface{}{
		PasswordChanged{
			Id:       req.Id,
			Password: req.Password,
		},
	})

	w.Write([]byte("OK"))
}
