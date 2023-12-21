package hrt

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Router redefines [chi.Router] to modify all method-routing functions to
// accept an [http.Handler] instead of a [http.HandlerFunc].
type Router interface {
	http.Handler
	chi.Routes

	// Use appends one or more middlewares onto the Router stack.
	Use(middlewares ...func(http.Handler) http.Handler)

	// With adds inline middlewares for an endpoint handler.
	With(middlewares ...func(http.Handler) http.Handler) Router

	// Group adds a new inline-Router along the current routing
	// path, with a fresh middleware stack for the inline-Router.
	Group(fn func(r Router)) Router

	// Route mounts a sub-Router along a `pattern“ string.
	Route(pattern string, fn func(r Router)) Router

	// Mount attaches another http.Handler along ./pattern/*
	Mount(pattern string, h http.Handler)

	// Handle and HandleFunc adds routes for `pattern` that matches
	// all HTTP methods.
	Handle(pattern string, h http.Handler)
	HandleFunc(pattern string, h http.HandlerFunc)

	// Method and MethodFunc adds routes for `pattern` that matches
	// the `method` HTTP method.
	Method(method, pattern string, h http.Handler)
	MethodFunc(method, pattern string, h http.HandlerFunc)

	// HTTP-method routing along `pattern`
	Connect(pattern string, h http.Handler)
	Delete(pattern string, h http.Handler)
	Get(pattern string, h http.Handler)
	Head(pattern string, h http.Handler)
	Options(pattern string, h http.Handler)
	Patch(pattern string, h http.Handler)
	Post(pattern string, h http.Handler)
	Put(pattern string, h http.Handler)
	Trace(pattern string, h http.Handler)

	// NotFound defines a handler to respond whenever a route could
	// not be found.
	NotFound(h http.HandlerFunc)

	// MethodNotAllowed defines a handler to respond whenever a method is
	// not allowed.
	MethodNotAllowed(h http.HandlerFunc)
}

// NewRouter creates a [chi.Router] wrapper that turns all method-routing
// functions to take a regular [http.Handler] instead of an [http.HandlerFunc].
// This allows [hrt.Wrap] to function properly. This router also has the given
// opts injected into its context, so there is no need to call [hrt.Use].
func NewRouter(opts Opts) Router {
	r := router{chi.NewRouter()}
	r.Use(Use(opts))
	return r
}

// NewPlainRouter is like [NewRouter] but does not inject any options into the
// context.
func NewPlainRouter() Router {
	return router{chi.NewRouter()}
}

// WrapRouter wraps a [chi.Router] to turn all method-routing functions to take
// a regular [http.Handler] instead of an [http.HandlerFunc]. This allows
// [hrt.Wrap] to function properly.
func WrapRouter(r chi.Router) Router {
	return router{r}
}

type router struct{ chi.Router }

func (r router) With(middlewares ...func(http.Handler) http.Handler) Router {
	return router{r.Router.With(middlewares...)}
}

func (r router) Group(fn func(r Router)) Router {
	return router{r.Router.Group(func(r chi.Router) {
		fn(router{r})
	})}
}

func (r router) Route(pattern string, fn func(r Router)) Router {
	return router{r.Router.Route(pattern, func(r chi.Router) {
		fn(router{r})
	})}
}

func (r router) Connect(pattern string, h http.Handler) {
	r.Router.Method("connect", pattern, h)
}

func (r router) Delete(pattern string, h http.Handler) {
	r.Router.Method("delete", pattern, h)
}

func (r router) Get(pattern string, h http.Handler) {
	r.Router.Method("get", pattern, h)
}

func (r router) Head(pattern string, h http.Handler) {
	r.Router.Method("head", pattern, h)
}

func (r router) Options(pattern string, h http.Handler) {
	r.Router.Method("options", pattern, h)
}

func (r router) Patch(pattern string, h http.Handler) {
	r.Router.Method("patch", pattern, h)
}

func (r router) Post(pattern string, h http.Handler) {
	r.Router.Method("post", pattern, h)
}

func (r router) Put(pattern string, h http.Handler) {
	r.Router.Method("put", pattern, h)
}

func (r router) Trace(pattern string, h http.Handler) {
	r.Router.Method("trace", pattern, h)
}
