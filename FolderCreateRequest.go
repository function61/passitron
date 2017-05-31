package main

import (
	"./util/cryptorandombytes" // FIXME
	"encoding/json"
	"errors"
	"net/http"
)

type FolderCreateRequest struct {
	ParentId string
	Name     string
}

func (f *FolderCreateRequest) Validate() error {
	if f.ParentId == "" {
		return errors.New("ParentId missing")
	}
	if f.Name == "" {
		return errors.New("Name missing")
	}

	return nil
}

func HandleFolderCreateRequest(w http.ResponseWriter, r *http.Request) {
	var req FolderCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ApplyEvents([]interface{}{
		FolderCreated{
			Id:       cryptorandombytes.Hex(4),
			ParentId: req.ParentId,
			Name:     req.Name,
		},
	})

	w.Write([]byte("OK"))
}
