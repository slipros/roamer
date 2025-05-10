// Package parser provides parsers for extracting data from HTTP requests.
package parser

import (
	"net/http"
	"reflect"
	"strings"
)

const (
	// TagHeader is the struct tag name used for parsing HTTP headers.
	// Fields tagged with this will be populated from matching HTTP headers.
	// Example: `header:"User-Agent"`
	TagHeader = "header"
)

// Header is a parser for extracting HTTP headers from requests.
// It supports both single header extraction and fallback to alternative headers
// if the primary header is not present.
type Header struct{}

// NewHeader creates a new Header parser.
//
// Example:
//
//	// Create a header parser
//	headerParser := parser.NewHeader()
//
//	// Use it with roamer
//	r := roamer.NewRoamer(
//	    roamer.WithParsers(headerParser),
//	)
//
//	// Example struct using the parser
//	type Request struct {
//	    UserAgent string `header:"User-Agent"`
//	    Accept    string `header:"Accept"`
//	    // Multiple headers with fallback (tries each one until it finds a non-empty value)
//	    ClientIP  string `header:"X-Forwarded-For,X-Real-IP,CF-Connecting-IP"`
//	}
func NewHeader() *Header {
	return &Header{}
}

// Parse extracts a header from an HTTP request based on the provided struct tag.
// If the header exists, it returns the value and true.
// If the header does not exist, it returns an empty string and false.
//
// The tag may contain a comma-separated list of header names to try.
// In this case, the parser will try each header in order until it finds
// a non-empty value.
//
// Parameters:
//   - r: The HTTP request to extract headers from.
//   - tag: The struct tag containing the header name(s).
//   - _: Cache parameter (not used by this parser).
//
// Returns:
//   - any: The parsed header value (string).
//   - bool: Whether a valid header value was found.
func (h *Header) Parse(r *http.Request, tag reflect.StructTag, _ Cache) (any, bool) {
	tagValue, ok := tag.Lookup(TagHeader)
	if !ok {
		return "", false
	}

	if strings.Contains(tagValue, SplitSymbol) {
		return h.manyValues(r, tagValue)
	}

	headerValue := r.Header.Get(tagValue)
	if len(headerValue) == 0 {
		return "", false
	}

	return headerValue, true
}

// Tag returns the name of the struct tag that this parser handles.
// For the Header parser, this is "header".
func (h *Header) Tag() string {
	return TagHeader
}

// manyValues handles the case where multiple header names are provided
// in a comma-separated list. It tries each header in sequence until it
// finds a non-empty value.
//
// Example: `header:"X-Forwarded-For,X-Real-IP,CF-Connecting-IP"`
// This will try X-Forwarded-For first, then X-Real-IP, then CF-Connecting-IP,
// and return the first non-empty value.
func (h *Header) manyValues(r *http.Request, tagValue string) (string, bool) {
	for _, v := range strings.Split(tagValue, SplitSymbol) {
		headerValue := r.Header.Get(strings.TrimSpace(v))
		if len(headerValue) == 0 {
			continue
		}

		return headerValue, true
	}

	return "", false
}
