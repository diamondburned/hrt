**Package hrt** implements a type-safe HTTP router. It aids in creating a
uniform API interface while making it easier to create API handlers.

HRT stands for (H)TTP (r)outer with (t)ypes.

## Example

Below is a trimmed down version of the Get example in the GoDoc.

```go
type EchoRequest struct {
	What string `query:"what"`
}

type EchoResponse struct {
	What string `json:"what"`
}

func handleEcho(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	return EchoResponse{What: req.What}, nil
}
```

```go
r := chi.NewRouter()
r.Use(hrt.Use(hrt.Opts{
    Encoder:     hrt.JSONEncoder,
    ErrorWriter: hrt.JSONErrorWriter("error"),
}))

r.Get("/echo", hrt.Wrap(handleEcho))
```

## Documentation

For documentation and examples, see [GoDoc](https://godoc.org/libdb.so/hrt/v2).

## Dependencies

HRT depends on [chi v5](https://pkg.go.dev/github.com/go-chi/chi/v5) for URL
parameters when routing. Apps that use HRT should also use chi for routing.

Note that it is still possible to make a custom URL parameter decoder that would
replace chi's, but it is not recommended.
