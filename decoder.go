package hrt

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/diamondburned/hrt/internal/rfutil"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

// Decoder describes a decoder that decodes the request type.
type Decoder interface {
	// Decode decodes the given value from the given reader.
	Decode(*http.Request, any) error
}

// MethodDecoder is an encoder that only encodes or decodes if the request
// method matches the methods in it.
type MethodDecoder map[string]Decoder

// Decode implements the Decoder interface.
func (e MethodDecoder) Decode(r *http.Request, v any) error {
	dec, ok := e[r.Method]
	if !ok {
		dec, ok = e["*"]
	}
	if !ok {
		return WrapHTTPError(http.StatusMethodNotAllowed, errors.New("method not allowed"))
	}
	return dec.Decode(r, v)
}

// URLDecoder decodes chi.URLParams and url.Values into a struct. It only does
// Decoding; the Encode method is a no-op. The decoder makes no effort to
// traverse the struct and decode nested structs. If neither a chi.URLParam nor
// a url.Value is found for a field, the field is left untouched.
//
// The following tags are supported:
//
// - `url` - uses chi.URLParam to decode the value.
// - `form` - uses r.FormValue to decode the value.
// - `query` - similar to `form`.
// - `schema` - similar to `form`, exists for compatibility with gorilla/schema.
// - `json` - uses either chi.URLParam or r.FormValue to decode the value.
//   This exists for compatibility with code generators.
//
// If a struct field has no tag, it is assumed to be the same as the field name.
// If a struct field has a tag, then only that tag is used.
//
// # Example
//
// The following Go type would be decoded to have 2 URL parameters:
//
//    type Data struct {
//        ID  string
//        Num int `url:"num"`
//        Nested struct {
//            ID string
//        }
//    }
//
var URLDecoder Decoder = urlDecoder{}

type urlDecoder struct{}

func (d urlDecoder) Decode(r *http.Request, v any) error {
	return rfutil.EachStructField(v, func(rft reflect.StructField, rfv reflect.Value) error {
		for _, tag := range []string{"form", "query", "schema"} {
			if tagValue := rft.Tag.Get(tag); tagValue != "" {
				val := r.FormValue(tagValue)
				return rfutil.SetPrimitiveFromString(rft.Type, rfv, val)
			}
		}

		if tagValue := rft.Tag.Get("url"); tagValue != "" {
			val := chi.URLParam(r, tagValue)
			return rfutil.SetPrimitiveFromString(rft.Type, rfv, val)
		}

		if tagValue := rft.Tag.Get("json"); tagValue != "" {
			val := chi.URLParam(r, tagValue)
			if val == "" {
				val = r.FormValue(tagValue)
			}
			return rfutil.SetPrimitiveFromString(rft.Type, rfv, val)
		}

		// Search for the URL parameters manually.
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			for i, k := range rctx.URLParams.Keys {
				if strings.EqualFold(k, rft.Name) {
					return rfutil.SetPrimitiveFromString(rfv.Type(), rfv, rctx.URLParams.Values[i])
				}
			}
		}

		// Trigger form parsing.
		r.FormValue("")

		// Search for URL form values manually.
		for k, v := range r.Form {
			if strings.EqualFold(k, rft.Name) {
				return rfutil.SetPrimitiveFromString(rfv.Type(), rfv, v[0])
			}
		}

		return nil // ignore
	})
}

// DecoderWithValidator wraps an encoder with one that calls Validate() on the
// value after decoding and before encoding if the value implements Validator.
func DecoderWithValidator(enc Decoder) Decoder {
	return validatorDecoder{enc}
}

type validatorDecoder struct{ dec Decoder }

func (e validatorDecoder) Decode(r *http.Request, v any) error {
	if err := e.dec.Decode(r, v); err != nil {
		return err
	}

	if validator, ok := v.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	return nil
}
