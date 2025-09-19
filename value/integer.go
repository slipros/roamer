package value

import (
	"math"
	"reflect"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/exp/constraints"

	rerr "github.com/slipros/roamer/err"
)

// SetInteger converts an integer value to the appropriate type for a target field.
// Handles conversion to strings, booleans, numeric types, and interfaces.
// Performs range checking to prevent overflow.
//
// Parameters:
//   - field: Target field to set (reflect.Value).
//   - number: Integer value to convert and set.
//
// Returns:
//   - error: If conversion or assignment fails.
func SetInteger[I constraints.Integer](field reflect.Value, number I) error {
	// Check if the field is settable
	if !field.CanSet() {
		return errors.Errorf("field of type %s is not settable", field.Type())
	}

	// Determine if the number is signed or unsigned
	var (
		isSigned  bool
		int64Val  int64
		uint64Val uint64
	)

	// Convert to int64 or uint64 for standard processing
	switch any(number).(type) {
	case int, int8, int16, int32, int64:
		isSigned = true
		int64Val = int64(number)
	default:
		// Must be an unsigned integer
		uint64Val = uint64(number)
	}

	switch field.Kind() {
	case reflect.String:
		// Convert integer to string using appropriate formatting based on type
		if isSigned {
			field.SetString(strconv.FormatInt(int64Val, 10))
		} else {
			field.SetString(strconv.FormatUint(uint64Val, 10))
		}
		return nil

	case reflect.Bool:
		// Convert integer to bool (true only if > 0, false otherwise)
		if isSigned {
			field.SetBool(int64Val > 0)
		} else {
			field.SetBool(uint64Val > 0)
		}
		return nil

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		if isSigned {
			// For signed input to signed field, check range
			if err := checkSignedIntegerRange(field, int64Val); err != nil {
				return err
			}
			field.SetInt(int64Val)
		} else {
			// For unsigned input to signed field, check overflow
			if uint64Val > math.MaxInt64 {
				return errors.Errorf("value %v overflows target type %s", uint64Val, field.Type())
			}
			// Then check range as a signed value
			if err := checkSignedIntegerRange(field, int64(uint64Val)); err != nil {
				return err
			}
			field.SetInt(int64(uint64Val))
		}
		return nil

	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		// For unsigned types, ensure signed values are not negative
		if isSigned && int64Val < 0 {
			return errors.Errorf("cannot set negative value %v to unsigned type %s", int64Val, field.Type())
		}

		// Calculate the value to set
		var valueToSet uint64
		if isSigned {
			valueToSet = uint64(int64Val)
		} else {
			valueToSet = uint64Val
		}

		// Check for range
		if err := checkUnsignedIntegerRange(field, valueToSet); err != nil {
			return err
		}

		field.SetUint(valueToSet)
		return nil

	case reflect.Float32, reflect.Float64:
		// Convert integer to float
		if isSigned {
			field.SetFloat(float64(int64Val))
		} else {
			field.SetFloat(float64(uint64Val))
		}
		return nil

	case reflect.Complex64, reflect.Complex128:
		// Set integer value to the real part, imaginary part is 0
		if isSigned {
			field.SetComplex(complex(float64(int64Val), 0))
		} else {
			field.SetComplex(complex(float64(uint64Val), 0))
		}
		return nil

	case reflect.Interface:
		// For any fields, just set the integer value
		if isSigned {
			field.Set(reflect.ValueOf(int64Val))
		} else {
			field.Set(reflect.ValueOf(uint64Val))
		}
		return nil

	case reflect.Ptr:
		// For pointer fields, check if valid and then dereference and call recursively
		if field.IsNil() {
			// Initialize nil pointers
			field.Set(reflect.New(field.Type().Elem()))
		}
		return SetInteger[I](field.Elem(), number)
	}

	// If the field doesn't match any of the above types, return an error with more details
	var valueStr string
	if isSigned {
		valueStr = strconv.FormatInt(int64Val, 10)
	} else {
		valueStr = strconv.FormatUint(uint64Val, 10)
	}

	return errors.Wrapf(rerr.NotSupported, "cannot set integer value %s to field of type %s",
		valueStr, field.Type())
}

// checkSignedIntegerRange verifies a signed integer value is within range for the target field.
// Prevents data loss or overflow from type conversions.
func checkSignedIntegerRange(field reflect.Value, number int64) error {
	// Store field kind to avoid multiple calls to field.Kind()
	kind := field.Kind()

	// Quick exit if we're not dealing with a signed integer type
	if kind < reflect.Int8 || (kind > reflect.Int64 && kind != reflect.Int) {
		return nil
	}

	// For signed values, check against the min/max range of the target type
	bitSize := field.Type().Bits()
	maxVal := int64(1)<<(bitSize-1) - 1
	minVal := -int64(1) << (bitSize - 1)

	if number > maxVal || number < minVal {
		return errors.Errorf("value %v is outside the range of target type %s [%d, %d]",
			number, field.Type(), minVal, maxVal)
	}

	return nil
}

// checkUnsignedIntegerRange verifies an unsigned integer value is within range for the target field.
// Checks for potential overflow based on bit size.
func checkUnsignedIntegerRange(field reflect.Value, number uint64) error {
	// For unsigned types, check for potential overflow
	bitSize := field.Type().Bits()
	maxVal := uint64(1)<<bitSize - 1

	if number > maxVal {
		return errors.Errorf("value %v overflows target type %s (max: %d)",
			number, field.Type(), maxVal)
	}

	return nil
}
