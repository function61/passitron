package restcommandapi

import (
	"errors"
	"github.com/function61/eventkit/eventlog"
	"github.com/function61/eventkit/httpcommand"
	"github.com/function61/gokit/httpauth"
	"github.com/function61/pi-security-module/pkg/commands"
	"github.com/function61/pi-security-module/pkg/httputil"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func Register(
	router *mux.Router,
	mwares httpauth.MiddlewareChainMap,
	eventLog eventlog.Log,
	appState *state.AppState,
	logger *log.Logger,
) error {
	handlers := commands.New(appState, logger)

	invoker := commands.CommandInvoker(handlers)

	router.HandleFunc("/command/{commandName}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		commandName := mux.Vars(r)["commandName"]

		httpErr := httpcommand.Serve(w, r, mwares, commandName, commands.Allocators, invoker, eventLog)
		if httpErr != nil {
			if httpErr.StatusCode > 0 {
				httputil.RespondHttpJson(httputil.GenericError(
					httpErr.ErrorCode,
					errors.New(httpErr.Description)),
					httpErr.StatusCode,
					w)
			}
		} else {
			httputil.RespondHttpJson(httputil.GenericSuccess(), http.StatusOK, w)
		}
	})).Methods(http.MethodPost)

	return nil
}
