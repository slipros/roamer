package roamer_test

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/err"
)

// readCloser wraps a bytes.Buffer to implement io.ReadCloser
type readCloser struct {
	*bytes.Buffer
}

func (rc *readCloser) Close() error {
	return nil
}

// ExampleIsDecodeError demonstrates how to detect and handle decode errors.
func ExampleIsDecodeError() {
	// Define a structure
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	// Create roamer with JSON decoder
	r := roamer.NewRoamer(
		roamer.WithDecoders(decoder.NewJSON()),
	)

	// Create request with invalid JSON body (malformed JSON)
	invalidJSON := `{invalid}` // Malformed JSON to cause parsing error
	req := &http.Request{
		Method: "POST",
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		Body:          &readCloser{bytes.NewBufferString(invalidJSON)},
		ContentLength: int64(len(invalidJSON)),
	}

	var user User
	err := r.Parse(req, &user)
	if err != nil {
		// Check if it's a decode error
		if decodeErr, isDecodeError := roamer.IsDecodeError(err); isDecodeError {
			fmt.Printf("Decode error occurred: %v\n", decodeErr)
			// Handle decode error specifically
			return
		}

		// Handle other types of errors
		fmt.Printf("Other error: %v\n", err)
		return
	}

	fmt.Printf("User parsed successfully: %+v\n", user)

	// Output:
	// Decode error occurred: decode `application/json` request body for `*roamer_test.User`: roamer_test.User.readFieldHash: expect ", but found i, error found in #2 byte of ...|{invalid}|..., bigger context ...|{invalid}|...
}

// ExampleIsSliceIterationError demonstrates detecting slice iteration errors.
func ExampleIsSliceIterationError() {
	// This example simulates a scenario where slice iteration might fail
	// In practice, this would occur during complex slice processing operations

	// Create a mock slice iteration error for demonstration
	originalErr := fmt.Errorf("validation failed")
	sliceErr := err.SliceIterationError{
		Err:   originalErr,
		Index: 2,
	}

	// Simulate checking the error
	var testErr error = sliceErr

	if iterErr, isSliceError := roamer.IsSliceIterationError(testErr); isSliceError {
		fmt.Printf("Slice iteration error at index %d: %v\n", iterErr.Index, iterErr.Err)
	} else {
		fmt.Printf("Not a slice iteration error: %v\n", testErr)
	}

	// Output:
	// Slice iteration error at index 2: validation failed
}
