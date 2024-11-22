package detailederror_test

import (
	"errors"
	"fmt"

	"github.com/IlyasYOY/detailederror"
)

func ExampleGetDetail() {
	key := "user"
	want := "ilya"
	err := errors.New("test")
	detailedErr := detailederror.With(err, key, want)

	got, ok := detailederror.GetDetail(detailedErr, key)

	fmt.Println(ok)
	fmt.Println(got)

	// Output:
	// true
	// ilya
}

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
