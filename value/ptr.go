// Package value provides utilities for type conversion and setting values in Go structs.
package value

import (
	"reflect"
)

// Pointer gets a pointer to the underlying value of a reflect.Value.
// Handles both existing pointers and non-pointer values that need addressing.
//
// Parameters:
//   - value: The reflect.Value to get a pointer to.
//
// Returns:
//   - any: A pointer to the value, or nil if not possible.
//   - bool: Whether the operation was successful.
func Pointer(value reflect.Value) (any, bool) {
	switch value.Kind() {
	case reflect.Ptr:
		// If already a pointer, just return the interface if not nil
		if value.IsNil() {
			return nil, false
		}
	default:
		// If not a pointer, try to get the address
		if value.Kind() != reflect.Ptr {
			if !value.CanAddr() {
				// Value is not addressable (e.g., the result of a function call)
				return nil, false
			}

			// Get the address of the value
			value = value.Addr()
		}
	}

	// Make sure we can convert to an any value
	if !value.CanInterface() {
		return nil, false
	}

	// Return the pointer as an any value
	return value.Interface(), true
}
