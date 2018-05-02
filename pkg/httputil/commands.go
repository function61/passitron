package httputil

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondHttpJson(out interface{}, httpCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)

	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.Printf("respondHttpJson: failed to encode JSON: %s", err.Error())
	}
}

type ResponseError struct {
	ErrorCode        string `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

func CommandValidationError(w http.ResponseWriter, r *http.Request, err error) {
	resp := &ResponseError{
		ErrorCode:        "input_validation_failed",
		ErrorDescription: err.Error(),
	}

	respondHttpJson(resp, http.StatusBadRequest, w)
}

func ErrorIfSealed(w http.ResponseWriter, r *http.Request, unsealed bool) bool {
	if !unsealed {
		CommandCustomError(w, r, "database_is_sealed", nil, http.StatusForbidden)
		return true
	}

	return false
}

func CommandCustomError(w http.ResponseWriter, r *http.Request, code string, err error, httpCode int) {
	errorDescription := ""
	if err != nil {
		errorDescription = err.Error()
	}

	resp := &ResponseError{
		ErrorCode:        code,
		ErrorDescription: errorDescription,
	}

	respondHttpJson(resp, httpCode, w)
}

func CommandGenericSuccess(w http.ResponseWriter, r *http.Request) {
	type ResponseSuccess struct {
		Status string `json:"status"`
	}

	respondHttpJson(&ResponseSuccess{
		Status: "success",
	}, http.StatusOK, w)
}
