package value

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

var str = "test_string"

type UnmarshallerText struct {
	S string
}

func (u *UnmarshallerText) UnmarshalText(text []byte) error {
	u.S = string(text)
	return nil
}

type UnmarshallerBinary struct {
	S string
}

func (u *UnmarshallerBinary) UnmarshalBinary(data []byte) error {
	u.S = string(data)
	return nil
}

func TestSetString(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		var testStruct struct {
			S string
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)
			require.Equal(t, str, testStruct.S)
		}
	})

	t.Run("Bool", func(t *testing.T) {
		str = "true"

		var testStruct struct {
			B bool
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			b, err := strconv.ParseBool(str)
			require.NoError(t, err)
			require.Equal(t, b, testStruct.B)
		}
	})

	t.Run("Int", func(t *testing.T) {
		str = "1"

		var testStruct struct {
			I int
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			parsed, err := strconv.ParseInt(str, 10, 0)
			require.NoError(t, err)

			require.Equal(t, int(parsed), testStruct.I)
		}
	})

	t.Run("Int8", func(t *testing.T) {
		str = "1"

		var testStruct struct {
			I int8
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			parsed, err := strconv.ParseInt(str, 10, 8)
			require.NoError(t, err)
			require.Equal(t, int8(parsed), testStruct.I)
		}
	})

	t.Run("Int16", func(t *testing.T) {
		str = "1"

		var testStruct struct {
			I int16
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			parsed, err := strconv.ParseInt(str, 10, 16)
			require.NoError(t, err)
			require.Equal(t, int16(parsed), testStruct.I)
		}
	})

	t.Run("Int32", func(t *testing.T) {
		str = "1"

		var testStruct struct {
			I int32
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			parsed, err := strconv.ParseInt(str, 10, 32)
			require.NoError(t, err)
			require.Equal(t, int32(parsed), testStruct.I)
		}
	})

	t.Run("Int64", func(t *testing.T) {
		str = "1"

		var testStruct struct {
			I int64
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			parsed, err := strconv.ParseInt(str, 10, 64)
			require.NoError(t, err)
			require.Equal(t, parsed, testStruct.I)
		}
	})

	t.Run("Uint", func(t *testing.T) {
		str = "1"

		var testStruct struct {
			I uint
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			parsed, err := strconv.ParseUint(str, 10, 0)
			require.NoError(t, err)

			require.Equal(t, uint(parsed), testStruct.I)
		}
	})

	t.Run("Uint8", func(t *testing.T) {
		str = "1"

		var testStruct struct {
			I uint8
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			parsed, err := strconv.ParseUint(str, 10, 8)
			require.NoError(t, err)
			require.Equal(t, uint8(parsed), testStruct.I)
		}
	})

	t.Run("Uint16", func(t *testing.T) {
		str = "1"

		var testStruct struct {
			I uint16
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			parsed, err := strconv.ParseUint(str, 10, 16)
			require.NoError(t, err)
			require.Equal(t, uint16(parsed), testStruct.I)
		}
	})

	t.Run("Uint32", func(t *testing.T) {
		str = "1"

		var testStruct struct {
			I uint32
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			parsed, err := strconv.ParseUint(str, 10, 32)
			require.NoError(t, err)
			require.Equal(t, uint32(parsed), testStruct.I)
		}
	})

	t.Run("Uint64", func(t *testing.T) {
		str = "1"

		var testStruct struct {
			I uint64
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			parsed, err := strconv.ParseUint(str, 10, 64)
			require.NoError(t, err)
			require.Equal(t, parsed, testStruct.I)
		}
	})

	t.Run("Float32", func(t *testing.T) {
		str = "1"

		var testStruct struct {
			F float32
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			parsed, err := strconv.ParseFloat(str, 32)
			require.NoError(t, err)

			require.Equal(t, float32(parsed), testStruct.F)
		}
	})

	t.Run("Float64", func(t *testing.T) {
		str = "1"

		var testStruct struct {
			F float64
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			parsed, err := strconv.ParseFloat(str, 64)
			require.NoError(t, err)

			require.Equal(t, parsed, testStruct.F)
		}
	})

	t.Run("Complex64", func(t *testing.T) {
		str = "1"

		var testStruct struct {
			C complex64
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			parsed, err := strconv.ParseComplex(str, 64)
			require.NoError(t, err)

			require.Equal(t, complex64(parsed), testStruct.C)
		}
	})

	t.Run("Complex128", func(t *testing.T) {
		str = "1"

		var testStruct struct {
			C complex128
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			parsed, err := strconv.ParseComplex(str, 128)
			require.NoError(t, err)

			require.Equal(t, parsed, testStruct.C)
		}
	})

	t.Run("Slice strings", func(t *testing.T) {
		var testStruct struct {
			SL []string
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			require.Equal(t, str, testStruct.SL[0])
		}
	})

	t.Run("Slice uint8", func(t *testing.T) {
		var testStruct struct {
			SL []uint8
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			require.Equal(t, []byte(str), testStruct.SL)
		}
	})

	t.Run("Unmarshaller text ", func(t *testing.T) {
		var testStruct struct {
			U UnmarshallerText
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			require.Equal(t, str, testStruct.U.S)
		}
	})

	t.Run("Unmarshaller binary ", func(t *testing.T) {
		var testStruct struct {
			U UnmarshallerBinary
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.NoError(t, err)

			require.Equal(t, str, testStruct.U.S)
		}
	})

	t.Run("Unsupported", func(t *testing.T) {
		var testStruct struct {
			M map[string]string
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetString(&fieldValue, str)
			require.Error(t, err)
		}
	})
}
