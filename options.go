package roamer

import (
	"github.com/SLIpros/roamer/decoder"
	"github.com/SLIpros/roamer/parser"
)

// OptionsFunc options.
type OptionsFunc func(*Roamer)

// SetParsers set parsers.
func SetParsers(parsers ...parser.Parser) OptionsFunc {
	return func(r *Roamer) {
		for _, p := range parsers {
			r.parsers[p.Tag()] = p
		}
	}
}

// SetDecoders set decoders.
func SetDecoders(decoders ...decoder.Decoder) OptionsFunc {
	return func(r *Roamer) {
		for _, d := range decoders {
			r.decoders[d.ContentType()] = d
		}
	}
}

// SetSkipFilled set skip filled.
func SetSkipFilled(skip bool) OptionsFunc {
	return func(r *Roamer) {
		r.skipFilled = skip
	}
}
