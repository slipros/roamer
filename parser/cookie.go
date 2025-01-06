package parser

import (
	"net/http"
	"reflect"
)

const (
	// TagCookie cookie tag.
	TagCookie = "cookie"
)

// Cookie is a cookie parser.
type Cookie struct{}

// NewCookie returns new cookie parser.
func NewCookie() *Cookie {
	return &Cookie{}
}

// Parse parse cookie.
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

// Tag returns working tag.
func (c *Cookie) Tag() string {
	return TagCookie
}
