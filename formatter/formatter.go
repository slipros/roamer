// Package formatter provides value formatters for post-processing parsed data.
// Formatters allow transforming values after they have been parsed from HTTP requests,
// such as trimming strings, converting case, or applying other transformations.
//
// The package is designed to be extensible, allowing users to create custom formatters
// for specific needs.
package formatter

import (
	"reflect"
	"strings"
)

const (
	SplitSymbol                  = ","
	SplitSymbolArgument          = "="
	SplitSymbolMultipleArguments = ":"
)

// NumericFormatterFunc is a function type for numeric transformations.
type NumericFormatterFunc = func(ptr any, arg string) error

// NumericFormatters is a map of named numeric formatting functions.
type NumericFormatters map[string]NumericFormatterFunc

// SliceFormatterFunc is a function type for slice transformations.
type SliceFormatterFunc = func(slice reflect.Value, arg string) error

// SliceFormatters is a map of named slice formatting functions.
type SliceFormatters map[string]SliceFormatterFunc

// ParseFormatter parses formatter name and arguments from tag part.
func ParseFormatter(tagPart string) (name, arg string) {
	if idx := strings.Index(tagPart, SplitSymbolArgument); idx != -1 {
		return strings.TrimSpace(tagPart[:idx]), tagPart[idx+1:]
	}

	return tagPart, ""
}

// SplitArgs splits a string of arguments into a slice of strings.
func SplitArgs(args string) []string {
	return strings.Split(args, SplitSymbolMultipleArguments)
}
