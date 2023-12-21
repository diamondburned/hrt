package hrt_test

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"libdb.so/hrt/v2"
	"libdb.so/hrt/v2/internal/ht"
)

// User is a simple user type.
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var (
	users   = make(map[int]User)
	usersMu sync.RWMutex
)

// GetUserRequest is a request that fetches a user by ID.
type GetUserRequest struct {
	ID int `url:"id"`
}

// Validate implements the hrt.Validator interface.
func (r GetUserRequest) Validate() error {
	if r.ID == 0 {
		return errors.New("invalid ID")
	}
	return nil
}

func handleGetUser(ctx context.Context, req GetUserRequest) (User, error) {
	usersMu.RLock()
	defer usersMu.RUnlock()

	user, ok := users[req.ID]
	if !ok {
		return User{}, hrt.WrapHTTPError(404, errors.New("user not found"))
	}

	return user, nil
}

// CreateUserRequest is a request that creates a user.
type CreateUserRequest struct {
	Name string `json:"name"`
}

// Validate implements the hrt.Validator interface.
func (r CreateUserRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

func handleCreateUser(ctx context.Context, req CreateUserRequest) (User, error) {
	user := User{
		ID:   len(users) + 1,
		Name: req.Name,
	}

	usersMu.Lock()
	users[user.ID] = user
	usersMu.Unlock()

	return user, nil
}

func Example_post() {
	r := chi.NewRouter()
	r.Use(hrt.Use(hrt.DefaultOpts))
	r.Route("/users", func(r chi.Router) {
		r.Method("get", "/{id}", hrt.Wrap(handleGetUser))
		r.Method("post", "/", hrt.Wrap(handleCreateUser))
	})

	srv := ht.NewServer(r)
	defer srv.Close()

	resps := []ht.Response{
		srv.MustGet("/users/1", nil),
		srv.MustPost("/users", "application/json", ht.AsJSON(map[string]any{})),
		srv.MustPost("/users", "application/json", ht.AsJSON(map[string]any{
			"name": "diamondburned",
		})),
		srv.MustGet("/users/1", nil),
	}

	for _, resp := range resps {
		fmt.Printf("HTTP %d: %s", resp.Status, resp.Body)
	}

	// Output:
	// HTTP 404: {"error":"404: user not found"}
	// HTTP 400: {"error":"400: name is required"}
	// HTTP 200: {"id":1,"name":"diamondburned"}
	// HTTP 200: {"id":1,"name":"diamondburned"}
}
