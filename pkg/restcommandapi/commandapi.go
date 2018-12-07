package restcommandapi

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/pkg/apitypes"
	"github.com/function61/pi-security-module/pkg/auth"
	"github.com/function61/pi-security-module/pkg/command"
	"github.com/function61/pi-security-module/pkg/commandhandlers"
	"github.com/function61/pi-security-module/pkg/event"
	"github.com/function61/pi-security-module/pkg/httputil"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func Register(router *mux.Router, mwares apitypes.MiddlewareChainMap, st *state.State) error {
	jwtSigner, err := auth.NewJwtSigner(st.GetJwtSigningKey())
	if err != nil {
		return err
	}

	router.HandleFunc("/command/{commandName}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commandName := mux.Vars(r)["commandName"]

		cmdStructBuilder, commandExists := commandhandlers.StructBuilders[commandName]
		if !commandExists {
			httputil.RespondHttpJson(httputil.GenericError("unsupported_command", nil), http.StatusBadRequest, w)
			return
		}

		cmdStruct := cmdStructBuilder()

		reqCtx := mwares[cmdStruct.MiddlewareChain()](w, r)
		if reqCtx == nil {
			return // middleware dealt with error response
		}

		userId := ""
		if reqCtx.User != nil {
			userId = reqCtx.User.Id
		}

		ctx := &command.Ctx{
			RemoteAddr: r.RemoteAddr,
			UserAgent:  r.Header.Get("User-Agent"),
			State:      st,
			Meta:       event.Meta(time.Now(), userId),
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

		if err := st.EventLog.Append(raisedEvents); err != nil {
			httputil.RespondHttpJson(httputil.GenericError("event_append_failed", err), http.StatusInternalServerError, w)
			return
		}

		if ctx.SendLoginCookieUserId != "" {
			token := jwtSigner.Sign(auth.UserDetails{
				Id: ctx.SendLoginCookieUserId,
			})
			http.SetCookie(w, auth.ToCookie(token))
		}

		httputil.RespondHttpJson(httputil.GenericSuccess(), http.StatusOK, w)
	})).Methods(http.MethodPost)

	return nil
}
