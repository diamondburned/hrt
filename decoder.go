package hrt

import (
	"net/http"
	"reflect"
	"strconv"

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
// tag if the `url` tag is not present.
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
	rv := reflect.Indirect(reflect.ValueOf(v))
	if !rv.IsValid() {
		return errors.New("invalid value")
	}

	if rv.Kind() != reflect.Struct {
		return errors.New("value is not a struct")
	}

	rt := rv.Type()
	nfields := rv.NumField()

	for i := 0; i < nfields; i++ {
		rfv := rv.Field(i)
		rft := rt.Field(i)
		if !rft.IsExported() {
			continue
		}

		var name string
		if tag := rft.Tag.Get("json"); tag != "" {
			name = tag
		} else if tag := rft.Tag.Get("url"); tag != "" {
			name = tag
		} else {
			name = rft.Name
		}

		value := chi.URLParam(r, name)
		if value == "" {
			value = r.FormValue(name)
		}
		if value == "" {
			continue
		}

		setPrimitiveFromString(rfv.Type(), rfv, value)
	}

	return nil
}

func setPrimitiveFromString(rf reflect.Type, rv reflect.Value, s string) error {
	switch rf.Kind() {
	case reflect.String:
		rv.SetString(s)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return errors.Wrap(err, "invalid int")
		}
		rv.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return errors.Wrap(err, "invalid uint")
		}
		rv.SetUint(i)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return errors.Wrap(err, "invalid float")
		}
		rv.SetFloat(f)

	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return errors.Wrap(err, "invalid bool")
		}
		rv.SetBool(b)
	}

	return nil
}
