package hrt_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"libdb.so/hrt/v2"
	"libdb.so/hrt/v2/internal/ht"
)

// EchoRequest is a simple request type that echoes the request.
type EchoRequest struct {
	What string `query:"what"`
}

// Validate implements the hrt.Validator interface.
func (r EchoRequest) Validate() error {
	if !strings.HasSuffix(r.What, "!") {
		return errors.New("enthusiasm required")
	}
	return nil
}

// EchoResponse is a simple response that follows after EchoRequest.
type EchoResponse struct {
	What string `json:"what"`
}

func handleEcho(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	return EchoResponse{What: req.What}, nil
}

func Example_get() {
	r := hrt.NewRouter()
	r.Use(hrt.Use(hrt.DefaultOpts))
	r.Get("/echo", hrt.Wrap(handleEcho))

	srv := ht.NewServer(r)
	defer srv.Close()

	resp := srv.MustGet("/echo", url.Values{"what": {"hi"}})
	fmt.Printf("HTTP %d: %s", resp.Status, resp.Body)

	resp = srv.MustGet("/echo", url.Values{"what": {"hi!"}})
	fmt.Printf("HTTP %d: %s", resp.Status, resp.Body)

	// Output:
	// HTTP 400: {"error":"400: enthusiasm required"}
	// HTTP 200: {"what":"hi!"}
}
