package formatter

import (
	"encoding/base64"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// defaultStringFormatters defines the built-in string formatting functions.
var defaultStringFormatters = StringsFormatters{
	"trim_space":    wrapStringFunc(strings.TrimSpace),
	"upper":         wrapStringFunc(strings.ToUpper),
	"lower":         wrapStringFunc(strings.ToLower),
	"title":         wrapStringFunc(toTitle),
	"snake":         wrapStringFunc(toSnake),
	"camel":         wrapStringFunc(toCamel),
	"kebab":         wrapStringFunc(toKebab),
	"base64_encode": wrapStringFunc(base64Encode),
	"base64_decode": wrapStringFunc(base64Decode),
	"url_encode":    wrapStringFunc(urlEncode),
	"url_decode":    wrapStringFunc(urlDecode),
	"sanitize_html": wrapStringFunc(sanitizeHTML),
	"reverse":       wrapStringFunc(reverse),
	"trim_prefix":   trimPrefix,
	"trim_suffix":   trimSuffix,
	"truncate":      truncate,
	"replace":       replace,
	"substring":     substring,
	"pad_left":      padLeft,
	"pad_right":     padRight,
}

// StringFormatterFunc is a function type for string transformations.
// It takes a string input and a slice of optional arguments, returns a transformed string output and error.
type StringFormatterFunc = func(string, string) (string, error)

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
// Formatters can optionally accept arguments separated by '=' sign, with multiple arguments separated by ':'.
//
// The formatters are applied in the order they appear in the tag. For example:
// `string:"trim_space,uppercase"` will first trim spaces, then convert to uppercase.
// `string:"trim_prefix=www.,truncate=10"` will remove "www." prefix then truncate to 10 characters.
// `string:"replace=old:new:2,pad_left=15:0"` will replace "old" with "new" up to 2 times, then pad left with "0" to 15 chars.
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

	str := *strPtr

	for _, f := range strings.Split(tagValue, SplitSymbol) {
		name, arg := ParseFormatter(f)

		formatter, ok := s.formatters[name]
		if !ok {
			return errors.WithStack(rerr.FormatterNotFound{Tag: TagString, Formatter: name})
		}

		var err error
		str, err = formatter(str, arg)
		if err != nil {
			return errors.Wrapf(err, "failed to apply formatter %s", name)
		}
	}

	*strPtr = str

	return nil
}

// Tag returns the name of the struct tag that this formatter handles.
// For the String formatter, this is "string".
func (s *String) Tag() string {
	return TagString
}

// wrapStringFunc wraps a simple string function to match StringFormatterFunc signature
func wrapStringFunc(fn func(string) string) StringFormatterFunc {
	return func(s string, _ string) (string, error) {
		return fn(s), nil
	}
}

// trimPrefix removes the prefix from the string
func trimPrefix(s string, arg string) (string, error) {
	if arg == "" {
		return "", errors.New("trim_prefix requires one argument: prefix to remove")
	}
	return strings.TrimPrefix(s, arg), nil
}

// trimSuffix removes the suffix from the string
func trimSuffix(s string, arg string) (string, error) {
	if arg == "" {
		return "", errors.New("trim_suffix requires one argument: suffix to remove")
	}
	return strings.TrimSuffix(s, arg), nil
}

// truncate shortens the string to the specified length
func truncate(s string, arg string) (string, error) {
	if arg == "" {
		return "", errors.New("truncate requires one argument: length")
	}

	length, err := strconv.Atoi(arg)
	if err != nil {
		return "", errors.Wrapf(err, "invalid length argument for truncate: %s", arg)
	}

	if length < 0 {
		return "", errors.New("truncate length cannot be negative")
	}

	if len(s) > length {
		return s[:length], nil
	}

	return s, nil
}

// replace replaces occurrences of old substring with new substring
func replace(s string, arg string) (string, error) {
	args := SplitArgs(arg)
	if len(args) < 2 || args[0] == "" {
		return "", errors.New("replace requires two arguments: old and new substrings")
	}

	oldPart, newPart := args[0], args[1]
	n := -1 // replace all occurrences by default

	if len(args) >= 3 {
		var err error
		n, err = strconv.Atoi(args[2])
		if err != nil {
			return "", errors.Wrapf(err, "invalid count argument for replace: %s", args[2])
		}
	}

	return strings.Replace(s, oldPart, newPart, n), nil
}

// substring extracts a substring from the string
func substring(s string, arg string) (string, error) {
	args := SplitArgs(arg)
	if len(args) == 0 || args[0] == "" {
		return "", errors.New("substring requires at least one argument: start index")
	}

	start, err := strconv.Atoi(args[0])
	if err != nil {
		return "", errors.Wrapf(err, "invalid start index for substring: %s", args[0])
	}

	end := len(s)
	if len(args) >= 2 {
		if args[1] == "" {
			return "", errors.New("invalid end index for substring: empty value")
		}
		end, err = strconv.Atoi(args[1])
		if err != nil {
			return "", errors.Wrapf(err, "invalid end index for substring: %s", args[1])
		}
	}

	// Bounds checks after parsing all args
	if start < 0 || start > len(s) {
		return "", nil
	}

	if end < start {
		return "", nil
	}

	if end > len(s) {
		end = len(s)
	}

	return s[start:end], nil
}

// padLeft pads the string with specified character to reach target length
func padLeft(s string, arg string) (string, error) {
	args := SplitArgs(arg)
	if len(args) == 0 || args[0] == "" {
		return "", errors.New("pad_left requires at least one argument: target length")
	}

	length, err := strconv.Atoi(args[0])
	if err != nil {
		return "", errors.Wrapf(err, "invalid length for pad_left: %s", args[0])
	}

	padChar := " "
	if len(args) >= 2 && len(args[1]) > 0 {
		padChar = string([]rune(args[1])[0]) // take first character
	}

	runeCount := utf8.RuneCountInString(s)
	if runeCount >= length {
		return s, nil
	}

	padding := strings.Repeat(padChar, length-runeCount)
	return padding + s, nil
}

// padRight pads the string with specified character to reach target length
func padRight(s string, arg string) (string, error) {
	args := SplitArgs(arg)
	if len(args) == 0 || args[0] == "" {
		return "", errors.New("pad_right requires at least one argument: target length")
	}

	length, err := strconv.Atoi(args[0])
	if err != nil {
		return "", errors.Wrapf(err, "invalid length for pad_right: %s", args[0])
	}

	padChar := " "
	if len(args) >= 2 && len(args[1]) > 0 {
		padChar = string([]rune(args[1])[0]) // take first character
	}

	runeCount := utf8.RuneCountInString(s)
	if runeCount >= length {
		return s, nil
	}

	padding := strings.Repeat(padChar, length-runeCount)
	return s + padding, nil
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
