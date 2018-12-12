package httpserver

import (
	"fmt"
	"github.com/function61/gokit/logger"
	"github.com/function61/gokit/stopper"
	"github.com/function61/pi-security-module/pkg/extractpublicfiles"
	"github.com/function61/pi-security-module/pkg/restcommandapi"
	"github.com/function61/pi-security-module/pkg/restqueryapi"
	"github.com/function61/pi-security-module/pkg/signingapi"
	"github.com/function61/pi-security-module/pkg/state"
	"github.com/function61/pi-security-module/pkg/u2futil"
	"github.com/function61/pi-security-module/pkg/version"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	certFile = "cert.pem"
	keyFile  = "key.pem"
)

var log = logger.New("httpserver")

func Run(stop *stopper.Stopper) error {
	defer stop.Done()

	downloadUrl := extractpublicfiles.PublicFilesDownloadUrl(version.Version)
	if version.IsDevVersion() {
		downloadUrl = ""
	}

	if err := extractpublicfiles.Run(downloadUrl); err != nil {
		return err
	}

	st := state.New()
	defer st.Close()

	handler, err := createHandler(st)
	if err != nil {
		return err
	}

	// FIXME: remove this crap bubblegum (uses global state)
	certBytes, errReadCertBytes := ioutil.ReadFile(certFile)
	if errReadCertBytes != nil {
		return errReadCertBytes
	}

	if err := u2futil.InjectCommonNameFromSslCertificate(certBytes); err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    ":443",
		Handler: handler,
	}

	log.Info(fmt.Sprintf("serving @ %s", srv.Addr))
	defer log.Info("stopped")

	go func() {
		<-stop.Signal

		if err := srv.Shutdown(nil); err != nil {
			log.Error(fmt.Sprintf("Shutdown(): %s", err.Error()))
		}
	}()

	if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
		return fmt.Errorf("ListenAndServeTLS(): %s", err.Error())
	}

	return nil
}

func createHandler(st *state.State) (http.Handler, error) {
	router := mux.NewRouter()

	middlewareChains, err := createMiddlewares(st)
	if err != nil {
		return nil, err
	}

	restqueryapi.Register(router, middlewareChains, st)

	if err := restcommandapi.Register(router, middlewareChains, st.EventLog, st); err != nil {
		return nil, err
	}

	signingapi.Setup(router, st)

	// this most generic catch-all route has to be introduced last
	if err := setupStaticFilesRouting(router, st); err != nil {
		return nil, err
	}

	return router, nil
}

func setupStaticFilesRouting(router *mux.Router, st *state.State) error {
	indexTemplate, err := ioutil.ReadFile("public/index.html.template")
	if err != nil {
		return err
	}

	index := strings.Replace(
		string(indexTemplate),
		"[$csrf_token]",
		st.GetCsrfToken(),
		-1)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write([]byte(index)); err != nil {
			panic(err)
		}
	})

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	return nil
}
