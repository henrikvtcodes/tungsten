package util

import "errors"

var (
	RecursionStubError = errors.New("recursion not enabled in this binary")
)
