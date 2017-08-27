package accountcommand

import (
	"encoding/json"
	"github.com/function61/pi-security-module/accountevent"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventapplicator"
	"github.com/function61/pi-security-module/util/eventbase"
	"net/http"
)

type ChangeDescriptionRequest struct {
	Id          string
	Description string
}

func HandleChangeDescriptionRequest(w http.ResponseWriter, r *http.Request) {
	var req ChangeDescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if state.AccountById(req.Id) == nil {
		util.CommandCustomError(w, r, "invalid_secret_id", nil, http.StatusNotFound)
		return
	}

	eventapplicator.ApplyEvent(accountevent.DescriptionChanged{
		Event:       eventbase.NewEvent(),
		Id:          req.Id,
		Description: req.Description,
	})

	util.CommandGenericSuccess(w, r)
}
