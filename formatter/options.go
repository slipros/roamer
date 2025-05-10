// Package formatter provides value formatters for post-processing parsed data.
package formatter

// StringOptionsFunc is a function type for configuring a String formatter.
// It follows the functional options pattern to provide a clean and extensible API.
type StringOptionsFunc = func(*String)

// WithStringFormatters replaces all string formatters with the provided ones.
// This completely overwrites any existing formatters, including the default ones.
//
// Example:
//
//	// Create a string formatter with custom formatters only
//	customFormatters := formatter.StringsFormatters{
//	    "uppercase": strings.ToUpper,
//	    "lowercase": strings.ToLower,
//	}
//
//	strFormatter := formatter.NewString(
//	    formatter.WithStringFormatters(customFormatters),
//	)
//
//	// In this case, the "trim_space" default formatter will no longer be available
func WithStringFormatters(formatters StringsFormatters) StringOptionsFunc {
	return func(s *String) {
		s.formatters = formatters
	}
}

// WithExtendedStringFormatters adds new string formatters to the existing ones.
// This preserves any existing formatters, including the default ones, and adds
// the provided formatters. If a formatter with the same name already exists,
// it will be replaced.
//
// Example:
//
//	// Create a string formatter with default formatters plus custom ones
//	customFormatters := formatter.StringsFormatters{
//	    "uppercase": strings.ToUpper,
//	    "lowercase": strings.ToLower,
//	}
//
//	strFormatter := formatter.NewString(
//	    formatter.WithExtendedStringFormatters(customFormatters),
//	)
//
//	// In this case, both the default formatters and the custom ones will be available
func WithExtendedStringFormatters(formatters StringsFormatters) StringOptionsFunc {
	return func(s *String) {
		for n, f := range formatters {
			s.formatters[n] = f
		}
	}
}
