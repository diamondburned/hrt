package hrt

import (
	"encoding/json"
	"net/http"
)

// DefaultEncoder is the default encoder used by the router. It decodes GET
// requests using the query string and URL parameter; everything else uses JSON.
var DefaultEncoder = CombinedEncoder{
	Encoder: JSONEncoder,
	Decoder: MethodDecoder{
		// For the sake of being RESTful, we use a URLDecoder for GET requests.
		"GET": URLDecoder,
		// Everything else will be decoded as JSON.
		"*": JSONEncoder,
	},
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

// Encode implements the Encoder interface.
func (e CombinedEncoder) Encode(w http.ResponseWriter, v any) error {
	return e.Encoder.Encode(w, v)
}

// Decode implements the Decoder interface.
func (e CombinedEncoder) Decode(r *http.Request, v any) error {
	return e.Decoder.Decode(r, v)
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
	if err := e.enc.Decode(r, v); err != nil {
		return err
	}

	if validator, ok := v.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	return nil
}
