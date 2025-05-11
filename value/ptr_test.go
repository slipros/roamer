package value

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPointer_Successfully tests successful operations of the Pointer function
func TestPointer_Successfully(t *testing.T) {
	tests := []struct {
		name        string
		setupValue  func() reflect.Value
		checkResult func(t *testing.T, result any, ok bool)
	}{
		{
			name: "Get pointer to a string field in struct",
			setupValue: func() reflect.Value {
				type TestStruct struct {
					Name string
				}
				s := TestStruct{Name: "test"}
				// Get field "Name" in the struct
				return reflect.ValueOf(&s).Elem().FieldByName("Name")
			},
			checkResult: func(t *testing.T, result any, ok bool) {
				// Check that we got a valid pointer
				assert.True(t, ok, "Should successfully get a pointer")
				assert.NotNil(t, result, "Result should not be nil")

				// Check that it's a string pointer
				strPtr, ok := result.(*string)
				assert.True(t, ok, "Result should be a *string")
				assert.Equal(t, "test", *strPtr, "Pointed value should be 'test'")

				// Check that modifying through the pointer works
				*strPtr = "modified"
				assert.Equal(t, "modified", *strPtr, "Should be able to modify the value through the pointer")
			},
		},
		{
			name: "Get pointer to an int field in struct",
			setupValue: func() reflect.Value {
				type TestStruct struct {
					Count int
				}
				s := TestStruct{Count: 42}
				// Get field "Count" in the struct
				return reflect.ValueOf(&s).Elem().FieldByName("Count")
			},
			checkResult: func(t *testing.T, result any, ok bool) {
				// Check that we got a valid pointer
				assert.True(t, ok, "Should successfully get a pointer")
				assert.NotNil(t, result, "Result should not be nil")

				// Check that it's an int pointer
				intPtr, ok := result.(*int)
				assert.True(t, ok, "Result should be an *int")
				assert.Equal(t, 42, *intPtr, "Pointed value should be 42")

				// Check that modifying through the pointer works
				*intPtr = 100
				assert.Equal(t, 100, *intPtr, "Should be able to modify the value through the pointer")
			},
		},
		{
			name: "Get pointer to a slice element",
			setupValue: func() reflect.Value {
				slice := []string{"one", "two", "three"}
				// Get the first element of the slice
				return reflect.ValueOf(slice).Index(0)
			},
			checkResult: func(t *testing.T, result any, ok bool) {
				// Check that we got a valid pointer
				assert.True(t, ok, "Should successfully get a pointer")
				assert.NotNil(t, result, "Result should not be nil")

				// Check that it's a string pointer
				strPtr, ok := result.(*string)
				assert.True(t, ok, "Result should be a *string")
				assert.Equal(t, "one", *strPtr, "Pointed value should be 'one'")

				// Check that modifying through the pointer works
				*strPtr = "modified"
				assert.Equal(t, "modified", *strPtr, "Should be able to modify the value through the pointer")
			},
		},
		{
			name: "Get pointer to a map value",
			setupValue: func() reflect.Value {
				m := map[string]int{"key": 123}
				// Create a pointer to the map
				mapPtr := &m
				// Get the value for "key" through pointer indirection
				return reflect.ValueOf(mapPtr).Elem().MapIndex(reflect.ValueOf("key"))
			},
			checkResult: func(t *testing.T, result any, ok bool) {
				// For maps, values are not addressable in Go, so we should get false
				assert.False(t, ok, "Map values are not addressable")
				assert.Nil(t, result, "Result should be nil for non-addressable values")
			},
		},
		{
			name: "Existing pointer value",
			setupValue: func() reflect.Value {
				s := "test string"
				ptr := &s
				return reflect.ValueOf(ptr)
			},
			checkResult: func(t *testing.T, result any, ok bool) {
				assert.True(t, ok, "Should successfully handle an existing pointer")
				assert.NotNil(t, result, "Result should not be nil")

				strPtr, ok := result.(*string)
				assert.True(t, ok, "Result should be a *string")
				assert.Equal(t, "test string", *strPtr, "Pointed value should be 'test string'")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup the value to test
			value := tc.setupValue()

			// Call the function being tested
			result, ok := Pointer(value)

			// Check the results
			tc.checkResult(t, result, ok)
		})
	}
}

// TestPointer_Failure tests failure scenarios of the Pointer function
func TestPointer_Failure(t *testing.T) {
	tests := []struct {
		name       string
		setupValue func() reflect.Value
	}{
		{
			name: "Nil pointer",
			setupValue: func() reflect.Value {
				var ptr *string = nil
				return reflect.ValueOf(ptr)
			},
		},
		{
			name: "Non-addressable value from function call",
			setupValue: func() reflect.Value {
				// This creates a non-addressable string value
				return reflect.ValueOf("test string")
			},
		},
		{
			name: "Non-addressable value from map lookup",
			setupValue: func() reflect.Value {
				m := map[string]int{"key": 123}
				return reflect.ValueOf(m).MapIndex(reflect.ValueOf("key"))
			},
		},
		{
			name: "Value from unexported field",
			setupValue: func() reflect.Value {
				type TestStruct struct {
					name string // Unexported field
				}
				s := TestStruct{name: "test"}
				return reflect.ValueOf(&s).Elem().FieldByName("name")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup the value to test
			value := tc.setupValue()

			// Call the function being tested
			result, ok := Pointer(value)

			// All these cases should fail
			assert.False(t, ok, "Should return false for failure cases")
			assert.Nil(t, result, "Result should be nil for failure cases")
		})
	}
}
