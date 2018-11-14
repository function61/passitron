package restcommandapi

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/pkg/command"
	"github.com/function61/pi-security-module/pkg/commandhandlers"
	"github.com/function61/pi-security-module/pkg/domain"
	"github.com/function61/pi-security-module/pkg/httputil"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

func Register(router *mux.Router, st *state.State) {
	router.HandleFunc("/command/{commandName}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commandName := mux.Vars(r)["commandName"]

		cmdStructBuilder, commandExists := commandhandlers.StructBuilders[commandName]
		if !commandExists {
			httputil.RespondHttpJson(httputil.GenericError("unsupported_command", nil), http.StatusBadRequest, w)
			return
		}

		cmdStruct := cmdStructBuilder()

		if cmdStruct.RequiresAuthentication() && errorIfSealed(st.IsUnsealed(), w) {
			return
		}

		ctx := &command.Ctx{
			State: st,
			Meta:  domain.Meta(time.Now(), domain.DefaultUserIdTODO),
		}

		if r.Header.Get("Content-Type") != "application/json" {
			httputil.RespondHttpJson(httputil.GenericError("expecting_content_type_json", errors.New("expecting Content-Type header with application/json")), http.StatusBadRequest, w)
			return
		}

		jsonDecoder := json.NewDecoder(r.Body)
		jsonDecoder.DisallowUnknownFields()
		if errJson := jsonDecoder.Decode(cmdStruct); errJson != nil {
			httputil.RespondHttpJson(httputil.GenericError("json_parsing_failed", errJson), http.StatusBadRequest, w)
			return
		}

		if errValidate := cmdStruct.Validate(); errValidate != nil {
			httputil.RespondHttpJson(httputil.GenericError("command_validation_failed", errValidate), http.StatusBadRequest, w)
			return
		}

		if errInvoke := cmdStruct.Invoke(ctx); errInvoke != nil {
			httputil.RespondHttpJson(httputil.GenericError("command_failed", errInvoke), http.StatusBadRequest, w)
			return
		}

		raisedEvents := ctx.GetRaisedEvents()

		log.Printf("Command %s raised %d event(s)", commandName, len(raisedEvents))

		st.EventLog.AppendBatch(raisedEvents)

		httputil.RespondHttpJson(httputil.GenericSuccess(), http.StatusOK, w)
	}))
}

func errorIfSealed(unsealed bool, w http.ResponseWriter) bool {
	if !unsealed {
		httputil.RespondHttpJson(httputil.GenericError("database_is_sealed", nil), http.StatusForbidden, w)
		return true
	}

	return false
}
