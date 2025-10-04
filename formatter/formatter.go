// Package formatter provides value formatters for post-processing parsed data.
//
// Formatters transform values after they have been extracted from HTTP requests,
// enabling data normalization, validation, and transformation. The package provides
// built-in formatters for common operations and supports custom formatter creation.
//
// # Built-in Formatters
//
//   - String: Text transformations (trim, case conversion, encoding, etc.)
//   - Numeric: Number constraints and operations (min, max, abs, round, etc.)
//   - Time: Time zone conversion and manipulation (timezone, truncate, etc.)
//   - Slice: Collection operations (unique, sort, limit, compact, etc.)
//
// # Basic Usage
//
//	type UserRequest struct {
//	    Email  string    `json:"email" string:"trim_space,lower"`
//	    Age    int       `query:"age" numeric:"min=0,max=120"`
//	    Tags   []string  `query:"tags" slice:"unique,sort"`
//	    Start  time.Time `json:"start" time:"timezone=UTC"`
//	}
//
//	r := roamer.NewRoamer(
//	    roamer.WithFormatters(
//	        formatter.NewString(),
//	        formatter.NewNumeric(),
//	        formatter.NewTime(),
//	        formatter.NewSlice(),
//	    ),
//	)
//
// # Formatter Chaining
//
// Multiple formatters can be chained using comma-separated values:
//
//	Email string `json:"email" string:"trim_space,lower,trim_suffix=@example.com"`
//
// Formatters are applied in the order they appear in the tag.
//
// # Custom Formatters
//
// Extend built-in formatters or create new ones:
//
//	customFormatter := func(s string, arg string) (string, error) {
//	    return strings.ReplaceAll(s, arg, "***"), nil
//	}
//
//	strFormatter := formatter.NewString(
//	    formatter.WithStringFormatter("redact", customFormatter),
//	)
//
// # Thread Safety
//
// All built-in formatters are safe for concurrent use and should be reused
// across multiple requests for optimal performance.
package formatter

import (
	"reflect"
	"strings"
)

const (
	// SplitSymbol is the default character used to separate multiple formatter operations
	// in a single format tag. For example: "trim_space,lower_case".
	SplitSymbol = ","

	// SplitSymbolArgument is the character used to separate formatter name from its argument.
	// For example: "max_length=100" where "=" separates "max_length" from "100".
	SplitSymbolArgument = "="

	// SplitSymbolMultipleArguments is the character used to separate multiple arguments
	// within a single formatter operation. For example: "range=1:100" where ":"
	// separates the minimum and maximum values.
	SplitSymbolMultipleArguments = ":"
)

// NumericFormatterFunc is a function type for numeric transformations.
// Functions of this type receive a pointer to a numeric value and an optional
// argument string, then perform transformations on the numeric data.
//
// Parameters:
//   - ptr: Pointer to the numeric value to be formatted.
//   - arg: Optional argument string containing formatting parameters.
//
// Returns:
//   - error: An error if formatting fails, or nil if successful.
type NumericFormatterFunc = func(ptr any, arg string) error

// NumericFormatters is a map of named numeric formatting functions.
// Keys are formatter names (used in struct tags), values are the
// corresponding formatter functions.
type NumericFormatters map[string]NumericFormatterFunc

// SliceFormatterFunc is a function type for slice transformations.
// Functions of this type receive a reflect.Value representing a slice
// and an optional argument string, then perform transformations on the slice data.
//
// Parameters:
//   - slice: reflect.Value representing the slice to be formatted.
//   - arg: Optional argument string containing formatting parameters.
//
// Returns:
//   - error: An error if formatting fails, or nil if successful.
type SliceFormatterFunc = func(slice reflect.Value, arg string) error

// SliceFormatters is a map of named slice formatting functions.
// Keys are formatter names (used in struct tags), values are the
// corresponding formatter functions.
type SliceFormatters map[string]SliceFormatterFunc

// ParseFormatter parses formatter name and arguments from a tag part.
// It separates the formatter name from its arguments using SplitSymbolArgument ("=").
// If no argument separator is found, returns the entire string as the name
// with an empty argument string.
//
// Example:
//   - "trim_space" -> name="trim_space", arg=""
//   - "max_length=100" -> name="max_length", arg="100"
//   - "range=1:100" -> name="range", arg="1:100"
//
// Parameters:
//   - tagPart: The formatter specification string from a struct tag.
//
// Returns:
//   - name: The formatter name (trimmed of whitespace).
//   - arg: The argument string (everything after the "=" symbol).
func ParseFormatter(tagPart string) (name, arg string) {
	if idx := strings.Index(tagPart, SplitSymbolArgument); idx != -1 {
		return strings.TrimSpace(tagPart[:idx]), tagPart[idx+1:]
	}

	return tagPart, ""
}

// SplitArgs splits a string of arguments into a slice of strings using
// SplitSymbolMultipleArguments (":") as the separator.
// This is useful for formatters that accept multiple parameters.
//
// Example:
//   - "1:100" -> ["1", "100"]
//   - "min:max:step" -> ["min", "max", "step"]
//   - "single" -> ["single"]
//
// Parameters:
//   - args: The argument string containing multiple values separated by ":".
//
// Returns:
//   - []string: A slice of individual argument strings.
func SplitArgs(args string) []string {
	return strings.Split(args, SplitSymbolMultipleArguments)
}
