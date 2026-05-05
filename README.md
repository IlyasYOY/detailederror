# detailederror

Error handling improvement we all wanted.

Feel free to fork the library to fit your needs. You might want to add some
extra integration to suit your project.

More usage examples are in [tests](./detailederror_test.go) and
[examples](./example_test.go).

## Details

```go
func ExampleGetDetails() {
	err := errors.New("test")
	detailedErr := detailederror.WithMany(
		err,
		"user1", "ilya1",
		"user2", "ilya2",
	)

	got := detailederror.GetDetails(detailedErr)

	for k, v := range got {
		fmt.Println(k, v)
	}

	// Unordered output:
	// user1 ilya1
	// user2 ilya2
}
```

## slog integrations

Details can be logged as structured fields with the standard [slog](https://pkg.go.dev/log/slog)
package. The middleware constructors log only returned errors. They do not
recover panics and do not log panics.

### HTTP

```go
logger := slog.Default()
middleware := detailederror.NewHTTPMiddleware(logger)

handler := middleware(func(w http.ResponseWriter, r *http.Request) error {
	err := errors.New("fetch user")
	return detailederror.WithMany(
		err,
		"user_id", "42",
		"operation", "fetch",
	)
})

http.HandleFunc("/users/42", handler)
```

If the handler returns the error above, the middleware logs the error message and
adds `user_id` and `operation` as structured `slog` fields.

### gRPC

```go
logger := slog.Default()
interceptor := detailederror.NewGRPCUnaryServerInterceptor(logger)

server := grpc.NewServer(
	grpc.UnaryInterceptor(interceptor),
)
```

The interceptor logs non-nil errors returned by unary RPC handlers. Details
attached with `With` or `WithMany` are added as top-level structured fields.

## Development

- [Makefile](./Makefile),
- [detailederror.go](./detailederror.go).

They are all you need! Documentation and make goals are there.

- `make test-watch` starts watcher process with test-on-save using
  [gotestsum](https://github.com/gotestyourself/gotestsum).
- `make` runs all checks at once.
