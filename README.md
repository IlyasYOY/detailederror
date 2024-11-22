# detailederror

Error handling improvement we all wanted.

Feel free to fork the library to fit your needs. You might want to add some
extra integration to suit your project.

More usage examples are in [tests](./detailederror_test.go) and
[examples](./example_test.go).

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
