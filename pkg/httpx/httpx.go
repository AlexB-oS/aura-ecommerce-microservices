// Package httpx contains small helpers for writing consistent JSON responses
// and a structured API error type. Keeping this in one shared place means every
// microservice speaks the exact same wire format, which makes the frontend and
// the API gateway simpler to reason about.
package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Envelope is the shape of every successful JSON response: { "data": ... }.
// Wrapping payloads keeps room to add top-level metadata (pagination, etc.)
// later without breaking clients.
type Envelope struct {
	Data any `json:"data"`
}

// ErrorBody is the shape of every error response: { "error": { ... } }.
type ErrorBody struct {
	Error APIError `json:"error"`
}

// APIError is a transport-friendly error. `Code` is a stable machine-readable
// string (e.g. "email_taken") the frontend can switch on; `Message` is human
// readable; `Fields` carries per-field validation messages when relevant.
type APIError struct {
	Status  int               `json:"-"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *APIError) Error() string { return e.Message }

// NewError builds an APIError. Use the typed constructors below for common cases.
func NewError(status int, code, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

// WithFields attaches field-level validation messages and returns the error so
// calls can be chained: httpx.ErrValidation("invalid").WithFields(m).
func (e *APIError) WithFields(fields map[string]string) *APIError {
	e.Fields = fields
	return e
}

// Common constructors keep error codes consistent across services.
func ErrBadRequest(msg string) *APIError   { return NewError(http.StatusBadRequest, "bad_request", msg) }
func ErrValidation(msg string) *APIError   { return NewError(http.StatusUnprocessableEntity, "validation_failed", msg) }
func ErrUnauthorized(msg string) *APIError { return NewError(http.StatusUnauthorized, "unauthorized", msg) }
func ErrForbidden(msg string) *APIError    { return NewError(http.StatusForbidden, "forbidden", msg) }
func ErrNotFound(msg string) *APIError     { return NewError(http.StatusNotFound, "not_found", msg) }
func ErrConflict(code, msg string) *APIError {
	return NewError(http.StatusConflict, code, msg)
}
func ErrInternal() *APIError {
	return NewError(http.StatusInternalServerError, "internal_error", "something went wrong on our side")
}

// JSON writes any value as a JSON envelope with the given status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	// A failure here means the connection is already gone; nothing useful to do.
	_ = json.NewEncoder(w).Encode(Envelope{Data: data})
}

// Error inspects err and writes the matching JSON error response. Unknown
// errors are deliberately collapsed to a generic 500 so we never leak internal
// details (SQL strings, stack info) to clients.
func Error(w http.ResponseWriter, err error) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		apiErr = ErrInternal()
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(apiErr.Status)
	_ = json.NewEncoder(w).Encode(ErrorBody{Error: *apiErr})
}

// Decode reads a JSON body into dst, rejecting unknown fields and empty bodies.
// It returns a ready-to-send APIError on failure so handlers stay terse.
func Decode(r *http.Request, dst any) error {
	if r.Body == nil {
		return ErrBadRequest("request body is required")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return ErrBadRequest("request body is not valid JSON: " + err.Error())
	}
	return nil
}
