package formatter

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// defaultStringFormatters defines the built-in string formatting functions.
// Currently, only "trim_space" is supported by default, which removes leading
// and trailing whitespace from strings.
var defaultStringFormatters = StringsFormatters{
	"trim_space": strings.TrimSpace,
}

// StringFormatterFunc is a function type for string transformations.
// It takes a string input and returns a transformed string output.
type StringFormatterFunc = func(string) string

// StringsFormatters is a map of named string formatting functions.
// The keys are the names that can be used in struct tags, and the values
// are the corresponding formatting functions.
type StringsFormatters map[string]StringFormatterFunc

const (
	// TagString is the struct tag name used for string formatting.
	// Fields tagged with this will have the specified formatters applied
	// after parsing.
	// Example: `string:"trim_space"`
	TagString = "string"
)

// WithStringFormatter adds a custom string formatter function.
// This allows extending the String formatter with custom transformations.
//
// Example:
//
//	// Add a custom formatter to convert strings to uppercase
//	upperFormatter := formatter.NewString(
//	    func(s *formatter.String) {
//	        s.formatters["uppercase"] = strings.ToUpper
//	    },
//	)
//
//	// Use it with roamer
//	r := roamer.NewRoamer(
//	    roamer.WithFormatters(upperFormatter),
//	)
//
//	// Example struct using the formatter
//	type UserInput struct {
//	    Email string `json:"email" string:"trim_space,uppercase"`
//	}
func WithStringFormatter(name string, formatter StringFormatterFunc) StringOptionsFunc {
	return func(s *String) {
		s.formatters[name] = formatter
	}
}

// String is a formatter for string values.
// It applies transformations to string fields based on the "string" struct tag.
type String struct {
	formatters StringsFormatters // Map of available string formatters
}

// NewString creates a String formatter that processes string values based on the "string" tag.
// Includes "trim_space" formatter by default, which removes leading/trailing whitespace.
//
// Example:
//
//	// Basic string formatter
//	strFormatter := formatter.NewString()
//
//	// With custom formatters
//	strFormatter := formatter.NewString(
//	    formatter.WithStringFormatter("uppercase", strings.ToUpper),
//	    formatter.WithStringFormatter("lowercase", strings.ToLower),
//	)
func NewString(opts ...StringOptionsFunc) *String {
	s := String{
		formatters: make(StringsFormatters),
	}

	// Copy default formatters to avoid modifying the shared map
	for name, fn := range defaultStringFormatters {
		s.formatters[name] = fn
	}

	for _, opt := range opts {
		opt(&s)
	}

	return &s
}

// Format applies string formatters to a field value based on the struct tag.
// It supports applying multiple formatters by separating them with commas.
//
// The formatters are applied in the order they appear in the tag. For example,
// `string:"trim_space,uppercase"` will first trim spaces, then convert to uppercase.
//
// Parameters:
//   - tag: The struct tag containing formatting instructions.
//   - ptr: A pointer to the string value to be formatted.
//
// Returns:
//   - error: An error if formatting fails or if a formatter is not found,
//     or nil if successful.
func (s *String) Format(tag reflect.StructTag, ptr any) error {
	tagValue, ok := tag.Lookup(TagString)
	if !ok {
		return nil
	}

	strPtr, ok := ptr.(*string)
	if !ok {
		return errors.Wrapf(rerr.NotSupported, "%T", ptr)
	}

	if !strings.Contains(tagValue, SplitSymbol) {
		formatter, ok := s.formatters[tagValue]
		if !ok {
			return errors.WithStack(rerr.FormatterNotFound{Tag: TagString, Formatter: tagValue})
		}

		*strPtr = formatter(*strPtr)

		return nil
	}

	str := *strPtr
	for _, tagValue := range strings.Split(tagValue, SplitSymbol) {
		if unicode.IsSpace(rune(tagValue[0])) || unicode.IsSpace(rune(tagValue[len(tagValue)-1])) {
			tagValue = strings.TrimSpace(tagValue)
		}

		formatter, ok := s.formatters[tagValue]
		if !ok {
			return errors.WithStack(rerr.FormatterNotFound{Tag: TagString, Formatter: tagValue})
		}

		str = formatter(str)
	}

	*strPtr = str

	return nil
}

// Tag returns the name of the struct tag that this formatter handles.
// For the String formatter, this is "string".
func (s *String) Tag() string {
	return TagString
}
