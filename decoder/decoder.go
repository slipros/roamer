// Package decoder decode content-type header and return it's values.
package decoder

import (
	"net/http"
)

// Decoder decoder.
//
//go:generate mockery --name=Decoder --outpkg=mockdecoder --output=./mockdecoder
type Decoder interface {
	ContentType() string
	Decode(r *http.Request, ptr any) error
}

// Decoders is a map of decoders where keys are content types for given decoders.
type Decoders map[string]Decoder
