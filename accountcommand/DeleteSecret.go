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

type DeleteSecretRequest struct {
	Account string
	Secret  string
}

func HandleDeleteSecretRequest(w http.ResponseWriter, r *http.Request) {
	var req DeleteSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if state.AccountById(req.Account) == nil {
		util.CommandCustomError(w, r, "invalid_account_id", nil, http.StatusNotFound)
		return
	}

	eventapplicator.ApplyEvent(accountevent.SecretDeleted{
		Event:   eventbase.NewEvent(),
		Account: req.Account,
		Secret:  req.Secret,
	})

	util.CommandGenericSuccess(w, r)
}
