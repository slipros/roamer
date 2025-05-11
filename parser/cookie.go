// Package parser provides parsers for extracting data from HTTP requests.
package parser

import (
	"net/http"
	"reflect"
)

const (
	// TagCookie is the struct tag name used for parsing HTTP cookies.
	// Fields tagged with this will be populated from matching HTTP cookies.
	// Example: `cookie:"session_id"`
	TagCookie = "cookie"
)

// Cookie is a parser for extracting cookies from HTTP requests.
// It allows easy access to cookie values using struct tags.
type Cookie struct{}

// NewCookie creates a new Cookie parser.
//
// Example:
//
//	// Create a cookie parser
//	cookieParser := parser.NewCookie()
//
//	// Use it with roamer
//	r := roamer.NewRoamer(
//	    roamer.WithParsers(cookieParser),
//	)
//
//	// Example struct using the parser
//	type Request struct {
//	    SessionID string `cookie:"session_id"`
//	    UserTheme string `cookie:"theme"`
//	}
func NewCookie() *Cookie {
	return &Cookie{}
}

// Parse extracts a cookie from an HTTP request based on the provided struct tag.
// If the cookie exists, it returns the *http.Cookie object and true.
// If the cookie does not exist, it returns an empty string and false.
//
// Note that this parser returns the entire *http.Cookie object, not just the value.
// This allows access to other cookie properties like Expires, MaxAge, etc.
// To access just the cookie value, use the Value field of the returned object.
//
// Parameters:
//   - r: The HTTP request to extract cookies from.
//   - tag: The struct tag containing the cookie name.
//   - _: Cache parameter (not used by this parser).
//
// Returns:
//   - any: The parsed cookie (*http.Cookie).
//   - bool: Whether the cookie was found.
func (c *Cookie) Parse(r *http.Request, tag reflect.StructTag, _ Cache) (any, bool) {
	tagValue, ok := tag.Lookup(TagCookie)
	if !ok {
		return "", false
	}

	v, err := r.Cookie(tagValue)
	if err != nil {
		return "", false
	}

	return v, true
}

// Tag returns the name of the struct tag that this parser handles.
// For the Cookie parser, this is "cookie".
func (c *Cookie) Tag() string {
	return TagCookie
}
