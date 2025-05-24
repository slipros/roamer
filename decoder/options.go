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

// WithContentType sets the Content-Type header value that a decoder will handle.
// Works with any decoder implementing contentTypeSetter interface.
//
// Example:
//
//	// Handle custom content types
//	jsonDecoder := decoder.NewJSON(
//	    decoder.WithContentType("application/x-json"),
//	)
//
//	xmlDecoder := decoder.NewXML(
//	    decoder.WithContentType("text/xml"),
//	)
func WithContentType[T contentTypeSetter](contentType string) func(T) {
	return func(d T) {
		d.setContentType(contentType)
	}
}

// WithSkipFilled controls whether decoders skip fields with non-zero values.
// When true, existing values won't be overwritten by decoded values.
//
// Example:
//
//	// Overwrite all fields, even if already filled
//	jsonDecoder := decoder.NewJSON(
//	    decoder.WithSkipFilled(false),
//	)
func WithSkipFilled[T skipFilledSetter](skip bool) func(T) {
	return func(d T) {
		d.setSkipFilled(skip)
	}
}
