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

// WithAssignExtensions registers additional assignment extension functions.
// These extensions provide custom value assignment capabilities for specific types
// that require special handling beyond standard type conversions.
//
// Assignment extensions are functions that take a value and return an assignment
// function if they can handle that value type. This allows for sophisticated
// type handling and custom conversion logic.
//
// Example:
//
//	customExtension := func(value any) (func(to reflect.Value) error, bool) {
//	    if customType, ok := value.(MyCustomType); ok {
//	        return func(to reflect.Value) error {
//	            // Custom assignment logic
//	            return assign.String(to, customType.String())
//	        }, true
//	    }
//	    return nil, false
//	}
//
//	r := roamer.NewRoamer(
//	    roamer.WithAssignExtensions(customExtension),
//	)
//
// Note: Extensions from parsers and decoders that implement AssignExtensions
// interface are automatically registered. Use this function for standalone
// extension functions that are not tied to specific parsers or decoders.
func WithAssignExtensions(extensions ...assign.ExtensionFunc) OptionsFunc {
	return func(r *Roamer) {
		r.assignExtensions = append(r.assignExtensions, extensions...)
	}
}

// WithPreserveBody enables preservation of the request body after decoding.
// The decoder reads the entire body into memory, decodes it, and then
// replaces http.Request.Body with a new io.ReadCloser containing the same data.
// This allows downstream handlers to read the body again.
//
// WARNING: This option increases memory usage as the entire request body
// is buffered in memory. Use with caution for large request bodies. Consider
// implementing size limits in your HTTP server to prevent excessive memory consumption.
//
// Example:
//
//	// Enable body preservation for middleware that needs to read body multiple times
//	r := roamer.NewRoamer(
//	    roamer.WithDecoders(decoder.NewJSON()),
//	    roamer.WithPreserveBody(),
//	)
//
//	// Now downstream handlers can also read the body
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    // Body was already read by roamer for parsing,
//	    // but can be read again here thanks to preservation
//	    body, _ := io.ReadAll(r.Body)
//	    // ... use body ...
//	}
func WithPreserveBody() OptionsFunc {
	return func(r *Roamer) {
		r.preserveBody = true
	}
}
