package roamer

import "net/http"

// Decoder is a decoder.
//
//go:generate mockery --name=Decoder --outpkg=mock --output=./mock
type Decoder interface {
	Decode(r *http.Request, ptr any) error
	ContentType() string
}

// Decoders is a map of decoders where keys are content types for given decoders.
type Decoders map[string]Decoder
