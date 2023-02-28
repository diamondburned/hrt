package hrt

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// HTTPError extends the error interface with an HTTP status code.
type HTTPError interface {
	error
	HTTPStatus() int
}

// ErrorHTTPStatus returns the HTTP status code for the given error. If the
// error is not an HTTPError, it returns defaultCode.
func ErrorHTTPStatus(err error, defaultCode int) int {
	var httpErr HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.HTTPStatus()
	}
	return defaultCode
}

type wrappedHTTPError struct {
	code int
	err  error
}

// WrapHTTPError wraps an error with an HTTP status code.
func WrapHTTPError(code int, err error) HTTPError {
	return wrappedHTTPError{code, err}
}

func (e wrappedHTTPError) HTTPStatus() int {
	return e.code
}

func (e wrappedHTTPError) Error() string {
	return fmt.Sprintf("%d: %s", e.code, e.err)
}

func (e wrappedHTTPError) Unwrap() error {
	return e.err
}

// ErrorWriter is a writer that writes an error to the response.
type ErrorWriter interface {
	WriteError(w http.ResponseWriter, err error)
}

// WriteErrorFunc is a function that implements the ErrorWriter interface.
type WriteErrorFunc func(w http.ResponseWriter, err error)

// WriteError implements the ErrorWriter interface.
func (f WriteErrorFunc) WriteError(w http.ResponseWriter, err error) {
	f(w, err)
}

// TextErrorWriter writes the error into the response in plain text. 500
// status code is used by default.
var TextErrorWriter ErrorWriter = textErrorWriter{}

type textErrorWriter struct{}

func (textErrorWriter) WriteError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(ErrorHTTPStatus(err, http.StatusInternalServerError))
	fmt.Fprintln(w, err)
}

// JSONErrorWriter writes the error into the response in JSON. 500 status code
// is used by default. The given field is used as the key for the error message.
func JSONErrorWriter(field string) ErrorWriter {
	return WriteErrorFunc(func(w http.ResponseWriter, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(ErrorHTTPStatus(err, http.StatusInternalServerError))

		msg := map[string]any{field: err.Error()}
		json.NewEncoder(w).Encode(msg)
	})
}
