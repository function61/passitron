package command

import (
	"encoding/json"
	"github.com/function61/pi-security-module/secret/event"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
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

	if state.SecretById(req.Id) == nil {
		util.CommandCustomError(w, r, "invalid_secret_id", nil, http.StatusNotFound)
		return
	}

	util.ApplyEvents([]interface{}{
		event.DescriptionChanged{
			Event:       eventbase.NewEvent(),
			Id:          req.Id,
			Description: req.Description,
		},
	})

	util.CommandGenericSuccess(w, r)
}
