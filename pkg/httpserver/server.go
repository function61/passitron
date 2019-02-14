package httpserver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/function61/gokit/cryptoutil"
	"github.com/function61/gokit/dynversion"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/stopper"
	"github.com/function61/pi-security-module/pkg/extractpublicfiles"
	"github.com/function61/pi-security-module/pkg/f61ui"
	"github.com/function61/pi-security-module/pkg/restcommandapi"
	"github.com/function61/pi-security-module/pkg/restqueryapi"
	"github.com/function61/pi-security-module/pkg/signingapi"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/function61/pi-security-module/pkg/u2futil"
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

func Run(stop *stopper.Stopper, logger *log.Logger) error {
	defer stop.Done()

	downloadUrl := extractpublicfiles.PublicFilesDownloadUrl(dynversion.Version)
	if dynversion.IsDevVersion() { // cannot be downloaded
		downloadUrl = ""
	}

	if err := extractpublicfiles.Run(downloadUrl); err != nil {
		return err
	}

	appState := state.New(logex.Prefix("state", logger))
	defer appState.Close()

	handler, err := createHandler(appState, logger)
	if err != nil {
		return err
	}

	// FIXME: remove this crap bubblegum (uses global state)
	certBytes, errReadCertBytes := ioutil.ReadFile(certFile)
	if errReadCertBytes != nil {
		return errReadCertBytes
	}

	cert, err := cryptoutil.ParsePemX509Certificate(bytes.NewBuffer(certBytes))
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
		"Serving @ %s (cert host %s, expires %s)",
		srv.Addr,
		cert.Subject.CommonName,
		cert.NotAfter.Format(time.RFC3339))

	if cert.NotAfter.Before(time.Now()) {
		logl.Error.Println("Certificate expired")
	}

	defer logl.Info.Println("Stopped")

	go func() {
		<-stop.Signal

		if err := srv.Shutdown(context.TODO()); err != nil {
			logl.Error.Printf("Shutdown(): %s", err.Error())
		}
	}()

	if err := srv.ListenAndServeTLS(certFile, keyFile); err != http.ErrServerClosed {
		return fmt.Errorf("ListenAndServeTLS(): %s", err.Error())
	}

	return nil
}

func createHandler(appState *state.AppState, logger *log.Logger) (http.Handler, error) {
	router := mux.NewRouter()

	middlewareChains, err := createMiddlewares(appState)
	if err != nil {
		return nil, err
	}

	restqueryapi.Register(router, middlewareChains, appState)

	if err := restcommandapi.Register(
		router,
		middlewareChains,
		appState.EventLog,
		appState,
		logex.Prefix("commandapi", logger),
	); err != nil {
		return nil, err
	}

	signingapi.Setup(router, appState)

	// this most generic catch-all route has to be introduced last
	if err := setupStaticFilesRouting(router, appState); err != nil {
		return nil, err
	}

	return router, nil
}

func setupStaticFilesRouting(router *mux.Router, appState *state.AppState) error {
	assetsPath := "/assets"

	publicFiles := http.FileServer(http.Dir("./public/"))
	router.HandleFunc("/", f61ui.IndexHtmlHandler(assetsPath, appState.GetCsrfToken()))
	router.PathPrefix(assetsPath + "/").Handler(http.StripPrefix(assetsPath+"/", publicFiles))
	router.Handle("/favicon.ico", publicFiles)
	router.Handle("/robots.txt", publicFiles)

	return nil
}
