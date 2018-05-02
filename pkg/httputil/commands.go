package httputil

import (
	"net/http"
)

type SimpleStatusedResponse struct {
	Status string `json:"status"`
}

// TODO: combine these

type ResponseError struct {
	ErrorCode        string `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

func CommandValidationError(w http.ResponseWriter, r *http.Request, err error) {
	RespondHttpJson(&ResponseError{
		ErrorCode:        "input_validation_failed",
		ErrorDescription: err.Error(),
	}, http.StatusBadRequest, w)
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

	RespondHttpJson(&ResponseError{
		ErrorCode:        code,
		ErrorDescription: errorDescription,
	}, httpCode, w)
}

func GenericSuccess() *SimpleStatusedResponse {
	return &SimpleStatusedResponse{
		Status: "success",
	}
}
