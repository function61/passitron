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

	ApplyEvents([]interface{}{
		SecretCreated{
			Id:       cryptorandombytes.Hex(4),
			FolderId: req.FolderId,
			Title:    req.Title,
			Username: req.Username,
			Password: req.Password,
		},
	})

	w.Write([]byte("OK"))
}
