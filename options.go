package roamer

import (
	"github.com/slipros/assign"
)

// OptionsFunc is a function type for configuring a Roamer instance.
// It follows the functional options pattern to provide a clean and
// extensible API for customizing the behavior of the parser.
type OptionsFunc func(*Roamer)

// WithParsers registers parsers that extract data from HTTP requests.
// Parsers handle different parts of a request based on struct tags.
//
// Example:
//
//	r := roamer.NewRoamer(
//	    roamer.WithParsers(
//	        parser.NewQuery(),    // 'query' tag for URL parameters
//	        parser.NewHeader(),   // 'header' tag for HTTP headers
//	        parser.NewCookie(),   // 'cookie' tag for cookies
//	    ),
//	)
func WithParsers(parsers ...Parser) OptionsFunc {
	return func(r *Roamer) {
		assignExtensions := make([]assign.ExtensionFunc, 0, len(parsers))
		for _, p := range parsers {
			r.parsers[p.Tag()] = p

			if ext, ok := p.(AssignExtensions); ok {
				assignExtensions = append(assignExtensions, ext.AssignExtensions()...)
			}
		}

		if len(assignExtensions) > 0 {
			r.assignExtensions = append(r.assignExtensions, assignExtensions...)
		}
	}
}

// WithDecoders registers decoders for parsing request bodies.
// Decoders handle different content types like JSON, XML, or form data.
//
// Example:
//
//	r := roamer.NewRoamer(
//	    roamer.WithDecoders(
//	        decoder.NewJSON(),               // JSON bodies
//	        decoder.NewFormURL(),            // URL-encoded forms
//	        decoder.NewMultipartFormData(),  // Multipart forms
//	    ),
//	)
func WithDecoders(decoders ...Decoder) OptionsFunc {
	return func(r *Roamer) {
		assignExtensions := make([]assign.ExtensionFunc, 0, len(decoders))
		for _, d := range decoders {
			r.decoders[d.ContentType()] = d

			if ext, ok := d.(AssignExtensions); ok {
				assignExtensions = append(assignExtensions, ext.AssignExtensions()...)
			}
		}

		if len(assignExtensions) > 0 {
			r.assignExtensions = append(r.assignExtensions, assignExtensions...)
		}
	}
}

// WithFormatters registers formatters that process values after parsing.
// Formatters handle operations like string trimming or case conversion.
//
// Example:
//
//	r := roamer.NewRoamer(
//	    roamer.WithFormatters(
//	        formatter.NewString(), // Apply 'string' tag formatters
//	    ),
//	)
//
//	// Example usage:
//	type User struct {
//	    Name string `json:"name" string:"trim_space"` // Trim spaces
//	}
func WithFormatters(formatters ...Formatter) OptionsFunc {
	return func(r *Roamer) {
		for _, f := range formatters {
			if i, ok := f.(ReflectValueFormatter); ok {
				r.reflectValueFormatters[f.Tag()] = i

				continue
			}

			r.formatters[f.Tag()] = f
		}
	}
}

// WithSkipFilled controls whether to skip fields with non-zero values.
// When true (default), existing non-zero values won't be overwritten.
//
// Example:
//
//	// Override even filled fields
//	r := roamer.NewRoamer(
//	    roamer.WithSkipFilled(false),
//	)
func WithSkipFilled(skip bool) OptionsFunc {
	return func(r *Roamer) {
		r.skipFilled = skip
	}
}
