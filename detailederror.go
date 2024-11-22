// # detailederror
//
// error handling improvement we all wanted.
//
// # Idea
//
// Make a tool similar to [context.Context], but passing values upstream the
// call stack. With API similar to this:
//
//	ctx := context.WithValue(context.Background(), k, "Go")
//
// This might come in handy when we want to get rid of error messages like this:
//
//	"error fetch user by id 13: error getting user from db by id 13: error user not found: blah-blah-blah".
//
// Using this library you can add all details necessary (id, username & etc)
// and get them later without affecting message itself & removing data duplication.
// The message may be:
//
//	"error fetch user: error getting user from db: blah-blah-blah. id:13,status=not found".
//
// I guess error above is much more easier to read. Moreover you can use [slog]
// package to format your message structurally.
package detailederror

import "errors"

type detailedError struct {
	err error
	key string
	val string
}

func (d *detailedError) Error() string {
	return d.err.Error()
}

// WithMany wraps error with multiple values at once.
// Incomplete pairs at the end of the pairs parameter are omitted:
//
//	detailederror.WithMany(err, "key-present", "value", "key-absent")
//
// You might want to see [With].
func WithMany(err error, pairs ...string) error {
	derr := err
	i := 0
	for i+1 < len(pairs) {
		derr = With(derr, pairs[i], pairs[i+1])
		i += 2
	}
	return derr
}

// With attaches metadata to [error] so we can extract it later using [GetDetail].
func With(err error, key, value string) error {
	return &detailedError{
		err: err,
		key: key,
		val: value,
	}
}

// GetDetail extract value attached using [With].
// Useful at places where we handle an error and want to create good logging
// record.
func GetDetail(err error, key string) (string, bool) {
	nextErr := err
	var derr *detailedError
	for errors.As(nextErr, &derr) {
		if derr.key == key {
			return derr.val, true
		}
		nextErr = derr.err
	}
	return "", false
}

// GetDetails extracts all details from the [error].
// Entries with the same key are overridden and wont show up in resulting map.
func GetDetails(err error) map[string]string {
	nextErr := err
	var derr *detailedError
	details := make(map[string]string, 0)
	for errors.As(nextErr, &derr) {
		details[derr.key] = derr.val
		nextErr = derr.err
	}
	return details
}
