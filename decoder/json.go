package decoder

import (
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

const (
	// ContentTypeJSON is the Content-Type header value for JSON requests.
	// This is used to match requests with the appropriate decoder.
	ContentTypeJSON = "application/json"
)

// json is a jsoniter instance configured to be compatible with the standard library.
// This provides better performance while maintaining compatibility with encoding/json.
var json = jsoniter.ConfigCompatibleWithStandardLibrary

// JSONOptionsFunc is a function type for configuring a JSON decoder.
// It follows the functional options pattern to provide a clean and extensible API.
type JSONOptionsFunc = func(*JSON)

// JSON is a decoder for handling JSON request bodies.
// It uses the jsoniter library for better performance compared to the standard library.
type JSON struct {
	contentType string // The Content-Type header value that this decoder handles
}

// NewJSON creates a JSON decoder for handling application/json content type.
// Uses jsoniter for improved performance over standard library.
//
// Example:
//
//	// Basic JSON decoder
//	jsonDecoder := decoder.NewJSON()
//
//	// Use with roamer
//	r := roamer.NewRoamer(roamer.WithDecoders(jsonDecoder))
//
//	// Example struct definition
//	type UserRequest struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
func NewJSON(opts ...JSONOptionsFunc) *JSON {
	j := JSON{
		contentType: ContentTypeJSON,
	}

	for _, opt := range opts {
		opt(&j)
	}

	return &j
}

// Decode parses a JSON request body into the provided pointer.
// It uses the jsoniter library for high-performance JSON parsing.
//
// The function handles empty bodies gracefully (treating them as an empty JSON object).
// For other parsing errors, the original error is returned.
//
// Parameters:
//   - r: The HTTP request containing the JSON body to decode.
//   - ptr: A pointer to the target structure where the decoded data will be stored.
//
// Returns:
//   - error: An error if decoding fails, or nil if successful.
func (j *JSON) Decode(r *http.Request, ptr any) error {
	if err := json.NewDecoder(r.Body).Decode(ptr); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}

	return nil
}

// ContentType returns the Content-Type header value that this decoder handles.
// For the JSON decoder, this is "application/json" by default.
// This method is used by the roamer package to match requests with the appropriate decoder.
func (j *JSON) ContentType() string {
	return j.contentType
}

// setContentType sets the Content-Type header value that this decoder handles.
// This is primarily used internally by option functions.
func (j *JSON) setContentType(contentType string) {
	j.contentType = contentType
}
