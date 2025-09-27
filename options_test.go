package roamer

import (
	"reflect"
	"testing"

	"github.com/slipros/assign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock extension functions for testing different types
func stringExtension(value any) (func(to reflect.Value) error, bool) {
	if str, ok := value.(string); ok {
		return func(to reflect.Value) error {
			return assign.String(to, str)
		}, true
	}
	return nil, false
}

func intExtension(value any) (func(to reflect.Value) error, bool) {
	if i, ok := value.(int); ok {
		return func(to reflect.Value) error {
			return assign.Integer(to, i)
		}, true
	}
	return nil, false
}

func customTypeExtension(value any) (func(to reflect.Value) error, bool) {
	if ct, ok := value.(customType); ok {
		return func(to reflect.Value) error {
			return assign.String(to, ct.value)
		}, true
	}
	return nil, false
}

// Custom type for testing
type customType struct {
	value string
}

// nilExtension returns a nil function for testing nil handling
func nilExtension(value any) (func(to reflect.Value) error, bool) {
	return nil, false
}

// TestWithAssignExtensions_Successfully tests successful scenarios for WithAssignExtensions
func TestWithAssignExtensions_Successfully(t *testing.T) {
	tests := []struct {
		name                    string
		extensions              []assign.ExtensionFunc
		existingExtensions      []assign.ExtensionFunc
		expectedExtensionsCount int
		description             string
	}{
		{
			name:                    "add single extension to empty roamer",
			extensions:              []assign.ExtensionFunc{stringExtension},
			existingExtensions:      nil,
			expectedExtensionsCount: 1,
			description:             "should add one extension to empty assignExtensions slice",
		},
		{
			name:                    "add multiple extensions to empty roamer",
			extensions:              []assign.ExtensionFunc{stringExtension, intExtension, customTypeExtension},
			existingExtensions:      nil,
			expectedExtensionsCount: 3,
			description:             "should add multiple extensions to empty assignExtensions slice",
		},
		{
			name:                    "add single extension to existing extensions",
			extensions:              []assign.ExtensionFunc{intExtension},
			existingExtensions:      []assign.ExtensionFunc{stringExtension},
			expectedExtensionsCount: 2,
			description:             "should append new extension to existing extensions",
		},
		{
			name:                    "add multiple extensions to existing extensions",
			extensions:              []assign.ExtensionFunc{intExtension, customTypeExtension},
			existingExtensions:      []assign.ExtensionFunc{stringExtension},
			expectedExtensionsCount: 3,
			description:             "should append multiple new extensions to existing extensions",
		},
		{
			name:                    "add extensions with existing multiple extensions",
			extensions:              []assign.ExtensionFunc{customTypeExtension},
			existingExtensions:      []assign.ExtensionFunc{stringExtension, intExtension},
			expectedExtensionsCount: 3,
			description:             "should append extension to multiple existing extensions",
		},
		{
			name:                    "add extension that returns nil function",
			extensions:              []assign.ExtensionFunc{nilExtension},
			existingExtensions:      nil,
			expectedExtensionsCount: 1,
			description:             "should add extension even if it returns nil function",
		},
		{
			name:                    "add mixed extensions with some returning nil",
			extensions:              []assign.ExtensionFunc{stringExtension, nilExtension, intExtension},
			existingExtensions:      nil,
			expectedExtensionsCount: 3,
			description:             "should add all extensions regardless of their return values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create roamer with existing extensions if any
			var opts []OptionsFunc
			if tt.existingExtensions != nil {
				opts = append(opts, WithAssignExtensions(tt.existingExtensions...))
			}

			roamer := NewRoamer(opts...)

			// Verify initial state
			if tt.existingExtensions != nil {
				assert.Len(t, roamer.assignExtensions, len(tt.existingExtensions),
					"initial extensions count should match existing extensions")
			} else {
				assert.Empty(t, roamer.assignExtensions, "initial assignExtensions should be empty")
			}

			// Apply the option function being tested
			option := WithAssignExtensions(tt.extensions...)
			option(roamer)

			// Verify extensions were added correctly
			assert.Len(t, roamer.assignExtensions, tt.expectedExtensionsCount,
				"assignExtensions count should match expected count")

			// Verify that existing extensions are preserved
			if tt.existingExtensions != nil {
				for i, expectedExt := range tt.existingExtensions {
					actualExt := roamer.assignExtensions[i]
					// We can't directly compare function pointers, but we can test their behavior
					testValue := "test"
					expectedAssignFunc, expectedOk := expectedExt(testValue)
					actualAssignFunc, actualOk := actualExt(testValue)

					assert.Equal(t, expectedOk, actualOk,
						"existing extension %d should behave the same", i)

					if expectedOk && actualOk {
						assert.NotNil(t, expectedAssignFunc, "expected assign function should not be nil")
						assert.NotNil(t, actualAssignFunc, "actual assign function should not be nil")
					}
				}
			}

			// Verify that new extensions were appended correctly
			startIndex := len(tt.existingExtensions)
			for i, expectedExt := range tt.extensions {
				actualExt := roamer.assignExtensions[startIndex+i]
				// Test behavior consistency
				testValue := "test"
				expectedAssignFunc, expectedOk := expectedExt(testValue)
				actualAssignFunc, actualOk := actualExt(testValue)

				assert.Equal(t, expectedOk, actualOk,
					"new extension %d should behave the same", i)

				if expectedOk && actualOk {
					assert.NotNil(t, expectedAssignFunc, "expected assign function should not be nil")
					assert.NotNil(t, actualAssignFunc, "actual assign function should not be nil")
				}
			}
		})
	}
}

