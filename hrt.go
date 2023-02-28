// Package hrt implements a type-safe HTTP router. It aids in creating a uniform
// API interface while making it easier to create API handlers.
//
// HRT stands for (H)TTP (r)outer with (t)ypes.
package hrt

import (
	"context"
	"net/http"
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
	ErrorWriter: TextErrorWriter,
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
func Wrap[RequestT, ResponseT any](f func(ctx context.Context, req RequestT) (ResponseT, error)) http.HandlerFunc {
	return Handler[RequestT, ResponseT](f).ServeHTTP
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
