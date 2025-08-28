package formatter

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createNumericTestTag(value string) reflect.StructTag {
	return reflect.StructTag(`numeric:"` + value + `"`)
}

func TestNumeric_Tag(t *testing.T) {
	f := NewNumeric()
	assert.Equal(t, "numeric", f.Tag())
}

func TestNumeric_Format_Successfully(t *testing.T) {
	f := NewNumeric()

	tests := []struct {
		name     string
		tag      reflect.StructTag
		input    any
		expected any
	}{
		// Min tests
		{name: "Min int below", tag: createNumericTestTag("min=10"), input: int(5), expected: int(10)},
		{name: "Min int above", tag: createNumericTestTag("min=10"), input: int(15), expected: int(15)},
		{name: "Min float64 below", tag: createNumericTestTag("min=10.5"), input: float64(5.5), expected: float64(10.5)},
		{name: "Min float64 above", tag: createNumericTestTag("min=10.5"), input: float64(15.5), expected: float64(15.5)},

		// Max tests
		{name: "Max int below", tag: createNumericTestTag("max=100"), input: int(50), expected: int(50)},
		{name: "Max int above", tag: createNumericTestTag("max=100"), input: int(150), expected: int(100)},
		{name: "Max float64 below", tag: createNumericTestTag("max=100.5"), input: float64(50.5), expected: float64(50.5)},
		{name: "Max float64 above", tag: createNumericTestTag("max=100.5"), input: float64(150.5), expected: float64(100.5)},

		// Combined tests
		{name: "Min/Max int within", tag: createNumericTestTag("min=10,max=100"), input: int(50), expected: int(50)},
		{name: "Min/Max int below", tag: createNumericTestTag("min=10,max=100"), input: int(5), expected: int(10)},
		{name: "Min/Max int above", tag: createNumericTestTag("min=10,max=100"), input: int(150), expected: int(100)},
		{name: "Min/Max float64 within", tag: createNumericTestTag("min=10.5,max=100.5"), input: float64(50.5), expected: float64(50.5)},
		{name: "Min/Max float64 below", tag: createNumericTestTag("min=10.5,max=100.5"), input: float64(5.5), expected: float64(10.5)},
		{name: "Min/Max float64 above", tag: createNumericTestTag("min=10.5,max=100.5"), input: float64(150.5), expected: float64(100.5)},

		// Abs tests
		{name: "Abs int positive", tag: createNumericTestTag("abs"), input: int(5), expected: int(5)},
		{name: "Abs int negative", tag: createNumericTestTag("abs"), input: int(-5), expected: int(5)},
		{name: "Abs float64 positive", tag: createNumericTestTag("abs"), input: float64(5.5), expected: float64(5.5)},
		{name: "Abs float64 negative", tag: createNumericTestTag("abs"), input: float64(-5.5), expected: float64(5.5)},

		// Round tests
		{name: "Round float64 up", tag: createNumericTestTag("round"), input: float64(5.7), expected: float64(6)},
		{name: "Round float64 down", tag: createNumericTestTag("round"), input: float64(5.3), expected: float64(5)},
		{name: "Round float64 middle", tag: createNumericTestTag("round"), input: float64(5.5), expected: float64(6)},

		// Ceil tests
		{name: "Ceil float64 up", tag: createNumericTestTag("ceil"), input: float64(5.3), expected: float64(6)},
		{name: "Ceil float64 exact", tag: createNumericTestTag("ceil"), input: float64(5.0), expected: float64(5)},

		// Floor tests
		{name: "Floor float64 down", tag: createNumericTestTag("floor"), input: float64(5.7), expected: float64(5)},
		{name: "Floor float64 exact", tag: createNumericTestTag("floor"), input: float64(5.0), expected: float64(5)},

		// No tag
		{name: "No tag", tag: reflect.StructTag(""), input: int(5), expected: int(5)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var ptr any
			switch v := tc.input.(type) {
			case int:
				val := v
				ptr = &val
			case float64:
				val := v
				ptr = &val
			}

			err := f.Format(tc.tag, ptr)
			require.NoError(t, err)

			switch v := tc.expected.(type) {
			case int:
				assert.Equal(t, v, *ptr.(*int))
			case float64:
				assert.Equal(t, v, *ptr.(*float64))
			}
		})
	}
}

func TestNumeric_Format_Failure(t *testing.T) {
	f := NewNumeric()

	tests := []struct {
		name    string
		tag     reflect.StructTag
		input   any
		wantErr string
	}{
		{name: "Unsupported type", tag: createNumericTestTag("min=10"), input: new(string), wantErr: "not supported"},
		{name: "Invalid min value", tag: createNumericTestTag("min=abc"), input: new(int), wantErr: "invalid min value"},
		{name: "Invalid max value", tag: createNumericTestTag("max=xyz"), input: new(int), wantErr: "invalid max value"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := f.Format(tc.tag, tc.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
