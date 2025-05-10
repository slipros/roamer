// Package err contains error definitions for the roamer package.
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
	// NoData is returned when parsed data cannot be found in the context.
	// This typically occurs when trying to retrieve data using ParsedDataFromContext
	// but no data of the expected type was previously stored in the context.
	NoData = errors.New("no data")

	// NilValue is returned when a nil pointer is provided to a function that expects
	// a valid pointer. This can happen with ParsedDataFromContext or Parse when
	// the destination pointer is nil.
	NilValue = errors.New("value is nil")

	// NotPtr is returned when a non-pointer value is provided to a function that
	// expects a pointer. This typically occurs when calling Parse with a non-pointer
	// destination.
	NotPtr = errors.New("not a ptr")

	// NotSupported is returned when attempting to parse or set a value of a type
	// that is not supported by the roamer package.
	NotSupported = errors.New("not supported type")

	// FieldIndexOutOfBounds is returned when attempting to access a struct field
	// using an index that is out of range. This should generally not occur during
	// normal operation and indicates an internal error.
	FieldIndexOutOfBounds = errors.New("field index out of bounds")
)

// DecodeError represents an error that occurred during the decoding of an HTTP
// request body. It wraps the underlying error that caused the decoding failure.
//
// DecodeError can be detected using the IsDecodeError function in the roamer package:
//
//	if decodeErr, ok := roamer.IsDecodeError(err); ok {
//	    // Handle decode error
//	}
type DecodeError struct {
	// Err is the underlying error that caused the decoding failure.
	// This could be a JSON parsing error, XML parsing error, etc.
	Err error
}

// Error returns a string representation of the decode error.
// This method implements the error interface.
func (d DecodeError) Error() string {
	return d.Err.Error()
}

// SliceIterationError represents an error that occurred while iterating over
// a slice during parsing or processing. It captures both the underlying error
// and the index of the slice element where the error occurred.
//
// SliceIterationError can be detected using the IsSliceIterationError function
// in the roamer package:
//
//	if iterErr, ok := roamer.IsSliceIterationError(err); ok {
//	    // Access iterErr.Index to get the index where the error occurred
//	    // Handle slice iteration error
//	}
type SliceIterationError struct {
	// Err is the underlying error that occurred during slice iteration.
	Err error

	// Index is the position in the slice where the error occurred.
	// This allows pinpointing which element caused the problem.
	Index int
}

// Error returns a string representation of the slice iteration error,
// including the index where the error occurred.
// This method implements the error interface.
func (s SliceIterationError) Error() string {
	return fmt.Sprintf("slice element with index %d: %v", s.Index, s.Err)
}

// FormatterNotFound is returned when a formatter tag references a formatter
// that is not registered. This typically occurs when using a formatter tag
// in a struct field but not registering the corresponding formatter.
type FormatterNotFound struct {
	// Tag is the struct tag name that references the formatter.
	Tag string

	// Formatter is the name of the formatter that could not be found.
	Formatter string
}

// Error returns a string representation of the formatter not found error.
// This method implements the error interface.
func (f FormatterNotFound) Error() string {
	return "formatter '" + f.Formatter + "' not found for tag '" + f.Tag + "'"
}
