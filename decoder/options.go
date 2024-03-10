package decoder

// contentTypeSetter content type setter.
type contentTypeSetter interface {
	setContentType(contentType string)
}

// skipFilledSetter skip filled setter.
type skipFilledSetter interface {
	setSkipFilled(skip bool)
}

// WithContentType sets content type.
func WithContentType[T contentTypeSetter](contentType string) func(T) {
	return func(d T) {
		d.setContentType(contentType)
	}
}

// WithSkipFilled sets skip filled.
func WithSkipFilled[T skipFilledSetter](skip bool) func(T) {
	return func(d T) {
		d.setSkipFilled(skip)
	}
}
