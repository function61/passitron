package accountcommand

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/accountevent"
	"github.com/function61/pi-security-module/util"
	"github.com/function61/pi-security-module/util/eventbase"
	"net/http"
)

type SetOtpTokenRequest struct {
	Id                 string
	OtpProvisioningUrl string
}

func (f *SetOtpTokenRequest) Validate() error {
	if f.Id == "" {
		return errors.New("Id missing")
	}
	if f.OtpProvisioningUrl == "" {
		return errors.New("OtpProvisioningUrl missing")
	}

	return nil
}

func HandleSetOtpTokenRequest(w http.ResponseWriter, r *http.Request) {
	var req SetOtpTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		util.CommandValidationError(w, r, err)
		return
	}

	util.ApplyEvent(accountevent.OtpTokenAdded{
		Event:              eventbase.NewEvent(),
		Account:            req.Id,
		Id:                 eventbase.RandomId(),
		OtpProvisioningUrl: req.OtpProvisioningUrl,
	})

	util.CommandGenericSuccess(w, r)
}
