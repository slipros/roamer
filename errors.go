package roamer

import (
	roamerError "github.com/SLIpros/roamer/err"
	"github.com/pkg/errors"
)

// IsDecodeError checks the error for belonging to decode error.
func IsDecodeError(err error) (*roamerError.DecodeError, bool) {
	var decodeErr *roamerError.DecodeError
	if errors.As(err, &decodeErr) {
		return decodeErr, true
	}

	return nil, false
}
