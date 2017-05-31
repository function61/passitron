package main

import (
	"encoding/json"
	"errors"
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
	if secretById(f.Id) == nil {
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ApplyEvents([]interface{}{
		UsernameChanged{
			Id:       req.Id,
			Username: req.Username,
		},
	})

	w.Write([]byte("OK"))
}
