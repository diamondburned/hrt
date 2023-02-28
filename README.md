Package hrt implements a type-safe HTTP router. It aids in creating a uniform
API interface while making it easier to create API handlers.

HRT stands for (H)TTP (r)outer with (t)ypes.

## Documentation

For documentation and examples, see [GoDoc](https://godoc.org/github.com/diamondburned/hrt).

## Dependencies

HRT depends on [chi v5](https://pkg.go.dev/github.com/go-chi/chi/v5) for URL
parameters when routing. Apps that use HRT should also use chi for routing.

Note that it is still possible to make a custom URL parameter decoder that would
replace chi's, but it is not recommended.
