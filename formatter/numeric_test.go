package formatter

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rerr "github.com/slipros/roamer/err"
)

func createNumericTestTag(value string) reflect.StructTag {
	return reflect.StructTag(`numeric:"` + value + `"`)
}

func TestNewNumeric(t *testing.T) {
	t.Parallel()

	t.Run("DefaultFormatters", func(t *testing.T) {
		t.Parallel()

		s := NewNumeric()
		require.NotNil(t, s)
		assert.Equal(t, TagNumeric, s.Tag())

		for name := range defaultNumericFormatters {
			_, ok := s.formatters[name]
			assert.True(t, ok, "default formatter %s not found", name)
		}
	})

	t.Run("WithCustomFormatter", func(t *testing.T) {
		t.Parallel()

		customFormatter := func(ptr any, arg string) error {
			return nil
		}
		f := NewNumeric(WithNumericFormatter("custom", customFormatter))
		require.NotNil(t, f)

		_, ok := f.formatters["custom"]
		assert.True(t, ok)
	})

	t.Run("WithCustomFormatters", func(t *testing.T) {
		t.Parallel()

		customFormatters := NumericFormatters{
			"custom1": func(ptr any, arg string) error { return nil },
			"custom2": func(ptr any, arg string) error { return nil },
		}
		f := NewNumeric(WithNumericFormatters(customFormatters))
		require.NotNil(t, f)

		_, ok := f.formatters["custom1"]
		assert.True(t, ok)
		_, ok = f.formatters["custom2"]
		assert.True(t, ok)
	})
}

func TestNumeric_Format_Successfully(t *testing.T) {
	t.Parallel()

	customFormatter := func(ptr any, arg string) error {
		switch v := ptr.(type) {
		case *int:
			*v = 100
		}
		return nil
	}

	f := NewNumeric(WithNumericFormatter("custom", customFormatter))

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

		// Custom formatter
		{name: "Custom formatter", tag: createNumericTestTag("custom"), input: int(5), expected: int(100)},
		// applyMinMax
		{name: "Min/Max int8 within", tag: createNumericTestTag("min=10,max=100"), input: int8(50), expected: int8(50)},
		{name: "Min/Max int16 below", tag: createNumericTestTag("min=10,max=100"), input: int16(5), expected: int16(10)},
		{name: "Min/Max int64 above", tag: createNumericTestTag("min=10,max=100"), input: int64(150), expected: int64(100)},
		{name: "Min/Max float32 above", tag: createNumericTestTag("min=10.5,max=100.5"), input: float32(150.5), expected: float32(100.5)},

		// applyAbs
		{name: "Abs int8 negative", tag: createNumericTestTag("abs"), input: int8(-5), expected: int8(5)},
		{name: "Abs int16 negative", tag: createNumericTestTag("abs"), input: int16(-5), expected: int16(5)},
		{name: "Abs int64 negative", tag: createNumericTestTag("abs"), input: int64(-5), expected: int64(5)},

		// applyFloatFunc
		{name: "Round float32 up", tag: createNumericTestTag("round"), input: float32(5.7), expected: float32(6)},
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
			case int8:
				val := v
				ptr = &val
			case int16:
				val := v
				ptr = &val
			case int64:
				val := v
				ptr = &val
			case float32:
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
			case int8:
				assert.Equal(t, v, *ptr.(*int8))
			case int16:
				assert.Equal(t, v, *ptr.(*int16))
			case int64:
				assert.Equal(t, v, *ptr.(*int64))
			case float32:
				assert.Equal(t, v, *ptr.(*float32))
			}
		})
	}
}

func TestNumeric_Format_Failure(t *testing.T) {
	t.Parallel()

	f := NewNumeric()

	tests := []struct {
		name  string
		tag   reflect.StructTag
		input any
		errAs error
		errIs error
	}{
		{name: "Unsupported type", tag: createNumericTestTag("min=10"), input: new(string), errIs: rerr.NotSupported},
		{name: "Invalid min value", tag: createNumericTestTag("min=abc"), input: new(int)},
		{name: "Invalid max value", tag: createNumericTestTag("max=xyz"), input: new(int)},
		{name: "Formatter not found", tag: createNumericTestTag("non_existent"), input: new(int), errAs: &rerr.FormatterNotFoundError{}},
		{name: "Min invalid value", tag: createNumericTestTag("min=abc"), input: new(int8)},
		{name: "Max invalid value", tag: createNumericTestTag("max=xyz"), input: new(int16)},
		{name: "Abs unsupported type", tag: createNumericTestTag("abs"), input: new(string), errIs: rerr.NotSupported},
		{name: "Float func unsupported type", tag: createNumericTestTag("round"), input: new(int), errIs: rerr.NotSupported},
		{name: "Min invalid value large number", tag: createNumericTestTag("min=99999999999999999999999999999999999999"), input: new(int8)},
		{name: "Min out of range", tag: createNumericTestTag("min=128"), input: new(int8)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := f.Format(tc.tag, tc.input)
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
