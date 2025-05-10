// Package decoder provides decoders for extracting data from HTTP request bodies.
package decoder

import (
	"encoding/xml"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const (
	// ContentTypeXML is the Content-Type header value for XML requests.
	// This is used to match requests with the appropriate decoder.
	ContentTypeXML = "application/xml"
)

// XMLOptionsFunc is a function type for configuring an XML decoder.
// It follows the functional options pattern to provide a clean and extensible API.
type XMLOptionsFunc = func(*XML)

// XML is a decoder for handling XML request bodies.
// It uses the standard library's encoding/xml package for XML parsing.
type XML struct {
	contentType string // The Content-Type header value that this decoder handles
}

// NewXML creates a new XML decoder with the specified options.
// By default, it handles requests with Content-Type "application/xml".
//
// Example:
//
//	// Create an XML decoder with default settings
//	xmlDecoder := decoder.NewXML()
//
//	// Create an XML decoder with custom Content-Type
//	xmlDecoder := decoder.NewXML(
//	    decoder.WithContentType("text/xml"),
//	)
//
//	// Use it with roamer
//	r := roamer.NewRoamer(
//	    roamer.WithDecoders(xmlDecoder),
//	)
//
//	// Example struct using XML tags
//	type BookRequest struct {
//	    Title  string `xml:"title"`
//	    Author string `xml:"author"`
//	    Year   int    `xml:"year"`
//	}
func NewXML(opts ...XMLOptionsFunc) *XML {
	x := XML{
		contentType: ContentTypeXML,
	}

	for _, opt := range opts {
		opt(&x)
	}

	return &x
}

// Decode parses an XML request body into the provided pointer.
// It uses the standard library's encoding/xml package for XML parsing.
//
// The function handles empty bodies gracefully (treating them as an empty XML document).
// For other parsing errors, the original error is returned.
//
// Parameters:
//   - r: The HTTP request containing the XML body to decode.
//   - ptr: A pointer to the target structure where the decoded data will be stored.
//
// Returns:
//   - error: An error if decoding fails, or nil if successful.
func (x *XML) Decode(r *http.Request, ptr any) error {
	if err := xml.NewDecoder(r.Body).Decode(ptr); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}

	return nil
}

// ContentType returns the Content-Type header value that this decoder handles.
// For the XML decoder, this is "application/xml" by default.
// This method is used by the roamer package to match requests with the appropriate decoder.
func (x *XML) ContentType() string {
	return x.contentType
}

// setContentType sets the Content-Type header value that this decoder handles.
// This is primarily used internally by option functions.
func (x *XML) setContentType(contentType string) {
	x.contentType = contentType
}
