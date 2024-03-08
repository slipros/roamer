package roamer

// OptionsFunc function for setting options.
type OptionsFunc func(*Roamer)

// WithParsers sets parsers.
func WithParsers(parsers ...Parser) OptionsFunc {
	return func(r *Roamer) {
		for _, p := range parsers {
			r.parsers[p.Tag()] = p
		}
	}
}

// WithDecoders sets decoders.
func WithDecoders(decoders ...Decoder) OptionsFunc {
	return func(r *Roamer) {
		for _, d := range decoders {
			r.decoders[d.ContentType()] = d
		}
	}
}

// WithSkipFilled sets skip filled.
func WithSkipFilled(skip bool) OptionsFunc {
	return func(r *Roamer) {
		r.skipFilled = skip
	}
}

// WithExperimentalFastStructFieldParser enables the use of experimental fast struct field parser.
func WithExperimentalFastStructFieldParser() OptionsFunc {
	return func(r *Roamer) {
		r.experimentalFastStructField = true
	}
}
