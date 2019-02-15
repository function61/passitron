package muxregistrator

import (
	"github.com/gorilla/mux"
	"net/http"
)

func New(router *mux.Router) func(string, string, http.HandlerFunc) {
	return func(method string, path string, fn http.HandlerFunc) {
		router.HandleFunc(path, fn).Methods(method)
	}
}
