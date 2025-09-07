package formatter

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rerr "github.com/slipros/roamer/err"
)

func TestNewString(t *testing.T) {
	t.Parallel()

	t.Run("DefaultFormatters", func(t *testing.T) {
		t.Parallel()

		s := NewString()
		require.NotNil(t, s)
		assert.Equal(t, TagString, s.Tag())

		// Check if all default formatters are loaded
		for name := range defaultStringFormatters {
			_, ok := s.formatters[name]
			assert.True(t, ok, "default formatter %s not found", name)
		}
	})

	t.Run("WithCustomFormatter", func(t *testing.T) {
		t.Parallel()

		customFormatter := func(s string, _ string) (string, error) {
			return "custom_" + s, nil
		}
		s := NewString(WithStringFormatter("custom", customFormatter))
		require.NotNil(t, s)

		// Test the custom formatter
		formatted, err := s.formatters["custom"]("test", "")
		require.NoError(t, err)
		assert.Equal(t, "custom_test", formatted)
	})

	t.Run("WithCustomFormatters", func(t *testing.T) {
		t.Parallel()

		customFormatters := StringsFormatters{
			"custom1": func(s string, _ string) (string, error) { return "c1_" + s, nil },
			"custom2": func(s string, _ string) (string, error) { return "c2_" + s, nil },
		}
		s := NewString(WithStringsFormatters(customFormatters))
		require.NotNil(t, s)

		// Test custom formatters
		f1, err := s.formatters["custom1"]("test", "")
		require.NoError(t, err)
		assert.Equal(t, "c1_test", f1)

		f2, err := s.formatters["custom2"]("test", "")
		require.NoError(t, err)
		assert.Equal(t, "c2_test", f2)
	})
}

func TestString_Format_Successfully(t *testing.T) {
	t.Parallel()

	s := NewString()

	testCases := []struct {
		name     string
		tag      string
		initial  string
		expected string
	}{
		{
			name:     "TrimSpace",
			tag:      `string:"trim_space"`,
			initial:  "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "ToUpper",
			tag:      `string:"upper"`,
			initial:  "hello",
			expected: "HELLO",
		},
		{
			name:     "ToLower",
			tag:      `string:"lower"`,
			initial:  "WORLD",
			expected: "world",
		},
		{
			name:     "Title",
			tag:      `string:"title"`,
			initial:  "hello world",
			expected: "Hello World",
		},
		{
			name:     "SnakeCase",
			tag:      `string:"snake"`,
			initial:  "helloWorld",
			expected: "hello_world",
		},
		{
			name:     "CamelCase",
			tag:      `string:"camel"`,
			initial:  "hello_world",
			expected: "HelloWorld",
		},
		{
			name:     "KebabCase",
			tag:      `string:"kebab"`,
			initial:  "helloWorld",
			expected: "hello-world",
		},
		{
			name:     "Base64Encode",
			tag:      `string:"base64_encode"`,
			initial:  "hello",
			expected: "aGVsbG8=",
		},
		{
			name:     "Base64Decode",
			tag:      `string:"base64_decode"`,
			initial:  "aGVsbG8=",
			expected: "hello",
		},
		{
			name:     "Base64Decode_Invalid",
			tag:      `string:"base64_decode"`,
			initial:  "invalid-base64",
			expected: "invalid-base64",
		},
		{
			name:     "URLEncode",
			tag:      `string:"url_encode"`,
			initial:  "a=b&c=d",
			expected: "a%3Db%26c%3Dd",
		},
		{
			name:     "URLDecode",
			tag:      `string:"url_decode"`,
			initial:  "a%3Db%26c%3Dd",
			expected: "a=b&c=d",
		},
		{
			name:     "URLDecode_Invalid",
			tag:      `string:"url_decode"`,
			initial:  "%invalid-url",
			expected: "%invalid-url",
		},
		{
			name:     "SanitizeHTML",
			tag:      `string:"sanitize_html"`,
			initial:  "<p>hello</p>",
			expected: "&lt;p&gt;hello&lt;/p&gt;",
		},
		{
			name:     "Reverse",
			tag:      `string:"reverse"`,
			initial:  "hello",
			expected: "olleh",
		},
		{
			name:     "TrimPrefix",
			tag:      `string:"trim_prefix=pre"`,
			initial:  "prefix_text",
			expected: "fix_text",
		},
		{
			name:     "TrimSuffix",
			tag:      `string:"trim_suffix=fix"`,
			initial:  "text_postfix",
			expected: "text_post",
		},
		{
			name:     "Truncate",
			tag:      `string:"truncate=5"`,
			initial:  "123456789",
			expected: "12345",
		},
		{
			name:     "Truncate_Shorter",
			tag:      `string:"truncate=10"`,
			initial:  "12345",
			expected: "12345",
		},
		{
			name:     "Replace",
			tag:      `string:"replace=old:new"`,
			initial:  "old string with old value",
			expected: "new string with new value",
		},
		{
			name:     "Replace_WithCount",
			tag:      `string:"replace=a:b:2"`,
			initial:  "aaabbbaaa",
			expected: "bbabbbaaa",
		},
		{
			name:     "Substring",
			tag:      `string:"substring=1:5"`,
			initial:  "0123456789",
			expected: "1234",
		},
		{
			name:     "Substring_ToEnd",
			tag:      `string:"substring=5"`,
			initial:  "0123456789",
			expected: "56789",
		},
		{
			name:     "Substring_StartOutOfBounds",
			tag:      `string:"substring=15"`,
			initial:  "0123456789",
			expected: "",
		},
		{
			name:     "Substring_EndOutOfBounds",
			tag:      `string:"substring=5:15"`,
			initial:  "0123456789",
			expected: "56789",
		},
		{
			name:     "Substring_StartAfterEnd",
			tag:      `string:"substring=5:1"`,
			initial:  "0123456789",
			expected: "",
		},
		{
			name:     "PadLeft",
			tag:      `string:"pad_left=10:_"`,
			initial:  "text",
			expected: "______text",
		},
		{
			name:     "PadLeft_NoChar",
			tag:      `string:"pad_left=10"`,
			initial:  "text",
			expected: "      text",
		},
		{
			name:     "PadLeft_Longer",
			tag:      `string:"pad_left=3:_"`,
			initial:  "text",
			expected: "text",
		},
		{
			name:     "PadRight",
			tag:      `string:"pad_right=10:_"`,
			initial:  "text",
			expected: "text______",
		},
		{
			name:     "PadRight_NoChar",
			tag:      `string:"pad_right=10"`,
			initial:  "text",
			expected: "text      ",
		},
		{
			name:     "PadRight_Longer",
			tag:      `string:"pad_right=3:_"`,
			initial:  "text",
			expected: "text",
		},
		{
			name:     "MultipleFormatters",
			tag:      `string:"trim_space,upper,truncate=5"`,
			initial:  "  hello world  ",
			expected: "HELLO",
		},
		{
			name:     "NoStringTag",
			tag:      `other:"tag"`,
			initial:  "no change",
			expected: "no change",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			val := tc.initial
			err := s.Format(reflect.StructTag(tc.tag), &val)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, val)
		})
	}
}

