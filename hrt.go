// Package hrt implements a type-safe HTTP router. It aids in creating a uniform
// API interface while making it easier to create API handlers.
package hrt

import (
	"context"
	"net/http"
	"reflect"
)

type ctxKey uint8

const (
	routerOptsCtxKey ctxKey = iota
	requestCtxKey
)

// RequestFromContext returns the request from the Handler's context.
func RequestFromContext(ctx context.Context) *http.Request {
	return ctx.Value(requestCtxKey).(*http.Request)
}

// Opts contains options for the router.
type Opts struct {
	Encoder     Encoder
	ErrorWriter ErrorWriter
}

// DefaultOpts is the default options for the router.
var DefaultOpts = Opts{
	Encoder:     DefaultEncoder,
	ErrorWriter: JSONErrorWriter("error"),
}

// OptsFromContext returns the options from the Handler's context. DefaultOpts
// is returned if no options are found.
func OptsFromContext(ctx context.Context) Opts {
	opts, ok := ctx.Value(routerOptsCtxKey).(Opts)
	if ok {
		return opts
	}
	return DefaultOpts
}

// WithOpts returns a new context with the given options.
func WithOpts(ctx context.Context, opts Opts) context.Context {
	return context.WithValue(ctx, routerOptsCtxKey, opts)
}

// Use creates a middleware that injects itself into each request's context.
func Use(opts Opts) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := WithOpts(r.Context(), opts)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// None indicates that the request has no body or the request does not return
// anything.
type None struct{}

// Empty is a value of None.
var Empty = None{}

// Handler describes a generic handler that takes in a type and returns a
// response.
type Handler[RequestT, ResponseT any] func(ctx context.Context, req RequestT) (ResponseT, error)

// Wrap wraps a handler into a http.Handler. It exists because Go's type
// inference doesn't work well with the Handler type.
func Wrap[RequestT, ResponseT any](f func(ctx context.Context, req RequestT) (ResponseT, error)) http.Handler {
	return Handler[RequestT, ResponseT](f)
}

// ServeHTTP implements the http.Handler interface.
func (h Handler[RequestT, ResponseT]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req RequestT

	// Context cycle! Let's go!!
	ctx := context.WithValue(r.Context(), requestCtxKey, r)

	opts := OptsFromContext(ctx)
	if _, ok := any(req).(None); !ok {
		if err := opts.Encoder.Decode(r, &req); err != nil {
			opts.ErrorWriter.WriteError(w, WrapHTTPError(http.StatusBadRequest, err))
			return
		}
	}

	resp, err := h(ctx, req)
	if err != nil {
		opts.ErrorWriter.WriteError(w, err)
		return
	}

	if _, ok := any(resp).(None); !ok {
		if err := opts.Encoder.Encode(w, resp); err != nil {
			opts.ErrorWriter.WriteError(w, WrapHTTPError(http.StatusInternalServerError, err))
			return
		}
	}
}

// HandlerIntrospection is a struct that contains information about a handler.
// This is primarily used for documentation.
type HandlerIntrospection struct {
	// FuncType is the type of the function.
	FuncType reflect.Type
	// RequestType is the type of the request parameter.
	RequestType reflect.Type
	// ResponseType is the type of the response parameter.
	ResponseType reflect.Type
}

// TryIntrospectingHandler checks if h is an hrt.Handler and returns its
// introspection if it is, otherwise it returns false.
func TryIntrospectingHandler(h http.Handler) (HandlerIntrospection, bool) {
	type introspector interface {
		Introspect() HandlerIntrospection
	}
	var _ introspector = Handler[None, None](nil)

	if h, ok := h.(introspector); ok {
		return h.Introspect(), true
	}
	return HandlerIntrospection{}, false
}

// Introspect returns information about the handler.
func (h Handler[RequestT, ResponseT]) Introspect() HandlerIntrospection {
	var req RequestT
	var resp ResponseT

	return HandlerIntrospection{
		FuncType:     reflect.TypeOf(h),
		RequestType:  reflect.TypeOf(req),
		ResponseType: reflect.TypeOf(resp),
	}
}
