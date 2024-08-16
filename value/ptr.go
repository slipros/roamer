package value

import (
	"reflect"
)

// Pointer returns pointer to value.
func Pointer(value reflect.Value) (any, bool) {
	switch value.Kind() {
	case reflect.Ptr:
		if value.IsNil() {
			return nil, false
		}
	default:
		if value.Kind() != reflect.Ptr {
			if !value.CanAddr() {
				return nil, false
			}

			value = value.Addr()
		}
	}

	if !value.CanInterface() {
		return nil, false
	}

	return value.Interface(), true
}
