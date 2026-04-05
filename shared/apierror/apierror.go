package apierror

import (
	"encoding/json"
	"net/http"
)

type ErrorCode string

const (
	ErrValidation ErrorCode = "VALIDATION_ERROR"
	ErrNotFound   ErrorCode = "NOT_FOUND"
	ErrUpstream   ErrorCode = "UPSTREAM_ERROR"
	ErrInternal   ErrorCode = "INTERNAL_ERROR"
)

type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details any       `json:"details,omitempty"`
}

type errorResponse struct {
	Error APIError `json:"error"`
}

func Respond(w http.ResponseWriter, status int, code ErrorCode, message string, details any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{
		Error: APIError{Code: code, Message: message, Details: details},
	})
}

func ValidationError(w http.ResponseWriter, message string, details any) {
	Respond(w, http.StatusBadRequest, ErrValidation, message, details)
}

func NotFoundError(w http.ResponseWriter, message string, details any) {
	Respond(w, http.StatusNotFound, ErrNotFound, message, details)
}

func UpstreamError(w http.ResponseWriter, message string) {
	Respond(w, http.StatusBadGateway, ErrUpstream, message, nil)
}

func InternalError(w http.ResponseWriter, message string) {
	Respond(w, http.StatusInternalServerError, ErrInternal, message, nil)
}
