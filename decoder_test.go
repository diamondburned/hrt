package hrt

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestURLDecoder(t *testing.T) {
	type Mega struct {
		String     string     `form:"string"`
		Number     float64    `form:"number"`
		Integer    int        `form:"integer"`
		Time       time.Time  `form:"time"`
		OptString  *string    `form:"optstring"`
		OptNumber  *float64   `form:"optnumber"`
		OptInteger *int       `form:"optinteger"`
		OptTime    *time.Time `form:"opttime"`
	}

	tests := []struct {
		name   string
		input  url.Values
		expect result[Mega]
	}{
		{
			name: "only required fields",
			input: url.Values{
				"string":  {"hello"},
				"number":  {"3.14"},
				"integer": {"42"},
				"time":    {"2021-01-01T00:00:00Z"},
			},
			expect: okResult(Mega{
				String:  "hello",
				Number:  3.14,
				Integer: 42,
				Time:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			}),
		},
		{
			name: "only optional fields",
			input: url.Values{
				"optstring":  {"world"},
				"optnumber":  {"2.71"},
				"optinteger": {"24"},
				"opttime":    {"2020-01-01T00:00:00Z"},
			},
			expect: okResult(Mega{
				OptString:  ptrTo("world"),
				OptNumber:  ptrTo(2.71),
				OptInteger: ptrTo(24),
				OptTime:    ptrTo(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := &http.Request{
				Form: test.input,
			}

			var got Mega
			err := URLDecoder.Decode(req, &got)
			res := combineResult(got, err)

			if !reflect.DeepEqual(test.expect, res) {
				t.Errorf("unexpected test result:\n"+
					"expected: %v\n"+
					"got:      %v\n", test.expect, res)
			}
		})
	}
}

type result[T any] struct {
	value T
	error string
}

func okResult[T any](value T) result[T] {
	return result[T]{value: value}
}

func combineResult[T any](value T, err error) result[T] {
	res := result[T]{value: value}
	if err != nil {
		res.error = err.Error()
	}
	return res
}

func ptrTo[T any](v T) *T { return &v }
