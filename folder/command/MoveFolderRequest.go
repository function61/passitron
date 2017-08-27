package command

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/folder/event"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
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
	if state.FolderById(f.Id) == nil {
		return errors.New("Folder by Id not found")
	}
	if state.FolderById(f.ParentId) == nil {
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
		util.CommandValidationError(w, r, err)
		return
	}

	util.ApplyEvent(event.FolderMoved{
		Event:    eventbase.NewEvent(),
		Id:       req.Id,
		ParentId: req.ParentId,
	})

	util.CommandGenericSuccess(w, r)
}
