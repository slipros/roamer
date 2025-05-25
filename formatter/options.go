package formatter

// StringOptionsFunc is a function type for configuring a String formatter.
// It follows the functional options pattern to provide a clean and extensible API.
type StringOptionsFunc = func(*String)

// WithStringFormatters replaces all string formatters with custom ones.
// Completely overwrites existing formatters, including the default ones.
//
// Example:
//
//	// Custom formatters only
//	customFormatters := formatter.StringsFormatters{
//	    "uppercase": strings.ToUpper,
//	    "lowercase": strings.ToLower,
//	}
//
//	strFormatter := formatter.NewString(
//	    formatter.WithStringFormatters(customFormatters),
//	)
func WithStringFormatters(formatters StringsFormatters) StringOptionsFunc {
	return func(s *String) {
		s.formatters = formatters
	}
}

// WithExtendedStringFormatters adds new formatters to the existing ones.
// Preserves default formatters while adding custom ones.
//
// Example:
//
//	// Add custom formatters while keeping defaults
//	customFormatters := formatter.StringsFormatters{
//	    "uppercase": strings.ToUpper,
//	    "lowercase": strings.ToLower,
//	}
//
//	strFormatter := formatter.NewString(
//	    formatter.WithExtendedStringFormatters(customFormatters),
//	)
func WithExtendedStringFormatters(formatters StringsFormatters) StringOptionsFunc {
	return func(s *String) {
		for n, f := range formatters {
			s.formatters[n] = f
		}
	}
}
