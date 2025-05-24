package parser

import (
	"net/http"
	"reflect"
	"strings"
	"unicode"
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

// NewHeader creates a Header parser for extracting HTTP headers from requests.
//
// Example:
//
//	type Request struct {
//	    UserAgent string `header:"User-Agent"`
//	    // Multiple headers with fallback (tries each until finding non-empty value)
//	    ClientIP  string `header:"X-Forwarded-For,X-Real-IP"`
//	}
func NewHeader() *Header {
	return &Header{}
}

// Parse extracts a header from an HTTP request based on the provided struct tag.
// The tag may contain a comma-separated list of header names, and the parser will
// try each header until finding a non-empty value.
//
// Parameters:
//   - r: The HTTP request to extract headers from.
//   - tag: The struct tag containing the header name(s).
//   - _: Cache parameter (not used).
//
// Returns:
//   - any: The header value (string).
//   - bool: Whether a valid header was found.
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

// manyValues handles multiple header names in a comma-separated list.
// It tries each header in sequence until finding a non-empty value.
//
// Example: `header:"X-Forwarded-For,X-Real-IP"`
func (h *Header) manyValues(r *http.Request, tagValue string) (string, bool) {
	for _, headerName := range strings.Split(tagValue, SplitSymbol) {
		if len(headerName) == 0 {
			continue
		}

		if unicode.IsSpace(rune(headerName[0])) || unicode.IsSpace(rune(headerName[len(headerName)-1])) {
			headerName = strings.TrimSpace(headerName)
		}

		headerValue := r.Header.Get(headerName)
		if len(headerValue) == 0 {
			continue
		}

		return headerValue, true
	}

	return "", false
}
