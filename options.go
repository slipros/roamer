// Package roamer provides a flexible HTTP request parser.
package roamer

// OptionsFunc is a function type for configuring a Roamer instance.
// It follows the functional options pattern to provide a clean and
// extensible API for customizing the behavior of the parser.
type OptionsFunc func(*Roamer)

// WithParsers registers one or more parsers with a Roamer instance.
// Parsers are responsible for extracting data from different parts of an HTTP request
// (headers, query parameters, cookies, etc.) based on struct tags.
//
// Example:
//
//	r := roamer.NewRoamer(
//	    roamer.WithParsers(
//	        parser.NewQuery(),    // Parse query parameters using the 'query' tag
//	        parser.NewHeader(),   // Parse headers using the 'header' tag
//	        parser.NewCookie(),   // Parse cookies using the 'cookie' tag
//	        parser.NewPath(),     // Parse path parameters using the 'path' tag
//	    ),
//	)
func WithParsers(parsers ...Parser) OptionsFunc {
	return func(r *Roamer) {
		for _, p := range parsers {
			r.parsers[p.Tag()] = p
		}
	}
}

// WithDecoders registers one or more decoders with a Roamer instance.
// Decoders are responsible for parsing the HTTP request body based on
// the Content-Type header (e.g., JSON, XML, form data).
//
// Example:
//
//	r := roamer.NewRoamer(
//	    roamer.WithDecoders(
//	        decoder.NewJSON(),            // Parse JSON request bodies
//	        decoder.NewFormURL(),         // Parse URL-encoded form data
//	        decoder.NewMultipartFormData(), // Parse multipart form data (with file uploads)
//	    ),
//	)
func WithDecoders(decoders ...Decoder) OptionsFunc {
	return func(r *Roamer) {
		for _, d := range decoders {
			r.decoders[d.ContentType()] = d
		}
	}
}

// WithFormatters registers one or more formatters with a Roamer instance.
// Formatters are used to process field values after they have been parsed,
// allowing for operations like string trimming, case conversion, etc.
//
// Example:
//
//	r := roamer.NewRoamer(
//	    roamer.WithFormatters(
//	        formatter.NewString(), // Apply string formatters based on the 'string' tag
//	    ),
//	)
//
//	// Usage in a struct:
//	type User struct {
//	    Name string `json:"name" string:"trim_space"` // Trim spaces from the parsed name
//	}
func WithFormatters(formatters ...Formatter) OptionsFunc {
	return func(r *Roamer) {
		for _, f := range formatters {
			r.formatters[f.Tag()] = f
		}
	}
}

// WithSkipFilled controls whether the parser should skip fields that already
// have non-zero values. When set to true (default), fields that are already
// filled will not be overwritten by parsed values.
//
// Example:
//
//	// Don't skip filled fields (overwrite all fields with parsed values)
//	r := roamer.NewRoamer(
//	    roamer.WithSkipFilled(false),
//	)
//
//	// Pre-fill some fields and let the parser fill the rest
//	user := User{Name: "Default Name"}
//	r.Parse(req, &user) // Will not overwrite the Name field if WithSkipFilled(true)
func WithSkipFilled(skip bool) OptionsFunc {
	return func(r *Roamer) {
		r.skipFilled = skip
	}
}

// WithExperimentalFastStructFieldParser enables the use of an experimental fast struct field parser.
// This can improve performance but may not be as stable as the standard parser.
//
// Warning: This is an experimental feature and may change or be removed in future versions.
//
// Example:
//
//	// Enable experimental fast struct field parser
//	r := roamer.NewRoamer(
//	    roamer.WithExperimentalFastStructFieldParser(),
//	)
func WithExperimentalFastStructFieldParser() OptionsFunc {
	return func(r *Roamer) {
		r.experimentalFastStructField = true
	}
}
