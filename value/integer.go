package value

import (
	"reflect"
	"strconv"

	"golang.org/x/exp/constraints"

	roamerError "github.com/SLIpros/roamer/error"
)

// SetInteger set integer number to field.
func SetInteger[I constraints.Integer](field *reflect.Value, number I) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(strconv.Itoa(int(number)))
		return nil
	case reflect.Bool:
		field.SetBool(number > 0)
		return nil
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		field.SetInt(int64(number))
		return nil
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		field.SetUint(uint64(number))
		return nil
	case reflect.Float32, reflect.Float64:
		field.SetFloat(float64(number))
		return nil
	case reflect.Interface:
		field.Set(reflect.ValueOf(number))
		return nil
	}

	return roamerError.ErrNotSupported
}
