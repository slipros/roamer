package cache

// StructureOptionsFunc is a function type for configuring a Structure cache instance.
// It follows the functional options pattern to provide flexible configuration.
type StructureOptionsFunc func(*Structure)

// WithDecoders configures the Structure cache with a list of decoder tag names.
// These are used to identify which struct tags should be processed for body decoding.
func WithDecoders(decoders []string) StructureOptionsFunc {
	return func(s *Structure) {
		s.decoders = decoders
	}
}

// WithParsers configures the Structure cache with a list of parser tag names.
// These are used to identify which struct tags should be processed for request parsing.
func WithParsers(parsers []string) StructureOptionsFunc {
	return func(s *Structure) {
		s.parsers = parsers
	}
}

// WithFormatters configures the Structure cache with a list of formatter tag names.
// These are used to identify which struct tags should be processed for value formatting.
func WithFormatters(formatters []string) StructureOptionsFunc {
	return func(s *Structure) {
		s.formatters = formatters
	}
}

// WithReflectValueFormatters configures the Structure cache with a list of reflect value formatter tag names.
// These are used to identify which struct tags should be processed for direct reflect.Value formatting.
func WithReflectValueFormatters(formatters []string) StructureOptionsFunc {
	return func(s *Structure) {
		s.reflectValueFormatters = formatters
	}
}
