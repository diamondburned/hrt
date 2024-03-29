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

// WrapHTTPError wraps an error with an HTTP status code. If the error is
// already of type HTTPError, it is returned as-is. To change the HTTP status
// code, use OverrideHTTPError.
func WrapHTTPError(code int, err error) HTTPError {
	var httpErr HTTPError
	if errors.As(err, &httpErr) {
		return httpErr
	}
	return wrappedHTTPError{code, err}
}

// NewHTTPError creates a new HTTPError with the given status code and message.
func NewHTTPError(code int, str string) HTTPError {
	return wrappedHTTPError{code, errors.New(str)}
}

// OverrideHTTPError overrides the HTTP status code of the given error. If the
// error is not of type HTTPError, it is wrapped with the given status code. If
// it is, the error is unwrapped and wrapped with the new status code.
func OverrideHTTPError(code int, err error) HTTPError {
	var httpErr HTTPError
	if errors.As(err, &httpErr) {
		err = errors.Unwrap(httpErr)
	}
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
