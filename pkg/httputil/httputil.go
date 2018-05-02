package httputil

import (
	"encoding/json"
	"log"
	"net/http"
)

type GenericResponse struct {
	Status           string `json:"status"`
	ErrorCode        string `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

func GenericError(code string, err error) *GenericResponse {
	errorDescription := ""
	if err != nil {
		errorDescription = err.Error()
	}

	return &GenericResponse{
		Status:           "error",
		ErrorCode:        code,
		ErrorDescription: errorDescription,
	}
}

func GenericSuccess() *GenericResponse {
	return &GenericResponse{
		Status: "success",
	}
}

func RespondHttpJson(out interface{}, httpCode int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	// https://stackoverflow.com/a/2068407
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	w.WriteHeader(httpCode)

	if err := json.NewEncoder(w).Encode(out); err != nil {
		log.Printf("respondHttpJson: failed to encode JSON: %s", err.Error())
	}
}