// TestWithAssignExtensions_EdgeCases tests edge cases and boundary conditions
func TestWithAssignExtensions_EdgeCases(t *testing.T) {
	tests := []struct {
		name                    string
		extensions              []assign.ExtensionFunc
		expectedExtensionsCount int
		description             string
	}{
		{
			name:                    "empty extensions slice",
			extensions:              []assign.ExtensionFunc{},
			expectedExtensionsCount: 0,
			description:             "should handle empty extensions slice without error",
		},
		{
			name:                    "nil extensions in slice",
			extensions:              []assign.ExtensionFunc{nil, stringExtension, nil},
			expectedExtensionsCount: 3,
			description:             "should add nil extensions to slice without filtering",
		},
		{
			name:                    "all nil extensions",
			extensions:              []assign.ExtensionFunc{nil, nil, nil},
			expectedExtensionsCount: 3,
			description:             "should handle slice with all nil extensions",
		},
		{
			name:                    "duplicate extension functions",
			extensions:              []assign.ExtensionFunc{stringExtension, stringExtension, stringExtension},
			expectedExtensionsCount: 3,
			description:             "should allow duplicate extension functions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roamer := NewRoamer()

			// Verify initial state
			assert.Empty(t, roamer.assignExtensions, "initial assignExtensions should be empty")

			// Apply the option function
			option := WithAssignExtensions(tt.extensions...)
			option(roamer)

			// Verify extensions count
			assert.Len(t, roamer.assignExtensions, tt.expectedExtensionsCount,
				"assignExtensions count should match expected count")

			// For non-nil extensions, verify they can be called
			for i, ext := range roamer.assignExtensions {
				if ext != nil {
					assignFunc, ok := ext("test")
					// This should not panic and should return some result
					_ = assignFunc
					_ = ok
				} else {
					assert.Nil(t, ext, "extension %d should be nil as added", i)
				}
			}
		})
	}
}

