package command

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/secret/event"
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.ApplyEvents([]interface{}{
		event.OtpTokenSet{
			Event:              eventbase.NewEvent(),
			Id:                 req.Id,
			OtpProvisioningUrl: req.OtpProvisioningUrl,
		},
	})

	w.Write([]byte("OK"))
}
