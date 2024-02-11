package roamer

import (
	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// IsDecodeError checks the error for belonging to decode error.
func IsDecodeError(err error) (rerr.DecodeError, bool) {
	var decodeErr rerr.DecodeError
	return decodeErr, errors.As(err, &decodeErr)
}

// IsSliceIterationError checks the error for belonging to slice iteration error.
func IsSliceIterationError(err error) (rerr.SliceIterationError, bool) {
	var iterationErr rerr.SliceIterationError
	return iterationErr, errors.As(err, &iterationErr)
}
