// Package err contains roamer errors.
//
// Since this package is dedicated to errors and the package is named "err",
// all errors elide the standard "Err" prefix.
//
//nolint:revive,errname,stylecheck
package err

import (
	"errors"
	"fmt"
)

var (
	// NoData unable to find parsed data.
	NoData = errors.New("no data")
	// NilValue value is nil.
	NilValue = errors.New("value is nil")
	// NotPtr not a pointer.
	NotPtr = errors.New("not a ptr")
	// NotSupported type is not supported.
	NotSupported = errors.New("not supported type")
	// FieldIndexOutOfBounds field index out of bounds.
	FieldIndexOutOfBounds = errors.New("field index out of bounds")
)

// DecodeError decode error.
type DecodeError struct {
	Err error
}

// Error returns string.
func (d DecodeError) Error() string {
	return d.Err.Error()
}

// SliceIterationError slice iteration error.
type SliceIterationError struct {
	Err   error
	Index int
}

// Error returns string.
func (s SliceIterationError) Error() string {
	return fmt.Sprintf("slice element with index %d: %v", s.Index, s.Err)
}

// FormatterNotFound not found formatter error.
type FormatterNotFound struct {
	Tag       string
	Formatter string
}

// Error returns string.
func (f FormatterNotFound) Error() string {
	return "formatter '" + f.Formatter + "' not found for tag '" + f.Tag + "'"
}
