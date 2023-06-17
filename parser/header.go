package parser

import (
	"net/http"
	"reflect"
)

const (
	// TagHeader header tag.
	TagHeader = "header"
)

// Header header parser.
type Header struct{}

// NewHeader returns new header parser.
func NewHeader() *Header {
	return &Header{}
}

// Parse parse header.
func (h *Header) Parse(r *http.Request, tag reflect.StructTag, _ Cache) (any, bool) {
	tagValue, ok := tag.Lookup(TagHeader)
	if !ok {
		return "", false
	}

	headerValue := r.Header.Get(tagValue)
	if len(headerValue) == 0 {
		return "", false
	}

	return headerValue, true
}

// Tag returns working tag.
func (h *Header) Tag() string {
	return TagHeader
}