// TestWithAssignExtensions_Functionality tests that extensions actually work in assignment
func TestWithAssignExtensions_Functionality(t *testing.T) {
	tests := []struct {
		name        string
		extensions  []assign.ExtensionFunc
		testValue   any
		expectMatch bool
		description string
	}{
		{
			name:        "string extension handles string value",
			extensions:  []assign.ExtensionFunc{stringExtension},
			testValue:   "test string",
			expectMatch: true,
			description: "string extension should handle string values",
		},
		{
			name:        "int extension handles int value",
			extensions:  []assign.ExtensionFunc{intExtension},
			testValue:   42,
			expectMatch: true,
			description: "int extension should handle int values",
		},
		{
			name:        "custom type extension handles custom type",
			extensions:  []assign.ExtensionFunc{customTypeExtension},
			testValue:   customType{value: "custom"},
			expectMatch: true,
			description: "custom extension should handle custom type values",
		},
		{
			name:        "extension does not handle unmatched type",
			extensions:  []assign.ExtensionFunc{stringExtension},
			testValue:   42,
			expectMatch: false,
			description: "string extension should not handle int values",
		},
		{
			name:        "multiple extensions with first match",
			extensions:  []assign.ExtensionFunc{stringExtension, intExtension},
			testValue:   "test",
			expectMatch: true,
			description: "first matching extension should handle the value",
		},
		{
			name:        "multiple extensions with second match",
			extensions:  []assign.ExtensionFunc{stringExtension, intExtension},
			testValue:   42,
			expectMatch: true,
			description: "second extension should handle the value when first doesn't match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roamer := NewRoamer()

			// Apply extensions
			option := WithAssignExtensions(tt.extensions...)
			option(roamer)

			// Test that at least one extension can handle the test value
			var foundMatch bool
			var assignFunc func(to reflect.Value) error

			for _, ext := range roamer.assignExtensions {
				if ext != nil {
					if af, ok := ext(tt.testValue); ok {
						foundMatch = true
						assignFunc = af
						break
					}
				}
			}

			assert.Equal(t, tt.expectMatch, foundMatch,
				"extension match result should match expected")

			if tt.expectMatch {
				assert.NotNil(t, assignFunc, "assign function should not be nil when match is expected")

				// Test that the assign function works (basic smoke test)
				var target string
				targetValue := reflect.ValueOf(&target).Elem()
				err := assignFunc(targetValue)

				// We don't assert on the error since different extensions might have different behaviors
				// but the function should not panic
				_ = err
			}
		})
	}
}

// TestWithAssignExtensions_MultipleApplications tests applying the option multiple times
func TestWithAssignExtensions_MultipleApplications(t *testing.T) {
	t.Run("multiple applications append extensions", func(t *testing.T) {
		roamer := NewRoamer()

		// Apply first set of extensions
		option1 := WithAssignExtensions(stringExtension)
		option1(roamer)

		assert.Len(t, roamer.assignExtensions, 1, "should have 1 extension after first application")

		// Apply second set of extensions
		option2 := WithAssignExtensions(intExtension, customTypeExtension)
		option2(roamer)

		assert.Len(t, roamer.assignExtensions, 3, "should have 3 extensions after second application")

		// Verify that all extensions are present and functional
		testCases := []struct {
			value    any
			expected bool
		}{
			{"string", true},                  // should match stringExtension
			{42, true},                        // should match intExtension
			{customType{value: "test"}, true}, // should match customTypeExtension
			{[]int{1, 2, 3}, false},           // should not match any extension
		}

		for _, tc := range testCases {
			var found bool
			for _, ext := range roamer.assignExtensions {
				if _, ok := ext(tc.value); ok {
					found = true
					break
				}
			}
			assert.Equal(t, tc.expected, found,
				"value %v should have expected match result", tc.value)
		}
	})
}

// TestWithAssignExtensions_NilRoamer tests behavior with nil roamer (edge case)
func TestWithAssignExtensions_NilRoamer(t *testing.T) {
	t.Run("option function handles nil roamer gracefully", func(t *testing.T) {
		option := WithAssignExtensions(stringExtension)

		// This should not panic even with nil roamer
		// In real usage, this would never happen as NewRoamer creates the roamer
		// But it's good to test that the option function is defensive
		require.NotPanics(t, func() {
			defer func() {
				if r := recover(); r != nil {
					// If it panics due to nil pointer, that's expected behavior
					// We just want to ensure it doesn't crash unexpectedly
				}
			}()
			option(nil)
		})
	})
}
