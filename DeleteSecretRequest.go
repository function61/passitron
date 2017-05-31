package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

type DeleteSecretRequest struct {
	Id string
}

func (f *DeleteSecretRequest) Validate() error {
	if f.Id == "" {
		return errors.New("Id missing")
	}

	return nil
}

func HandleDeleteSecretRequest(w http.ResponseWriter, r *http.Request) {
	var req DeleteSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ApplyEvents([]interface{}{
		SecretDeleted{
			Id: req.Id,
		},
	})

	w.Write([]byte("OK"))
}
