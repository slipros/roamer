// Package value provides utilities for type conversion and setting values in Go structs.
package value

import (
	"reflect"
)

// Pointer gets a pointer to the underlying value of a reflect.Value.
// This is useful when you need to pass a pointer to a field to a function
// that requires a pointer argument.
//
// The function handles both values that are already pointers and values
// that need to be converted to pointers. For non-pointer values, the function
// attempts to get the address of the value (which may fail if the value is
// not addressable).
//
// Parameters:
//   - value: The reflect.Value to get a pointer to.
//
// Returns:
//   - any: A pointer to the value, or nil if not possible.
//   - bool: Whether the operation was successful.
//
// Example usage (internal to the package):
//
//	// Get a pointer to a struct field
//	field := reflect.ValueOf(&myStruct).Elem().FieldByName("Name")
//	ptr, ok := Pointer(field)
//	if !ok {
//	    return errors.New("failed to get pointer to field")
//	}
//
//	// Now ptr is a *string that can be passed to a function
//	// expecting a string pointer
//	stringPtr := ptr.(*string)
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

	// Make sure we can convert to an interface{} value
	if !value.CanInterface() {
		return nil, false
	}

	// Return the pointer as an interface{} value
	return value.Interface(), true
}
