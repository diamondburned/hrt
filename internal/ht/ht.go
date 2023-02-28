// Package ht contains HTTP testing utilities.
package ht

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// Response is a minimal HTTP response.
type Response struct {
	Status int
	Body   []byte
}

// Server is a test server that can be used to test HTTP handlers.
type Server struct {
	httptest.Server
}

// NewServer creates a new test server with the given handler.
func NewServer(h http.Handler) *Server {
	s := &Server{*httptest.NewUnstartedServer(h)}
	s.Start()
	return s
}

// Close closes the server and cleans up any resources.
func (s *Server) Close() {
	s.Server.Close()
}

// MustGet performs a GET request to the given path with the given query
// parameters. It panics if the request fails.
func (s *Server) MustGet(path string, v url.Values) Response {
	url := s.URL + path
	if v != nil {
		url += "?" + v.Encode()
	}

	r, err := s.Client().Get(url)
	if err != nil {
		panic(err)
	}

	defer r.Body.Close()

	rbody, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	return Response{r.StatusCode, rbody}
}

// MustPost performs a POST request to the given path with the given value to
// be used as a JSON body.
func (s *Server) MustPost(path, contentType string, body []byte) Response {
	r, err := s.Client().Post(s.URL+path, contentType, bytes.NewReader(body))
	if err != nil {
		panic(err)
	}

	defer r.Body.Close()

	rbody, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	return Response{r.StatusCode, rbody}
}

// AsJSON unmarshals the response body as JSON. It panics if the unmarshaling
// fails.
func AsJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
