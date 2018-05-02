package httputil

import (
	"net/http"
)

type ResponseError struct {
	ErrorCode        string `json:"error_code"`
	ErrorDescription string `json:"error_description"`
}

func CommandValidationError(w http.ResponseWriter, r *http.Request, err error) {
	resp := &ResponseError{
		ErrorCode:        "input_validation_failed",
		ErrorDescription: err.Error(),
	}

	RespondHttpJson(resp, http.StatusBadRequest, w)
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

	RespondHttpJson(resp, httpCode, w)
}

func CommandGenericSuccess(w http.ResponseWriter, r *http.Request) {
	type ResponseSuccess struct {
		Status string `json:"status"`
	}

	RespondHttpJson(&ResponseSuccess{
		Status: "success",
	}, http.StatusOK, w)
}
