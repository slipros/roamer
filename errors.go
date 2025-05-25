package roamer

import (
	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// IsDecodeError checks if an error is a DecodeError from request body parsing.
// Useful for providing specific error handling for body decoding failures.
//
// Example:
//
//	if err := roamer.Parse(req, &data); err != nil {
//	    if decodeErr, ok := roamer.IsDecodeError(err); ok {
//	        // Special handling for decode errors
//	        http.Error(w, "Invalid request format", http.StatusBadRequest)
//	        return
//	    }
//	    // Handle other errors
//	}
func IsDecodeError(err error) (rerr.DecodeError, bool) {
	var decodeErr rerr.DecodeError
	return decodeErr, errors.As(err, &decodeErr)
}

// IsSliceIterationError checks if an error occurred during slice iteration.
// Provides access to the specific index where the error occurred.
//
// Example:
//
//	if err := processItems(items); err != nil {
//	    if iterErr, ok := roamer.IsSliceIterationError(err); ok {
//	        // Access problem index with iterErr.Index
//	        return fmt.Errorf("item %d invalid: %w", iterErr.Index, iterErr.Err)
//	    }
//	}
func IsSliceIterationError(err error) (rerr.SliceIterationError, bool) {
	var iterationErr rerr.SliceIterationError
	return iterationErr, errors.As(err, &iterationErr)
}
