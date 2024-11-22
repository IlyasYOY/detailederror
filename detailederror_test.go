package detailederror_test

import (
	"errors"
	"testing"

	"github.com/IlyasYOY/detailederror"
	"github.com/google/go-cmp/cmp"
)

func TestWithMany(t *testing.T) {
	err := errors.New("test")

	gotErr := detailederror.WithMany(err, "user", "ilya", "level", "info")

	if gotErr == nil {
		t.Errorf("error must not be nil: %v", gotErr)
	}
}

func TestWith(t *testing.T) {
	err := errors.New("test")

	gotErr := detailederror.With(err, "user", "ilya")

	if gotErr == nil {
		t.Errorf("error must not be nil: %v", gotErr)
	}
}

func TestGetDetails_FromWithManyWithAbsentKeyPair(t *testing.T) {
	err := errors.New("test")
	detailedErr := detailederror.WithMany(
		err,
		"user1", "ilya1",
		"user2",
	)

	got := detailederror.GetDetails(detailedErr)

	if diff := cmp.Diff(map[string]string{
		"user1": "ilya1",
	}, got); diff != "" {
		t.Errorf("(-want +got):\\n%s", diff)
	}
}

func TestGetDetails_FromWithMany(t *testing.T) {
	err := errors.New("test")
	detailedErr := detailederror.WithMany(
		err,
		"user1", "ilya1",
		"user2", "ilya2",
	)

	got := detailederror.GetDetails(detailedErr)

	if diff := cmp.Diff(map[string]string{
		"user1": "ilya1",
		"user2": "ilya2",
	}, got); diff != "" {
		t.Errorf("(-want +got):\\n%s", diff)
	}
}

func TestGetDetail_Nested(t *testing.T) {
	err := errors.New("test")
	key1 := "user1"
	want1 := "ilya1"
	detailedErr1 := detailederror.With(err, key1, want1)
	key2 := "user2"
	want2 := "ilya2"
	detailedErr2 := detailederror.With(detailedErr1, key2, want2)

	got2, ok2 := detailederror.GetDetail(detailedErr2, key2)
	got1, ok1 := detailederror.GetDetail(detailedErr2, key1)

	if !ok1 {
		t.Fatalf("value not found by key: %q", key1)
	}
	if !ok2 {
		t.Fatalf("value not found by key: %q", key2)
	}
	if got2 != want2 {
		t.Errorf("value must be equal to %q", want2)
	}
	if got1 != want1 {
		t.Errorf("value must be equal to %q", want1)
	}
}

func TestGetDetails_Nested(t *testing.T) {
	err := errors.New("test")
	detailedErr1 := detailederror.With(err, "user1", "ilya1")
	detailedErr2 := detailederror.With(detailedErr1, "user2", "ilya2")

	got := detailederror.GetDetails(detailedErr2)

	if diff := cmp.Diff(map[string]string{
		"user1": "ilya1",
		"user2": "ilya2",
	}, got); diff != "" {
		t.Errorf("(-want +got):\\n%s", diff)
	}
}

func TestGetDetail(t *testing.T) {
	err := errors.New("test")
	key := "user"
	want := "ilya"
	detailedErr := detailederror.With(err, key, want)

	got, ok := detailederror.GetDetail(detailedErr, key)

	if !ok {
		t.Fatalf("value not found by key: %q", key)
	}
	if got != want {
		t.Errorf("value must be equal to %q", want)
	}
}

func TestGetDetail_Absent(t *testing.T) {
	err := errors.New("test")
	want := "ilya"
	detailedErr := detailederror.With(err, "name", want)

	_, ok := detailederror.GetDetail(detailedErr, "user")

	if ok {
		t.Fatalf("value was found")
	}
}
