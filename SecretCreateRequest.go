package main

import (
	"./util/cryptorandombytes" // FIXME
	"encoding/json"
	"errors"
	"net/http"
)

type SecretCreateRequest struct {
	FolderId string
	Title    string
	Username string
	Password string
}

func (f *SecretCreateRequest) Validate() error {
	if f.FolderId == "" {
		return errors.New("FolderId missing")
	}
	if f.Title == "" {
		return errors.New("Title missing")
	}

	return nil
}

func HandleSecretCreateRequest(w http.ResponseWriter, r *http.Request) {
	var req SecretCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	secretId := cryptorandombytes.Hex(4)

	events := []interface{}{
		SecretCreated{
			Id:       secretId,
			FolderId: req.FolderId,
			Title:    req.Title,
		},
	}

	if req.Username != "" {
		events = append(events, UsernameChanged{
			Id:       secretId,
			Username: req.Username,
		})
	}

	if req.Password != "" {
		events = append(events, PasswordChanged{
			Id:       secretId,
			Password: req.Password,
		})
	}

	ApplyEvents(events)

	w.Write([]byte("OK"))
}
