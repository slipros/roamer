package formatter

import (
	"encoding/base64"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// defaultStringFormatters defines the built-in string formatting functions.
var defaultStringFormatters = StringsFormatters{
	"trim_space":    strings.TrimSpace,
	"upper":         strings.ToUpper,
	"lower":         strings.ToLower,
	"title":         toTitle,
	"snake":         toSnake,
	"camel":         toCamel,
	"kebab":         toKebab,
	"base64_encode": base64Encode,
	"base64_decode": base64Decode,
	"url_encode":    urlEncode,
	"url_decode":    urlDecode,
	"sanitize_html": sanitizeHTML,
	"reverse":       reverse,
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

// WithStringsFormatters adds multiple custom string formatter functions from a map.
// This allows for bulk addition of formatters.
func WithStringsFormatters(formatters StringsFormatters) StringOptionsFunc {
	return func(s *String) {
		for name, formatter := range formatters {
			s.formatters[name] = formatter
		}
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

	formatters := strings.Split(tagValue, SplitSymbol)
	str := *strPtr

	for _, f := range formatters {
		name, arg := parseFormatter(f)
		name = strings.TrimSpace(name)

		formatter, ok := s.formatters[name]
		if ok {
			str = formatter(str)
			continue
		}

		// Handle formatters with arguments
		switch name {
		case "trim_prefix":
			str = strings.TrimPrefix(str, arg)
		case "trim_suffix":
			str = strings.TrimSuffix(str, arg)
		case "truncate":
			length, err := strconv.Atoi(arg)
			if err != nil {
				return errors.Wrapf(err, "invalid argument for truncate: %s", arg)
			}
			if len(str) > length {
				str = str[:length]
			}
		default:
			return errors.WithStack(rerr.FormatterNotFound{Tag: TagString, Formatter: name})
		}
	}

	*strPtr = str

	return nil
}

func parseFormatter(tagPart string) (name, arg string) {
	if idx := strings.Index(tagPart, "="); idx != -1 {
		return tagPart[:idx], tagPart[idx+1:]
	}
	return tagPart, ""
}

// Tag returns the name of the struct tag that this formatter handles.
// For the String formatter, this is "string".
func (s *String) Tag() string {
	return TagString
}

func toTitle(s string) string {
	var result strings.Builder
	capNext := true
	for _, r := range s {
		if unicode.IsSpace(r) {
			result.WriteRune(r)
			capNext = true
		} else if capNext {
			result.WriteRune(unicode.ToUpper(r))
			capNext = false
		} else {
			result.WriteRune(unicode.ToLower(r))
		}
	}
	return result.String()
}

func toSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func toCamel(s string) string {
	var result strings.Builder
	upper := true
	for _, r := range s {
		if r == '_' {
			upper = true
		} else if upper {
			result.WriteRune(unicode.ToUpper(r))
			upper = false
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func toKebab(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('-')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func base64Decode(s string) string {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return s // Or handle error appropriately
	}
	return string(decoded)
}

func urlEncode(s string) string {
	return url.QueryEscape(s)
}

func urlDecode(s string) string {
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return s // Or handle error appropriately
	}
	return decoded
}

func sanitizeHTML(s string) string {
	// A very basic sanitizer. For robust protection, a library like bluemonday is recommended.
	return strings.ReplaceAll(strings.ReplaceAll(s, "<", "&lt;"), ">", "&gt;")
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
