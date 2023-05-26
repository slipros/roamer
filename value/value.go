package value

import (
	"reflect"

	roamerError "github.com/SLIpros/roamer/error"
)

// Set set value to field.
func Set(field *reflect.Value, value any) error {
	if field.Kind() == reflect.Pointer && field.IsNil() {
		// init ptr
		field.Set(reflect.New(field.Type().Elem()))
		*field = reflect.Indirect(*field)
	}

	switch t := value.(type) {
	case string:
		return SetString(field, t)
	case *string:
		return SetString(field, *t)
	case int:
		return SetInteger(field, t)
	case *int:
		return SetInteger(field, *t)
	case int8:
		return SetInteger(field, t)
	case *int8:
		return SetInteger(field, *t)
	case int16:
		return SetInteger(field, t)
	case *int16:
		return SetInteger(field, *t)
	case int32:
		return SetInteger(field, t)
	case *int32:
		return SetInteger(field, *t)
	case int64:
		return SetInteger(field, t)
	case *int64:
		return SetInteger(field, *t)
	case uint:
		return SetInteger(field, t)
	case *uint:
		return SetInteger(field, *t)
	case uint8:
		return SetInteger(field, t)
	case *uint8:
		return SetInteger(field, *t)
	case uint16:
		return SetInteger(field, t)
	case *uint16:
		return SetInteger(field, *t)
	case uint32:
		return SetInteger(field, t)
	case *uint32:
		return SetInteger(field, *t)
	case uint64:
		return SetInteger(field, t)
	case *uint64:
		return SetInteger(field, *t)
	case float32:
		return SetFloat(field, t)
	case *float32:
		return SetFloat(field, *t)
	case float64:
		return SetFloat(field, t)
	case *float64:
		return SetFloat(field, *t)
	case []string:
		return SetSliceString(field, t)
	}

	valueType := reflect.TypeOf(value)
	if valueType.Kind() == reflect.Pointer {
		// deref ptr
		valueType = valueType.Elem()
	}

	if field.Type().AssignableTo(valueType) {
		field.Set(reflect.Indirect(reflect.ValueOf(value)))
	}

	return roamerError.ErrNotSupported
}
