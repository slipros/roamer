package error

import "errors"

var (
	// ErrNoData unable to find parsed data.
	ErrNoData = errors.New("no data")
	// ErrNil nil.
	ErrNil = errors.New("nil")
	// ErrNotPtr not a pointer.
	ErrNotPtr = errors.New("not a ptr")
	// ErrNotSupported type is not supported.
	ErrNotSupported = errors.New("not supported type")
)
