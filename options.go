package roamer

// OptionsFunc options func.
type OptionsFunc func(*Roamer)

// SetParsers sets parsers.
func SetParsers(parsers ...Parser) OptionsFunc {
	return func(r *Roamer) {
		for _, p := range parsers {
			r.parsers[p.Tag()] = p
		}
	}
}

// SetDecoders sets decoders.
func SetDecoders(decoders ...Decoder) OptionsFunc {
	return func(r *Roamer) {
		for _, d := range decoders {
			r.decoders[d.ContentType()] = d
		}
	}
}

// SetSkipFilled sets skip filled.
func SetSkipFilled(skip bool) OptionsFunc {
	return func(r *Roamer) {
		r.skipFilled = skip
	}
}
