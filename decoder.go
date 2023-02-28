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
// For the sake of supporting code generators, the decoder also reads the `json`
// tag if the `url` tag is not present. If a struct field has no tag, it is
// assumed to be the same as the field name. If a struct field has a tag, then
// only that tag is used.
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

func lookupURL(r *http.Request, name string) string {
	value := chi.URLParam(r, name)
	if value == "" {
		value = r.FormValue(name)
	}
	return value
}

var urlDecoderTags = []string{"url", "json"}

func (d urlDecoder) Decode(r *http.Request, v any) error {
	return rfutil.EachStructField(v, func(rft reflect.StructField, rfv reflect.Value) error {
		for _, tag := range urlDecoderTags {
			tagValue := rft.Tag.Get(tag)
			if tagValue == "" {
				continue
			}

			val := lookupURL(r, tag)
			if val == "" {
				return nil
			}

			return rfutil.SetPrimitiveFromString(rfv.Type(), rfv, val)
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
