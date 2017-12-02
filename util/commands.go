package util

import (
	"encoding/json"
	"net/http"
)

type ResponseError struct {
	ErrorCode        string `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

func CommandValidationError(w http.ResponseWriter, r *http.Request, err error) {
	resp := ResponseError{
		ErrorCode:        "input_validation_failed",
		ErrorDescription: err.Error(),
	}

	respJson, marshalErr := json.Marshal(resp)
	if marshalErr != nil {
		http.Error(w, "Error marshaling JSON: "+marshalErr.Error(), http.StatusInternalServerError)
		return
	}

	http.Error(w, string(respJson), http.StatusBadRequest)
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

	resp := ResponseError{
		ErrorCode:        code,
		ErrorDescription: errorDescription,
	}

	respJson, marshalErr := json.Marshal(resp)
	if marshalErr != nil {
		http.Error(w, "Error marshaling JSON: "+marshalErr.Error(), http.StatusInternalServerError)
		return
	}

	http.Error(w, string(respJson), httpCode)
}

func CommandGenericSuccess(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"status\": \"success\"}"))
}
