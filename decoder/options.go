package decoder

// contentTypeSetter content type setter.
type contentTypeSetter interface {
	setContentType(contentType string)
}

// WithContentType sets content type.
func WithContentType[T contentTypeSetter](contentType string) func(T) {
	return func(d T) {
		d.setContentType(contentType)
	}
}
