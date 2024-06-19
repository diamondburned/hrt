// Package rfutil contains reflect utilities.
package rfutil

import (
	"encoding"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
)

var textUnmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()

// SetPrimitiveFromString sets the value of a primitive type from a string. It
// supports strings, ints, uints, floats and bools. If s is empty, the value is
// left untouched.
func SetPrimitiveFromString(rf reflect.Type, rv reflect.Value, s string) error {
	if s == "" {
		return nil
	}

	if rf.Kind() == reflect.Ptr {
		rf = rf.Elem()

		newValue := reflect.New(rf)
		rv.Set(newValue)
		rv = newValue.Elem()
	}

	switch rf.Kind() {
	case reflect.String:
		rv.SetString(s)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, rf.Bits())
		if err != nil {
			return errors.Wrap(err, "invalid int")
		}
		rv.SetInt(i)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(s, 10, rf.Bits())
		if err != nil {
			return errors.Wrap(err, "invalid uint")
		}
		rv.SetUint(i)
		return nil

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, rf.Bits())
		if err != nil {
			return errors.Wrap(err, "invalid float")
		}
		rv.SetFloat(f)
		return nil

	case reflect.Bool:
		// False means omitted according to MDN.
		rv.SetBool(s != "")
		return nil
	}

	if reflect.PointerTo(rf).Implements(textUnmarshalerType) {
		unmarshaler := rv.Addr().Interface().(encoding.TextUnmarshaler)
		if err := unmarshaler.UnmarshalText([]byte(s)); err != nil {
			return errors.Wrap(err, "text unmarshaling")
		}
	}

	return nil
}

// EachStructField calls the given function for each field of the given struct.
func EachStructField(v any, f func(reflect.StructField, reflect.Value) error) error {
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

		if err := f(rft, rfv); err != nil {
			return err
		}
	}

	return nil
}
