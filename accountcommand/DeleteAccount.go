package accountcommand

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/accountevent"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
	"net/http"
)

type DeleteAccountRequest struct {
	Id string
}

func (f *DeleteAccountRequest) Validate() error {
	if f.Id == "" {
		return errors.New("Id missing")
	}

	return nil
}

func HandleDeleteAccountRequest(w http.ResponseWriter, r *http.Request) {
	var req DeleteAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		util.CommandValidationError(w, r, err)
		return
	}

	util.ApplyEvents([]interface{}{
		accountevent.AccountDeleted{
			Event: eventbase.NewEvent(),
			Id:    req.Id,
		},
	})

	util.CommandGenericSuccess(w, r)
}
