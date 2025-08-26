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

	TagXML = "xml"
)

// XMLOptionsFunc is a function type for configuring an XML decoder.
// It follows the functional options pattern to provide a clean and extensible API.
type XMLOptionsFunc = func(*XML)

// XML is a decoder for handling XML request bodies.
// It uses the standard library's encoding/xml package for XML parsing.
type XML struct {
	contentType string // The Content-Type header value that this decoder handles
}

// NewXML creates an XML decoder for handling application/xml content.
// Uses standard library's encoding/xml package for XML parsing.
//
// Example:
//
//	// Default XML decoder
//	xmlDecoder := decoder.NewXML()
//
//	// With custom Content-Type
//	xmlDecoder := decoder.NewXML(decoder.WithContentType("text/xml"))
//
//	// Example struct
//	type BookRequest struct {
//	    Title  string `xml:"title"`
//	    Author string `xml:"author"`
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
// Handles empty bodies gracefully as empty XML documents.
//
// Parameters:
//   - r: The HTTP request with XML body.
//   - ptr: Target structure pointer.
//
// Returns:
//   - error: Error if decoding fails, nil if successful.
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

// Tag returns the struct tag name used for XML field mapping.
// For the XML decoder, this is "xml" by default.
func (x *XML) Tag() string {
	return TagXML
}

// setContentType sets the Content-Type header value that this decoder handles.
// This is primarily used internally by option functions.
func (x *XML) setContentType(contentType string) {
	x.contentType = contentType
}
