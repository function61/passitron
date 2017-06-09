package command

import (
	"github.com/function61/pi-security-module/util/cryptorandombytes"
	"encoding/json"
	"errors"
	"net/http"
	"github.com/function61/pi-security-module/secret/event"
	"github.com/function61/pi-security-module/util"
)

type SecretCreateRequest struct {
	FolderId string
	Title    string
	Username string
	Password string
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	secretId := cryptorandombytes.Hex(4)

	events := []interface{}{
		event.SecretCreated{
			Id:       secretId,
			FolderId: req.FolderId,
			Title:    req.Title,
		},
	}

	if req.Username != "" {
		events = append(events, event.UsernameChanged{
			Id:       secretId,
			Username: req.Username,
		})
	}

	if req.Password != "" {
		events = append(events, event.PasswordChanged{
			Id:       secretId,
			Password: req.Password,
		})
	}

	util.ApplyEvents(events)

	w.Write([]byte("OK"))
}