func TestString_Format_Failure(t *testing.T) {
	t.Parallel()

	s := NewString(
		WithStringFormatter("error_formatter", func(s string, _ string) (string, error) {
			return "", assert.AnError
		}),
	)

	testCases := []struct {
		name         string
		tag          string
		initialValue any
		errAs        error
		errIs        error
	}{
		{
			name:         "NotAStringPointer",
			tag:          `string:"trim_space"`,
			initialValue: new(int),
			errIs:        rerr.NotSupported,
		},
		{
			name:         "FormatterNotFoundError",
			tag:          `string:"non_existent"`,
			initialValue: new(string),
			errAs:        &rerr.FormatterNotFoundError{},
		},
		{
			name:         "FormatterReturnsError",
			tag:          `string:"error_formatter"`,
			initialValue: new(string),
		},
		{
			name:         "TrimPrefix_MissingArg",
			tag:          `string:"trim_prefix="`,
			initialValue: new(string),
		},
		{
			name:         "TrimSuffix_MissingArg",
			tag:          `string:"trim_suffix="`,
			initialValue: new(string),
		},
		{
			name:         "Truncate_EmptyArg",
			tag:          `string:"truncate="`,
			initialValue: new(string),
		},
		{
			name:         "Truncate_InvalidArg",
			tag:          `string:"truncate=abc"`,
			initialValue: new(string),
		},
		{
			name:         "Truncate_NegativeLength",
			tag:          `string:"truncate=-1"`,
			initialValue: new(string),
		},
		{
			name:         "Replace_MissingArgs",
			tag:          `string:"replace=old"`,
			initialValue: new(string),
		},
		{
			name:         "Replace_InvalidCount",
			tag:          `string:"replace=a:b:c"`,
			initialValue: new(string),
		},
		{
			name:         "Substring_EmptyArg",
			tag:          `string:"substring="`,
			initialValue: new(string),
		},
		{
			name:         "Substring_InvalidStartIndex",
			tag:          `string:"substring=a:5"`,
			initialValue: new(string),
		},
		{
			name:         "Substring_InvalidEndIndex",
			tag:          `string:"substring=1:b"`,
			initialValue: new(string),
		},
		{
			name:         "Substring_EmptyEnd",
			tag:          `string:"substring=1:"`,
			initialValue: new(string),
		},
		{
			name:         "PadLeft_EmptyArg",
			tag:          `string:"pad_left="`,
			initialValue: new(string),
		},
		{
			name:         "PadLeft_InvalidArg",
			tag:          `string:"pad_left=abc"`,
			initialValue: new(string),
		},
		{
			name:         "PadRight_EmptyArg",
			tag:          `string:"pad_right="`,
			initialValue: new(string),
		},
		{
			name:         "PadRight_InvalidArg",
			tag:          `string:"pad_right=abc"`,
			initialValue: new(string),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			val := tc.initialValue
			err := s.Format(reflect.StructTag(tc.tag), val)

			if tc.errIs != nil {
				require.ErrorIs(t, err, tc.errIs)
			} else if tc.errAs != nil {
				require.ErrorAs(t, err, &tc.errAs)
			} else {
				require.Error(t, err)
			}
		})
	}
}
