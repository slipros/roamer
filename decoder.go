// Package roamer provides a flexible HTTP request parser.
package roamer

import "net/http"

// Decoder is an interface for components that decode HTTP request bodies
// based on the Content-Type header. Different decoders can handle different
// formats (JSON, XML, form data, etc.).
//
// Implementing a custom decoder allows extending the functionality of the roamer
// package to support additional content types or custom parsing logic.
//
//go:generate mockery --name=Decoder --outpkg=mock --output=./mock
type Decoder interface {
	// Decode parses the body of an HTTP request into the provided pointer.
	// The implementation should determine how to handle the request body
	// based on the Content-Type header.
	//
	// Parameters:
	//   - r: The HTTP request containing the body to decode.
	//   - ptr: A pointer to the target value where the decoded data will be stored.
	//
	// Returns:
	//   - error: An error if decoding fails, or nil if successful.
	Decode(r *http.Request, ptr any) error

	// ContentType returns the Content-Type header value that this decoder handles.
	// For example, a JSON decoder might return "application/json",
	// an XML decoder might return "application/xml", etc.
	ContentType() string
}

// Decoders is a map of registered decoders where keys are the Content-Type
// header values returned by the Decoder.ContentType() method.
type Decoders map[string]Decoder
