package formatter

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

var defaultStringFormatters = StringsFormatters{
	"trim_space": strings.TrimSpace,
}

// StringFormatterFunc string formatter func.
type StringFormatterFunc = func(string) string

// StringsFormatters strings formatters.
type StringsFormatters map[string]StringFormatterFunc

const (
	// TagString string tag.
	TagString = "string"
)

// String is a string formatter.
type String struct {
	formatters StringsFormatters
}

// NewString returns new string formatter.
func NewString(opts ...StringOptionsFunc) *String {
	s := String{
		formatters: defaultStringFormatters,
	}

	for _, opt := range opts {
		opt(&s)
	}

	return &s
}

// Format format string.
func (s *String) Format(tag reflect.StructTag, ptr any) error {
	tagValue, ok := tag.Lookup(TagString)
	if !ok {
		return nil
	}

	strPtr, ok := ptr.(*string)
	if !ok {
		return errors.Wrapf(rerr.NotSupported, "%T", ptr)
	}

	if strings.Contains(tagValue, ",") {
		str := *strPtr
		for _, tagValue := range strings.Split(tagValue, ",") {
			name := strings.TrimSpace(tagValue)
			formatter, ok := s.formatters[name]
			if !ok {
				return errors.WithStack(rerr.FormatterNotFound{Tag: TagString, Formatter: name})
			}

			str = formatter(str)
		}

		*strPtr = str

		return nil
	}

	formatter, ok := s.formatters[tagValue]
	if !ok {
		return errors.WithStack(rerr.FormatterNotFound{Tag: TagString, Formatter: tagValue})
	}

	*strPtr = formatter(*strPtr)

	return nil
}

// Tag returns working tag.
func (s *String) Tag() string {
	return TagString
}
