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

	// TagJSON is the struct tag name used for JSON field mapping.
	TagJSON = "json"
)

// JSONOptionsFunc is a function type for configuring a JSON decoder.
// It follows the functional options pattern to provide a clean and extensible API.
type JSONOptionsFunc = func(*JSON)

// JSON is a decoder for handling JSON request bodies.
//
// It uses the jsoniter library (github.com/json-iterator/go) for better performance
// compared to the standard library's encoding/json. The decoder is configured to be
// compatible with the standard library's behavior by default.
//
// # Performance
//
// jsoniter provides significant performance improvements over encoding/json:
//   - Faster encoding and decoding
//   - Lower memory allocations
//   - API-compatible with standard library
//
// # Thread Safety
//
// The JSON decoder is safe for concurrent use across multiple goroutines.
type JSON struct {
	decoder     jsoniter.API
	contentType string // The Content-Type header value that this decoder handles
}

// NewJSON creates a JSON decoder for handling application/json content type.
//
// By default, the decoder uses jsoniter with standard library compatible configuration,
// which provides better performance while maintaining the same behavior as encoding/json.
//
// # Configuration
//
// The decoder can be customized using functional options:
//   - WithContentType: Override the default "application/json" content type
//
// # Default Behavior
//
//   - Content-Type: application/json
//   - Tag name: json
//   - Configuration: Compatible with standard library
//   - Empty body handling: Treated as empty object/value
//
// Parameters:
//   - opts: Optional configuration functions to customize the decoder.
//
// Returns:
//   - *JSON: A configured JSON decoder instance.
//
// Example:
//
//	// Basic JSON decoder
//	jsonDecoder := decoder.NewJSON()
//
//	// JSON decoder with custom content type
//	jsonDecoder := decoder.NewJSON(
//	    decoder.WithContentType("application/x-json"),
//	)
//
//	// Use with roamer
//	r := roamer.NewRoamer(roamer.WithDecoders(jsonDecoder))
//
//	// Example struct
//	type UserRequest struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	    Age   int    `json:"age,omitempty"`
//	}
func NewJSON(opts ...JSONOptionsFunc) *JSON {
	j := JSON{
		decoder:     jsoniter.ConfigCompatibleWithStandardLibrary,
		contentType: ContentTypeJSON,
	}

	for _, opt := range opts {
		opt(&j)
	}

	return &j
}

// Decode parses a JSON request body into the provided pointer.
//
// The method reads the request body from r.Body and decodes it into the destination
// pointed to by ptr using jsoniter for high-performance parsing. The decoder is
// compatible with standard library's encoding/json behavior.
//
// # Empty Body Handling
//
// If the request body is empty (io.EOF), the method returns nil without error,
// leaving the destination in its zero state. This allows graceful handling of
// requests with no body content.
//
// # Error Handling
//
// The method returns an error if:
//   - JSON syntax is invalid
//   - JSON structure doesn't match the destination type
//   - Required fields are missing (when using json:",required" tag)
//   - Type conversion fails
//
// Errors from jsoniter are returned directly and can be examined for details.
//
// # Supported Types
//
//   - Structs with json tags
//   - Maps (map[string]any, map[string]string, etc.)
//   - Slices and arrays
//   - Basic types (when ptr is *string, *int, etc.)
//
// Parameters:
//   - r: The HTTP request containing the JSON body. Body will be read but not closed.
//   - ptr: A pointer to the destination where decoded data will be stored. Must not be nil.
//
// Returns:
//   - error: An error if decoding fails, or nil if successful.
//
// Example:
//
//	type UserRequest struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	    Age   int    `json:"age"`
//	}
//
//	var user UserRequest
//	err := jsonDecoder.Decode(r, &user)
//	if err != nil {
//	    // Handle JSON parsing error
//	    return fmt.Errorf("invalid JSON: %w", err)
//	}
func (j *JSON) Decode(r *http.Request, ptr any) error {
	if err := j.decoder.NewDecoder(r.Body).Decode(ptr); err != nil {
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

// Tag returns the struct tag name used for JSON field mapping.
// For the JSON decoder, this is "json" by default.
func (j *JSON) Tag() string {
	return TagJSON
}

// setContentType sets the Content-Type header value that this decoder handles.
// This is primarily used internally by option functions.
func (j *JSON) setContentType(contentType string) {
	j.contentType = contentType
}
