package formatter

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rerr "github.com/slipros/roamer/err"
)

func createSliceTestTag(value string) reflect.StructTag {
	return reflect.StructTag(`slice:"` + value + `"`)
}

func TestNewSlice(t *testing.T) {
	t.Parallel()

	t.Run("DefaultFormatters", func(t *testing.T) {
		t.Parallel()

		s := NewSlice()
		require.NotNil(t, s)
		assert.Equal(t, TagSlice, s.Tag())

		for name := range defaultSliceFormatters {
			_, ok := s.formatters[name]
			assert.True(t, ok, "default formatter %s not found", name)
		}
	})

	t.Run("WithCustomFormatter", func(t *testing.T) {
		t.Parallel()

		customFormatter := func(slice reflect.Value, arg string) error {
			return nil
		}
		f := NewSlice(WithSliceFormatter("custom", customFormatter))
		require.NotNil(t, f)

		_, ok := f.formatters["custom"]
		assert.True(t, ok)
	})

	t.Run("WithCustomFormatters", func(t *testing.T) {
		t.Parallel()

		customFormatters := SliceFormatters{
			"custom1": func(slice reflect.Value, arg string) error { return nil },
			"custom2": func(slice reflect.Value, arg string) error { return nil },
		}
		f := NewSlice(WithSliceFormatters(customFormatters))
		require.NotNil(t, f)

		_, ok := f.formatters["custom1"]
		assert.True(t, ok)
		_, ok = f.formatters["custom2"]
		assert.True(t, ok)
	})
}

func TestSlice_Format_Successfully(t *testing.T) {
	t.Parallel()

	customFormatter := func(slice reflect.Value, arg string) error {
		newSlice := reflect.MakeSlice(slice.Type(), 0, slice.Len())
		newSlice = reflect.Append(newSlice, reflect.ValueOf("custom"))
		slice.Set(newSlice)
		return nil
	}

	f := NewSlice(WithSliceFormatter("custom", customFormatter))

	tests := []struct {
		name     string
		tag      reflect.StructTag
		input    any
		expected any
	}{
		{
			name:     "Unique strings",
			tag:      createSliceTestTag("unique"),
			input:    &[]string{"a", "b", "a", "c", "b"},
			expected: &[]string{"a", "b", "c"},
		},
		{
			name:     "Unique ints",
			tag:      createSliceTestTag("unique"),
			input:    &[]int{1, 2, 1, 3, 2},
			expected: &[]int{1, 2, 3},
		},
		{
			name:     "Already unique",
			tag:      createSliceTestTag("unique"),
			input:    &[]string{"a", "b", "c"},
			expected: &[]string{"a", "b", "c"},
		},
		{
			name:     "Empty slice",
			tag:      createSliceTestTag("unique"),
			input:    &[]string{},
			expected: &[]string{},
		},

		// Sort tests
		{
			name:     "Sort strings asc",
			tag:      createSliceTestTag("sort"),
			input:    &[]string{"c", "a", "b"},
			expected: &[]string{"a", "b", "c"},
		},
		{
			name:     "Sort strings desc",
			tag:      createSliceTestTag("sort_desc"),
			input:    &[]string{"c", "a", "b"},
			expected: &[]string{"c", "b", "a"},
		},
		{
			name:     "Sort ints asc",
			tag:      createSliceTestTag("sort"),
			input:    &[]int{3, 1, 2},
			expected: &[]int{1, 2, 3},
		},
		{
			name:     "Sort ints desc",
			tag:      createSliceTestTag("sort_desc"),
			input:    &[]int{3, 1, 2},
			expected: &[]int{3, 2, 1},
		},

		// Combined unique and sort
		{
			name:     "Unique and sort",
			tag:      createSliceTestTag("unique,sort"),
			input:    &[]string{"c", "a", "b", "c", "a"},
			expected: &[]string{"a", "b", "c"},
		},

		// Compact tests
		{
			name:     "Compact strings",
			tag:      createSliceTestTag("compact"),
			input:    &[]string{"a", "", "b", "", "c"},
			expected: &[]string{"a", "b", "c"},
		},
		{
			name:     "Compact ints",
			tag:      createSliceTestTag("compact"),
			input:    &[]int{1, 0, 2, 0, 3},
			expected: &[]int{1, 2, 3},
		},

		// Limit tests
		{
			name:     "Limit slice smaller",
			tag:      createSliceTestTag("limit=2"),
			input:    &[]string{"a", "b", "c"},
			expected: &[]string{"a", "b"},
		},
		{
			name:     "Limit slice larger",
			tag:      createSliceTestTag("limit=5"),
			input:    &[]string{"a", "b", "c"},
			expected: &[]string{"a", "b", "c"},
		},

		// Custom formatter
		{
			name:     "Custom formatter",
			tag:      createSliceTestTag("custom"),
			input:    &[]string{"a", "b", "c"},
			expected: &[]string{"custom"},
		},
		// applySort
		{name: "Sort float64 asc", tag: createSliceTestTag("sort"), input: &[]float64{3.3, 1.1, 2.2}, expected: &[]float64{1.1, 2.2, 3.3}},
		{name: "Sort float64 desc", tag: createSliceTestTag("sort_desc"), input: &[]float64{3.3, 1.1, 2.2}, expected: &[]float64{3.3, 2.2, 1.1}},

		// applyCompact
		{name: "Compact empty slice", tag: createSliceTestTag("compact"), input: &[]string{}, expected: &[]string{}},

		// applyLimit
		{name: "Limit with negative value", tag: createSliceTestTag("limit=-1"), input: &[]string{"a", "b", "c"}, expected: &[]string{}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := f.Format(tc.tag, tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, tc.input)
		})
	}
}

func TestSlice_FormatReflectValue_Successfully(t *testing.T) {
	t.Parallel()

	f := NewSlice()

	tests := []struct {
		name     string
		tag      reflect.StructTag
		input    any
		expected any
	}{
		{
			name:     "Unique strings",
			tag:      createSliceTestTag("unique"),
			input:    &[]string{"a", "b", "a", "c", "b"},
			expected: &[]string{"a", "b", "c"},
		},
		{
			name:     "No tag",
			tag:      reflect.StructTag(""),
			input:    &[]string{"a", "b", "c"},
			expected: &[]string{"a", "b", "c"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val := reflect.ValueOf(tc.input)
			err := f.FormatReflectValue(tc.tag, val)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, tc.input)
		})
	}
}

func TestSlice_Format_Failure(t *testing.T) {
	t.Parallel()

	f := NewSlice()

	tests := []struct {
		name  string
		tag   reflect.StructTag
		input any
		errAs error
		errIs error
	}{
		{name: "Unsupported type", tag: createSliceTestTag("unique"), input: new(int), errIs: rerr.NotSupported},
		{name: "Not a slice pointer", tag: createSliceTestTag("unique"), input: "not a slice", errIs: rerr.NotSupported},
		{name: "Invalid limit value", tag: createSliceTestTag("limit=abc"), input: &[]string{"a", "b"}},
		{name: "Formatter not found", tag: createSliceTestTag("non_existent"), input: &[]string{}, errAs: &rerr.FormatterNotFoundError{}},
		{name: "Not a pointer", tag: createSliceTestTag("unique"), input: []string{}, errIs: rerr.NotSupported},
		{name: "Sort unsupported type", tag: createSliceTestTag("sort"), input: &[]bool{true, false}, errIs: rerr.NotSupported},
		{name: "Not a slice pointer but with tag", tag: createSliceTestTag("unique"), input: new(int), errIs: rerr.NotSupported},
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
