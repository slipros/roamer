// Package roamer provides a flexible HTTP request parser.
package roamer

import (
	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// IsDecodeError checks whether an error is a DecodeError from the roamer package.
// This function can be used to specifically handle errors that occurred during
// the decoding of request bodies.
//
// Example:
//
//	if err := roamer.Parse(req, &data); err != nil {
//	    if decodeErr, ok := roamer.IsDecodeError(err); ok {
//	        // Handle decode error specifically
//	        log.Printf("Failed to decode request body: %v", decodeErr)
//	        // Return appropriate HTTP status code
//	        http.Error(w, "Invalid request format", http.StatusBadRequest)
//	        return
//	    }
//	    // Handle other errors
//	    http.Error(w, "Failed to process request", http.StatusInternalServerError)
//	    return
//	}
func IsDecodeError(err error) (rerr.DecodeError, bool) {
	var decodeErr rerr.DecodeError
	return decodeErr, errors.As(err, &decodeErr)
}

// IsSliceIterationError checks whether an error is a SliceIterationError from the roamer package.
// This function can be used to specifically handle errors that occurred during
// the iteration over a slice (e.g., when processing a batch of items).
//
// Example:
//
//	if err := roamer.Parse(req, &items); err != nil {
//	    if iterErr, ok := roamer.IsSliceIterationError(err); ok {
//	        // Handle slice iteration error specifically
//	        log.Printf("Failed to process item at index %d: %v", iterErr.Index, iterErr.Err)
//	        http.Error(w, fmt.Sprintf("Invalid item at position %d", iterErr.Index), http.StatusBadRequest)
//	        return
//	    }
//	    // Handle other errors
//	    http.Error(w, "Failed to process request", http.StatusInternalServerError)
//	    return
//	}
func IsSliceIterationError(err error) (rerr.SliceIterationError, bool) {
	var iterationErr rerr.SliceIterationError
	return iterationErr, errors.As(err, &iterationErr)
}
