package formatter

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringFormatter_Format_Successfully(t *testing.T) {
	tests := []struct {
		name       string
		formatter  *String
		tag        reflect.StructTag
		initialVal string
		wantVal    string
	}{
		{
			name:       "trim_space",
			formatter:  NewString(),
			tag:        reflect.StructTag(`string:"trim_space"`),
			initialVal: "  test  ",
			wantVal:    "test",
		},
		{
			name:       "multiple formatters",
			formatter:  NewString(WithStringsFormatters(StringsFormatters{"uppercase": wrapStringFunc(strings.ToUpper)})),
			tag:        reflect.StructTag(`string:"trim_space,uppercase"`),
			initialVal: "  test  ",
			wantVal:    "TEST",
		},
		{
			name:       "no tag",
			formatter:  NewString(),
			tag:        reflect.StructTag(``),
			initialVal: "  test  ",
			wantVal:    "  test  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := tt.initialVal
			err := tt.formatter.Format(tt.tag, &val)
			require.NoError(t, err)
			require.Equal(t, tt.wantVal, val)
		})
	}
}

func TestStringFormatter_Format_Failure(t *testing.T) {
	tests := []struct {
		name      string
		formatter *String
		tag       reflect.StructTag
		ptr       any
	}{
		{
			name:      "formatter not found",
			formatter: NewString(),
			tag:       reflect.StructTag(`string:"nonexistent"`),
			ptr:       new(string),
		},
		{
			name:      "not a pointer",
			formatter: NewString(),
			tag:       reflect.StructTag(`string:"trim_space"`),
			ptr:       "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.formatter.Format(tt.tag, tt.ptr)
			require.Error(t, err)
		})
	}
}
