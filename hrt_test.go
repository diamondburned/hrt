package hrt

import (
	"context"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func TestHandler_Introspect(t *testing.T) {
	handler := Wrap(func(ctx context.Context, req echoRequest) (echoResponse, error) {
		return echoResponse{What: req.What}, nil
	})
	introspection, ok := TryIntrospectingHandler(handler)
	if !ok {
		t.Fatal("hrt.Handler is not introspectable")
	}
	t.Log(introspection)
}

type echoRequest struct {
	What string `query:"what"`
}

func (r echoRequest) Validate() error {
	if !strings.HasSuffix(r.What, "!") {
		return errors.New("enthusiasm required")
	}
	return nil
}

type echoResponse struct {
	What string `json:"what"`
}
