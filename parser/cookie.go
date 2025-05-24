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

// NewCookie creates a Cookie parser for extracting cookie values from HTTP requests.
//
// Example:
//
//	type Request struct {
//	    SessionID string `cookie:"session_id"`
//	    UserTheme string `cookie:"theme"`
//	}
func NewCookie() *Cookie {
	return &Cookie{}
}

// Parse extracts a cookie from an HTTP request based on the struct tag.
// Returns the *http.Cookie object (not just the value) when found, allowing
// access to properties like Expires, MaxAge, etc.
//
// Parameters:
//   - r: The HTTP request to extract cookies from.
//   - tag: The struct tag containing the cookie name.
//   - _: Cache parameter (not used).
//
// Returns:
//   - any: The cookie (*http.Cookie) or empty string if not found.
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
