package formatter

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	rerr "github.com/slipros/roamer/err"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestTag creates a struct tag for testing
func createTestTag(value string) reflect.StructTag {
	return reflect.StructTag(`string:"` + value + `"`)
}

// TestString_Tag tests the Tag method of the String formatter
func TestString_Tag(t *testing.T) {
	f := NewString()
	assert.Equal(t, "string", f.Tag(), "Tag should return 'string'")
}

// TestString_NewString tests the NewString constructor
func TestString_NewString(t *testing.T) {
	// Test default constructor
	f := NewString()
	assert.NotNil(t, f)
	assert.Contains(t, f.formatters, "trim_space", "Default formatters should include 'trim_space'")
	assert.NotNil(t, f.formatters["trim_space"], "The 'trim_space' formatter should be a valid function")

	// Test with custom formatter
	reverseFn := func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	}

	f = NewString(WithStringFormatter("reverse", reverseFn))
	assert.Contains(t, f.formatters, "trim_space", "Default formatters should be preserved")
	assert.Contains(t, f.formatters, "reverse", "Custom formatter should be added")
	assert.NotNil(t, f.formatters["reverse"], "The 'reverse' formatter should be a valid function")

	// Test result of custom formatter
	result := f.formatters["reverse"]("hello")
	assert.Equal(t, "olleh", result, "Custom formatter should work correctly")
}

// TestString_Format_Successfully tests successful formatting scenarios
func TestString_Format_Successfully(t *testing.T) {
	// Define custom formatters for tests
	uppercaseFn := strings.ToUpper
	lowercaseFn := strings.ToLower
	removeVowelsFn := func(s string) string {
		return strings.Map(func(r rune) rune {
			switch r {
			case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
				return -1
			default:
				return r
			}
		}, s)
	}

	// Create formatter with custom functions
	f := NewString(
		WithStringFormatter("uppercase", uppercaseFn),
		WithStringFormatter("lowercase", lowercaseFn),
		WithStringFormatter("remove_vowels", removeVowelsFn),
	)

	tests := []struct {
		name     string
		tag      reflect.StructTag
		input    string
		expected string
	}{
		{
			name:     "Apply trim_space",
			tag:      createTestTag("trim_space"),
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "Apply uppercase",
			tag:      createTestTag("uppercase"),
			input:    "hello world",
			expected: "HELLO WORLD",
		},
		{
			name:     "Apply lowercase",
			tag:      createTestTag("lowercase"),
			input:    "HELLO WORLD",
			expected: "hello world",
		},
		{
			name:     "Apply remove_vowels",
			tag:      createTestTag("remove_vowels"),
			input:    "hello world",
			expected: "hll wrld",
		},
		{
			name:     "Apply multiple formatters",
			tag:      createTestTag("trim_space,uppercase"),
			input:    "  hello world  ",
			expected: "HELLO WORLD",
		},
		{
			name:     "Apply multiple formatters in order",
			tag:      createTestTag("trim_space,uppercase,remove_vowels"),
			input:    "  hello world  ",
			expected: "HLL WRLD",
		},
		{
			name:     "No tag value - should not modify",
			tag:      reflect.StructTag(""),
			input:    "hello world",
			expected: "hello world",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a string pointer for the test
			input := tc.input

			// Apply the formatter
			err := f.Format(tc.tag, &input)

			// Check results
			require.NoError(t, err)
			assert.Equal(t, tc.expected, input, "String formatter should correctly apply transformations")
		})
	}
}

// TestString_Format_Failure tests failure scenarios for formatting
func TestString_Format_Failure(t *testing.T) {
	f := NewString()

	tests := []struct {
		name      string
		tag       reflect.StructTag
		input     any // Changed to interface{} to test non-string pointers
		expectErr error
	}{
		{
			name:      "Non-string pointer",
			tag:       createTestTag("trim_space"),
			input:     new(int),
			expectErr: rerr.NotSupported,
		},
		{
			name:      "Unknown formatter",
			tag:       createTestTag("non_existent"),
			input:     new(string),
			expectErr: rerr.FormatterNotFound{Tag: TagString, Formatter: "non_existent"},
		},
		{
			name:      "Unknown formatter in multi-formatter tag",
			tag:       createTestTag("trim_space,non_existent"),
			input:     new(string),
			expectErr: rerr.FormatterNotFound{Tag: TagString, Formatter: "non_existent"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Apply the formatter
			err := f.Format(tc.tag, tc.input)

			// Check results
			require.Error(t, err)

			var formatterErr rerr.FormatterNotFound
			if errors.As(tc.expectErr, &formatterErr) {
				var actual rerr.FormatterNotFound
				if assert.ErrorAs(t, err, &actual) {
					assert.Equal(t, formatterErr.Tag, actual.Tag)
					assert.Equal(t, formatterErr.Formatter, actual.Formatter)
				}
			}
		})
	}
}

// TestWithStringFormatters tests the WithStringFormatters option
func TestWithStringFormatters_Successfully(t *testing.T) {
	// Create custom formatters map
	customFormatters := StringsFormatters{
		"reverse": func(s string) string {
			runes := []rune(s)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			return string(runes)
		},
		"double": func(s string) string {
			return s + s
		},
	}

	// Test WithStringFormatters option
	f := NewString(WithStringFormatters(customFormatters))

	// Verify formatters are added and defaults are replaced
	assert.NotContains(t, f.formatters, "trim_space", "Default formatters should be replaced")
	assert.Contains(t, f.formatters, "reverse", "Custom formatter should be added")
	assert.Contains(t, f.formatters, "double", "Custom formatter should be added")

	// Test formatters work
	input := "hello"

	// Test reverse formatter
	reverseTag := createTestTag("reverse")
	reverseInput := input
	err := f.Format(reverseTag, &reverseInput)
	require.NoError(t, err)
	assert.Equal(t, "olleh", reverseInput)

	// Test double formatter
	doubleTag := createTestTag("double")
	doubleInput := input
	err = f.Format(doubleTag, &doubleInput)
	require.NoError(t, err)
	assert.Equal(t, "hellohello", doubleInput)
}

// TestWithExtendedStringFormatters tests the WithExtendedStringFormatters option
func TestWithExtendedStringFormatters_Successfully(t *testing.T) {
	// Create base formatter with a custom formatter and the default trim_space
	baseFormatters := StringsFormatters{
		"base":       func(s string) string { return "base:" + s },
		"trim_space": strings.TrimSpace, // Include built-in formatter explicitly
	}

	// Create extended formatters
	extendedFormatters := StringsFormatters{
		"extended": func(s string) string { return "extended:" + s },
	}

	// Test extending formatters
	f := NewString(
		WithStringFormatters(baseFormatters),
		WithExtendedStringFormatters(extendedFormatters),
	)

	// Verify all formatters are present
	assert.Contains(t, f.formatters, "trim_space", "Base formatters should include trim_space")
	assert.Contains(t, f.formatters, "base", "Base formatter should be present")
	assert.Contains(t, f.formatters, "extended", "Extended formatter should be added")

	// Test the formatters work
	input := "test"

	// Test base formatter
	baseTag := createTestTag("base")
	baseInput := input
	err := f.Format(baseTag, &baseInput)
	require.NoError(t, err)
	assert.Equal(t, "base:test", baseInput)

	// Test extended formatter
	extendedTag := createTestTag("extended")
	extendedInput := input
	err = f.Format(extendedTag, &extendedInput)
	require.NoError(t, err)
	assert.Equal(t, "extended:test", extendedInput)
}
