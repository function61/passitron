package httpserver

import (
	"context"
	"github.com/function61/gokit/cryptoutil"
	"github.com/function61/gokit/dynversion"
	"github.com/function61/gokit/httputils"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/taskrunner"
	"github.com/function61/passitron/pkg/apitypes"
	"github.com/function61/passitron/pkg/commands"
	"github.com/function61/passitron/pkg/extractpublicfiles"
	"github.com/function61/passitron/pkg/f61ui"
	"github.com/function61/passitron/pkg/restqueryapi"
	"github.com/function61/passitron/pkg/signingapi"
	"github.com/function61/passitron/pkg/state"
	"github.com/function61/passitron/pkg/u2futil"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	certFile = "cert.pem"
	keyFile  = "key.pem"
)

func Run(ctx context.Context, logger *log.Logger) error {
	downloadUrl := extractpublicfiles.BintrayDownloadUrl(
		"function61",
		"dl",
		"pi-security-module/"+dynversion.Version+"/"+extractpublicfiles.PublicFilesArchiveFilename)
	if dynversion.IsDevVersion() { // cannot be downloaded
		downloadUrl = ""
	}

	if err := extractpublicfiles.Run(downloadUrl, extractpublicfiles.PublicFilesArchiveFilename, logex.Prefix("extractpublicfiles", logger)); err != nil {
		return err
	}

	appState, err := state.New(logex.Prefix("state", logger))
	if err != nil {
		return err
	}

	handler, err := createHandler(appState, logger)
	if err != nil {
		return err
	}

	// FIXME: remove this crap bubblegum (uses global state)
	certBytes, errReadCertBytes := ioutil.ReadFile(certFile)
	if errReadCertBytes != nil {
		return errReadCertBytes
	}

	cert, err := cryptoutil.ParsePemX509Certificate(certBytes)
	if err != nil {
		return err
	}

	u2futil.InjectCommonNameFromSslCertificate(cert)

	srv := &http.Server{
		Addr:    ":443",
		Handler: handler,
	}

	logl := logex.Levels(logger)

	logl.Info.Printf(
		"%s starting @ %s (cert host %s, expires %s)",
		dynversion.Version,
		srv.Addr,
		cert.Subject.CommonName,
		cert.NotAfter.Format(time.RFC3339))
	defer logl.Info.Printf("stopped")

	if cert.NotAfter.Before(time.Now()) {
		logl.Error.Println("TLS certificate expired")
	}

	tasks := taskrunner.New(ctx, logger)

	tasks.Start("listener "+srv.Addr, func(_ context.Context, _ string) error {
		return httputils.RemoveGracefulServerClosedError(srv.ListenAndServeTLS(certFile, keyFile))
	})

	tasks.Start("listenershutdowner", httputils.ServerShutdownTask(srv))

	return tasks.Wait()
}

func createHandler(appState *state.AppState, logger *log.Logger) (http.Handler, error) {
	router := mux.NewRouter()

	middlewareChains, err := createMiddlewares(appState)
	if err != nil {
		return nil, err
	}

	restqueryapi.Register(router, middlewareChains, appState)

	if err := commands.Register(
		router,
		middlewareChains,
		appState.EventLog,
		appState,
		logex.Prefix("commandapi", logger),
	); err != nil {
		return nil, err
	}

	signingapi.Setup(router, middlewareChains, appState)

	// this most generic catch-all route has to be introduced last
	if err := setupStaticFilesRouting(router, appState); err != nil {
		return nil, err
	}

	return router, nil
}

func setupStaticFilesRouting(router *mux.Router, appState *state.AppState) error {
	assetsPath := "/assets"

	publicFiles := http.FileServer(http.Dir("./public/"))
	router.PathPrefix(assetsPath + "/").Handler(http.StripPrefix(assetsPath+"/", publicFiles))
	router.Handle("/favicon.ico", publicFiles)
	router.Handle("/robots.txt", publicFiles)

	// handle all UI paths
	uiHandler := f61ui.IndexHtmlHandler(assetsPath)
	apitypes.RegisterUiRoutes(router, uiHandler)
	router.HandleFunc("/", uiHandler)

	return nil
}
