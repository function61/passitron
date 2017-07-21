package command

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/secret/event"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
	"net/http"
)

type DeleteSecretRequest struct {
	Id string
}

func (f *DeleteSecretRequest) Validate() error {
	if f.Id == "" {
		return errors.New("Id missing")
	}

	return nil
}

func HandleDeleteSecretRequest(w http.ResponseWriter, r *http.Request) {
	var req DeleteSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.ApplyEvents([]interface{}{
		event.SecretDeleted{
			Event: eventbase.NewEvent(),
			Id:    req.Id,
		},
	})

	w.Write([]byte("OK"))
}
