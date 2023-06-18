package value

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetInteger(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		var testStruct struct {
			S string
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int8(num))
			require.NoError(t, err)
			require.Equal(t, strconv.Itoa(num), testStruct.S)
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int16(num))
			require.NoError(t, err)
			require.Equal(t, strconv.Itoa(num), testStruct.S)
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int32(num))
			require.NoError(t, err)
			require.Equal(t, strconv.Itoa(num), testStruct.S)
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int64(num))
			require.NoError(t, err)
			require.Equal(t, strconv.Itoa(num), testStruct.S)
		}
	})

	t.Run("Bool", func(t *testing.T) {
		var testStruct struct {
			B bool
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			num := 1

			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int8(num))
			require.NoError(t, err)

			require.Equal(t, num > 0, testStruct.B)
		}

		for i := 0; i < v.NumField(); i++ {
			num := 0

			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int8(num))
			require.NoError(t, err)

			require.Equal(t, num > 0, testStruct.B)
		}

		for i := 0; i < v.NumField(); i++ {
			num := 1

			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int16(num))
			require.NoError(t, err)

			require.Equal(t, num > 0, testStruct.B)
		}

		for i := 0; i < v.NumField(); i++ {
			num := 0

			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int16(num))
			require.NoError(t, err)

			require.Equal(t, num > 0, testStruct.B)
		}

		for i := 0; i < v.NumField(); i++ {
			num := 1

			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int32(num))
			require.NoError(t, err)

			require.Equal(t, num > 0, testStruct.B)
		}

		for i := 0; i < v.NumField(); i++ {
			num := 0

			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int32(num))
			require.NoError(t, err)

			require.Equal(t, num > 0, testStruct.B)
		}

		for i := 0; i < v.NumField(); i++ {
			num := 1

			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int64(num))
			require.NoError(t, err)

			require.Equal(t, num > 0, testStruct.B)
		}

		for i := 0; i < v.NumField(); i++ {
			num := 0

			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int64(num))
			require.NoError(t, err)

			require.Equal(t, num > 0, testStruct.B)
		}
	})

	t.Run("Int", func(t *testing.T) {
		var testStruct struct {
			I   int
			I8  int8
			I16 int64
			I32 int32
			I64 int64
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int8(num))
			require.NoError(t, err)
			require.Equal(t, int64(num), fieldValue.Int())
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int16(num))
			require.NoError(t, err)
			require.Equal(t, int64(num), fieldValue.Int())
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int32(num))
			require.NoError(t, err)
			require.Equal(t, int64(num), fieldValue.Int())
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int64(num))
			require.NoError(t, err)
			require.Equal(t, int64(num), fieldValue.Int())
		}
	})

	t.Run("Uint", func(t *testing.T) {
		var testStruct struct {
			I   uint
			I8  uint8
			I16 uint64
			I32 uint32
			I64 uint64
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int8(num))
			require.NoError(t, err)
			require.Equal(t, uint64(num), fieldValue.Uint())
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int16(num))
			require.NoError(t, err)
			require.Equal(t, uint64(num), fieldValue.Uint())
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int32(num))
			require.NoError(t, err)
			require.Equal(t, uint64(num), fieldValue.Uint())
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int64(num))
			require.NoError(t, err)
			require.Equal(t, uint64(num), fieldValue.Uint())
		}
	})

	t.Run("Float", func(t *testing.T) {
		var testStruct struct {
			FL32 float32
			FL64 float64
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int8(num))
			require.NoError(t, err)
			require.Equal(t, float64(num), fieldValue.Float())
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int16(num))
			require.NoError(t, err)
			require.Equal(t, float64(num), fieldValue.Float())
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int32(num))
			require.NoError(t, err)
			require.Equal(t, float64(num), fieldValue.Float())
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int64(num))
			require.NoError(t, err)
			require.Equal(t, float64(num), fieldValue.Float())
		}
	})

	t.Run("Unsupported", func(t *testing.T) {
		var testStruct struct {
			SL []string
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int8(num))
			require.Error(t, err)
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int16(num))
			require.Error(t, err)
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int32(num))
			require.Error(t, err)
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetInteger(&fieldValue, int64(num))
			require.Error(t, err)
		}
	})
}
