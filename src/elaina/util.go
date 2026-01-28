package elaina

// AssertIsNil panics if the given error is not nil. This should only be used in scenarios where the error is both
// unrecoverable and caused by a developer mistake
func AssertIsNil(err error) {
	if err != nil {
		panic(err)
	}
}
