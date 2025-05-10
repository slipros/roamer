// Package decoder provides decoders for extracting data from HTTP request bodies.
package decoder

// contentTypeSetter is an interface for decoders that can have their
// Content-Type header value customized. This is an internal interface
// used to implement generic options.
type contentTypeSetter interface {
	setContentType(contentType string)
}

// skipFilledSetter is an interface for decoders that can control
// whether to skip already filled fields. This is an internal interface
// used to implement generic options.
type skipFilledSetter interface {
	setSkipFilled(skip bool)
}

// WithContentType creates an option function that sets the Content-Type
// header value for a decoder. This allows customizing which Content-Type
// header values the decoder will handle.
//
// This is a generic function that works with any decoder type that
// implements the contentTypeSetter interface.
//
// Example:
//
//	// Create a JSON decoder that handles "application/x-json" instead of "application/json"
//	jsonDecoder := decoder.NewJSON(
//	    decoder.WithContentType("application/x-json"),
//	)
//
//	// Create a XML decoder that handles "text/xml" instead of "application/xml"
//	xmlDecoder := decoder.NewXML(
//	    decoder.WithContentType("text/xml"),
//	)
func WithContentType[T contentTypeSetter](contentType string) func(T) {
	return func(d T) {
		d.setContentType(contentType)
	}
}

// WithSkipFilled creates an option function that controls whether the decoder
// should skip fields that already have non-zero values. When set to true,
// fields that are already filled will not be overwritten by parsed values.
//
// This is a generic function that works with any decoder type that
// implements the skipFilledSetter interface.
//
// Example:
//
//	// Create a JSON decoder that will overwrite all fields, even if they're already filled
//	jsonDecoder := decoder.NewJSON(
//	    decoder.WithSkipFilled(false),
//	)
//
//	// Create a Form decoder that will skip already filled fields
//	formDecoder := decoder.NewFormURL(
//	    decoder.WithSkipFilled(true),
//	)
func WithSkipFilled[T skipFilledSetter](skip bool) func(T) {
	return func(d T) {
		d.setSkipFilled(skip)
	}
}
