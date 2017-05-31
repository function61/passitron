package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

type MoveFolderRequest struct {
	Id       string
	ParentId string
}

func (f *MoveFolderRequest) Validate() error {
	if f.Id == "" {
		return errors.New("Id missing")
	}
	if f.ParentId == "" {
		return errors.New("ParentId missing")
	}
	if folderById(f.Id) == nil {
		return errors.New("Folder by Id not found")
	}
	if folderById(f.ParentId) == nil {
		return errors.New("Folder by ParentId not found")
	}

	return nil
}

func HandleMoveFolderRequest(w http.ResponseWriter, r *http.Request) {
	var req MoveFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ApplyEvents([]interface{}{
		FolderMoved{
			Id:       req.Id,
			ParentId: req.ParentId,
		},
	})

	w.Write([]byte("OK"))
}
