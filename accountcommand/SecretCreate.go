package accountcommand

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/accountevent"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
	"net/http"
)

type SecretCreateRequest struct {
	FolderId string
	Title    string
	Username string
	Password string
	// TODO: repeat password, but optional
}

func (f *SecretCreateRequest) Validate() error {
	if f.FolderId == "" {
		return errors.New("FolderId missing")
	}
	if f.Title == "" {
		return errors.New("Title missing")
	}

	return nil
}

func HandleSecretCreateRequest(w http.ResponseWriter, r *http.Request) {
	var req SecretCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		util.CommandValidationError(w, r, err)
		return
	}

	accountId := eventbase.RandomId()

	events := []eventbase.EventInterface{
		accountevent.AccountCreated{
			Event:    eventbase.NewEvent(),
			Id:       accountId,
			FolderId: req.FolderId,
			Title:    req.Title,
		},
	}

	if req.Username != "" {
		events = append(events, accountevent.UsernameChanged{
			Event:    eventbase.NewEvent(),
			Id:       accountId,
			Username: req.Username,
		})
	}

	if req.Password != "" {
		events = append(events, accountevent.PasswordAdded{
			Event:    eventbase.NewEvent(),
			Account:  accountId,
			Id:       eventbase.RandomId(),
			Password: req.Password,
		})
	}

	state.Inst.EventLog.AppendBatch(events)

	util.CommandGenericSuccess(w, r)
}
