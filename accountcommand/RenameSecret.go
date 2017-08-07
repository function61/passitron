package accountcommand

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/accountevent"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
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
		util.CommandValidationError(w, r, err)
		return
	}

	util.ApplyEvents([]interface{}{
		accountevent.AccountRenamed{
			Event: eventbase.NewEvent(),
			Id:    req.Id,
			Title: req.Title,
		},
	})

	util.CommandGenericSuccess(w, r)
}
