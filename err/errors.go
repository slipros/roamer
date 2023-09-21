// Package err contains roamer errors.
//
// Since this package is dedicated to errors and the package is named "err",
// all errors elide the standard "Err" prefix.
//
//nolint:revive,errname,stylecheck
package err

import "errors"

var (
	// NoData unable to find parsed data.
	NoData = errors.New("no data")
	// NilValue value is nil.
	NilValue = errors.New("value is nil")
	// NotPtr not a pointer.
	NotPtr = errors.New("not a ptr")
	// NotSupported type is not supported.
	NotSupported = errors.New("not supported type")
)

// DecodeError decode error.
type DecodeError struct {
	Err error
}

// Error returns string.
func (d *DecodeError) Error() string {
	return d.Err.Error()
}
