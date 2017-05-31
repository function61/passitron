package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

type RenameFolderRequest struct {
	Id   string
	Name string
}

func (f *RenameFolderRequest) Validate() error {
	if f.Id == "" {
		return errors.New("Id missing")
	}
	if f.Name == "" {
		return errors.New("Name missing")
	}

	return nil
}

func HandleRenameFolderRequest(w http.ResponseWriter, r *http.Request) {
	var req RenameFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ApplyEvents([]interface{}{
		FolderRenamed{
			Id:   req.Id,
			Name: req.Name,
		},
	})

	w.Write([]byte("OK"))
}
