package command

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/folder/event"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/cryptorandombytes"
	"github.com/function61/pi-security-module/util/eventbase"
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
		util.CommandValidationError(w, r, err)
		return
	}

	util.ApplyEvents([]interface{}{
		event.FolderCreated{
			Event:    eventbase.NewEvent(),
			Id:       cryptorandombytes.Hex(4),
			ParentId: req.ParentId,
			Name:     req.Name,
		},
	})

	util.CommandGenericSuccess(w, r)
}
