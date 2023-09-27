package roamer

import (
	rerr "github.com/SLIpros/roamer/err"
	"github.com/pkg/errors"
)

// IsDecodeError checks the error for belonging to decode error.
func IsDecodeError(err error) (*rerr.DecodeError, bool) {
	var decodeErr *rerr.DecodeError
	if errors.As(err, &decodeErr) {
		return decodeErr, true
	}

	return nil, false
}

// IsSliceIterationError checks the error for belonging to slice iteration error.
func IsSliceIterationError(err error) (*rerr.SliceIterationError, bool) {
	var iterationErr *rerr.SliceIterationError
	if errors.As(err, &iterationErr) {
		return iterationErr, true
	}

	return nil, false
}
