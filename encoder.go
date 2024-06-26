package hrt

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

// DefaultEncoder is the default encoder used by the router. It decodes GET
// requests using the query string and URL parameter; everything else uses JSON.
//
// For the sake of being RESTful, we use a URLDecoder for GET requests.
// Everything else will be decoded as JSON.
var DefaultEncoder = CombinedEncoder{
	Encoder: EncoderWithValidator(JSONEncoder),
	Decoder: DecoderWithValidator(MethodDecoder{
		"GET": URLDecoder,
		"*":   JSONEncoder,
	}),
}

// Encoder describes an encoder that encodes or decodes the request and response
// types.
type Encoder interface {
	// Encode encodes the given value into the given writer.
	Encode(http.ResponseWriter, any) error
	// An encoder must be able to decode the same type it encodes.
	Decoder
}

// CombinedEncoder combines an encoder and decoder pair into one.
type CombinedEncoder struct {
	Encoder Encoder
	Decoder Decoder
}

var _ Encoder = CombinedEncoder{}

// Encode implements the Encoder interface.
func (e CombinedEncoder) Encode(w http.ResponseWriter, v any) error {
	return e.Encoder.Encode(w, v)
}

// Decode implements the Decoder interface.
func (e CombinedEncoder) Decode(r *http.Request, v any) error {
	return e.Decoder.Decode(r, v)
}

// UnencodableEncoder is an encoder that can only decode and not encode.
// It wraps an existing decoder.
// Calling Encode will return a 500 error, as it is considered a bug to return
// anything.
type UnencodableEncoder struct {
	Decoder
}

var _ Encoder = UnencodableEncoder{}

func (e UnencodableEncoder) Encode(w http.ResponseWriter, v any) error {
	return WrapHTTPError(http.StatusInternalServerError, errors.New("cannot encode"))
}

// JSONEncoder is an encoder that encodes and decodes JSON.
var JSONEncoder Encoder = jsonEncoder{}

type jsonEncoder struct{}

func (e jsonEncoder) Encode(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func (e jsonEncoder) Decode(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// Validator describes a type that can validate itself.
type Validator interface {
	Validate() error
}

// EncoderWithValidator wraps an encoder with one that calls Validate() on the
// value after decoding and before encoding if the value implements Validator.
func EncoderWithValidator(enc Encoder) Encoder {
	return validatorEncoder{enc}
}

type validatorEncoder struct{ enc Encoder }

func (e validatorEncoder) Encode(w http.ResponseWriter, v any) error {
	if validator, ok := v.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	if err := e.enc.Encode(w, v); err != nil {
		return err
	}

	return nil
}

func (e validatorEncoder) Decode(r *http.Request, v any) error {
	return (validatorDecoder{e.enc}).Decode(r, v)
}
