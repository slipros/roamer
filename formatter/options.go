package formatter

// StringOptionsFunc is a function type for configuring a String formatter.
// It follows the functional options pattern to provide a clean and extensible API.
type StringOptionsFunc = func(*String)

// WithStringFormatter adds a custom string formatter function.
func WithStringFormatter(name string, formatter StringFormatterFunc) StringOptionsFunc {
	return func(s *String) {
		s.formatters[name] = formatter
	}
}

// WithStringsFormatters adds multiple custom string formatter functions from a map.
func WithStringsFormatters(formatters StringsFormatters) StringOptionsFunc {
	return func(s *String) {
		for name, formatter := range formatters {
			s.formatters[name] = formatter
		}
	}
}

// NumericOptionsFunc is a function type for configuring a Numeric formatter.
type NumericOptionsFunc = func(*Numeric)

// WithNumericFormatter adds a custom numeric formatter function.
func WithNumericFormatter(name string, formatter NumericFormatterFunc) NumericOptionsFunc {
	return func(n *Numeric) {
		n.formatters[name] = formatter
	}
}

// WithNumericFormatters adds multiple custom numeric formatter functions from a map.
func WithNumericFormatters(formatters NumericFormatters) NumericOptionsFunc {
	return func(n *Numeric) {
		for name, formatter := range formatters {
			n.formatters[name] = formatter
		}
	}
}

// SliceOptionsFunc is a function type for configuring a Slice formatter.
type SliceOptionsFunc = func(*Slice)

// WithSliceFormatter adds a custom slice formatter function.
func WithSliceFormatter(name string, formatter SliceFormatterFunc) SliceOptionsFunc {
	return func(s *Slice) {
		s.formatters[name] = formatter
	}
}

// WithSliceFormatters adds multiple custom slice formatter functions from a map.
func WithSliceFormatters(formatters SliceFormatters) SliceOptionsFunc {
	return func(s *Slice) {
		for name, formatter := range formatters {
			s.formatters[name] = formatter
		}
	}
}

// TimeOptionsFunc is a function type for configuring a Time formatter.
type TimeOptionsFunc = func(*Time)

// WithTimeFormatter adds a custom time formatter function.
func WithTimeFormatter(name string, formatter TimeFormatterFunc) TimeOptionsFunc {
	return func(t *Time) {
		t.formatters[name] = formatter
	}
}

// WithTimeFormatters adds multiple custom time formatter functions from a map.
func WithTimeFormatters(formatters TimeFormatters) TimeOptionsFunc {
	return func(t *Time) {
		for name, formatter := range formatters {
			t.formatters[name] = formatter
		}
	}
}
