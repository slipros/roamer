package formatter

// StringOptionsFunc function for setting string options.
type StringOptionsFunc = func(*String)

// WithStringFormatters sets string formatters.
func WithStringFormatters(formatters StringsFormatters) StringOptionsFunc {
	return func(s *String) {
		s.formatters = formatters
	}
}

// WithExtendedStringFormatters extend string formatters.
func WithExtendedStringFormatters(formatters StringsFormatters) StringOptionsFunc {
	return func(s *String) {
		for n, f := range formatters {
			s.formatters[n] = f
		}
	}
}
