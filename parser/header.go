package parser

import (
	"net/http"
	"reflect"
	"strings"
)

const (
	// TagHeader header tag.
	TagHeader = "header"
)

// Header is a header parser.
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

	if strings.Contains(tagValue, SplitSymbol) {
		return h.manyValues(r, tagValue)
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
