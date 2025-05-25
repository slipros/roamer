// Package err contains error definitions for the roamer package.
//
// Since this package is dedicated to errors and the package is named "err",
// all errors elide the standard "Err" prefix.
//
//nolint:staticcheck
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
)

// DecodeError represents a failure during HTTP request body decoding.
// Wraps the underlying error from JSON, XML, or other format parsing.
//
// Can be detected with roamer.IsDecodeError():
//
//	if decodeErr, ok := roamer.IsDecodeError(err); ok {
//	    // Handle body parsing error
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

// Unwrap returns the underlying error that caused the decoding failure.
// This method implements the interface for errors.Unwrap(), allowing the use
// of errors.Is() and errors.As() functions to examine the error chain.
func (d DecodeError) Unwrap() error {
	return d.Err
}

// SliceIterationError occurs when processing a slice element fails.
// Contains both the underlying error and the index where it occurred.
//
// Can be detected with roamer.IsSliceIterationError():
//
//	if iterErr, ok := roamer.IsSliceIterationError(err); ok {
//	    fmt.Printf("Error at element %d: %v", iterErr.Index, iterErr.Err)
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

// Unwrap returns the underlying error that occurred during slice iteration.
// This method implements the interface for errors.Unwrap(), allowing the use
// of errors.Is() and errors.As() functions to examine the error chain.
// This is useful for checking what specific type of error occurred while
// still maintaining the context of which slice element caused it.
func (s SliceIterationError) Unwrap() error {
	return s.Err
}

// FormatterNotFound occurs when a tag references a non-existent formatter.
// This happens when a struct uses a formatter that hasn't been registered.
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
