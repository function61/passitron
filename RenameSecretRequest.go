package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

type RenameSecretRequest struct {
	Id    string
	Title string
}

func (f *RenameSecretRequest) Validate() error {
	if f.Id == "" {
		return errors.New("Id missing")
	}
	if f.Title == "" {
		return errors.New("Title missing")
	}

	return nil
}

func HandleRenameSecretRequest(w http.ResponseWriter, r *http.Request) {
	var req RenameSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ApplyEvents([]interface{}{
		SecretRenamed{
			Id:    req.Id,
			Title: req.Title,
		},
	})

	w.Write([]byte("OK"))
}
