package common

import "errors"

// AssertIsNil panics if the given error is not nil. This should only be used in scenarios where the error is both
// unrecoverable and caused by a developer mistake
func AssertIsNil(err error) {
	if err != nil {
		panic(err)
	}
}

// AssertTrue panics if the given boolean is not true. This should only be used in scenarios where the boolean being
// false indicates a developer mistake.
func AssertTrue(b bool, description string) {
	if b == false {
		panic(errors.New(description))
	}
}
