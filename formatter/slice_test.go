package formatter

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createSliceTestTag(value string) reflect.StructTag {
	return reflect.StructTag(`slice:"` + value + `"`)
}

func TestSlice_Tag(t *testing.T) {
	f := NewSlice()
	assert.Equal(t, "slice", f.Tag())
}

func TestSlice_Format_Successfully(t *testing.T) {
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := f.Format(tc.tag, tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, tc.input)
		})
	}
}

func TestSlice_Format_Failure(t *testing.T) {
	f := NewSlice()

	tests := []struct {
		name    string
		tag     reflect.StructTag
		input   any
		wantErr string
	}{
		{name: "Unsupported type", tag: createSliceTestTag("unique"), input: new(int), wantErr: "not supported"},
		{name: "Not a slice pointer", tag: createSliceTestTag("unique"), input: "not a slice", wantErr: "not supported"},
		{name: "Invalid limit value", tag: createSliceTestTag("limit=abc"), input: &[]string{"a", "b"}, wantErr: "invalid limit value"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := f.Format(tc.tag, tc.input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
