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

// Decoders decoders.
type Decoders map[string]Decoder
