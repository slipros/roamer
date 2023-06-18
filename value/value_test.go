package value

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		var testStruct struct {
			S string
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, str)
			require.NoError(t, err)
			require.Equal(t, str, testStruct.S)
		}

		var testStructP struct {
			SP *string
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, str)
			require.NoError(t, err)
			require.Equal(t, str, testStruct.S)
		}
	})

	t.Run("Int", func(t *testing.T) {
		var testStruct struct {
			I int
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, num)
			require.NoError(t, err)
			require.Equal(t, num, testStruct.I)
		}

		var testStructP struct {
			I *int
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)

			n := num
			err := Set(&fieldValue, &n)
			require.NoError(t, err)
			require.Equal(t, n, *testStructP.I)
		}
	})

	t.Run("Int8", func(t *testing.T) {
		var testStruct struct {
			I int8
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, int8(num))
			require.NoError(t, err)
			require.Equal(t, int8(num), testStruct.I)
		}

		var testStructP struct {
			I *int8
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)

			n := int8(num)
			err := Set(&fieldValue, &n)
			require.NoError(t, err)
			require.Equal(t, n, *testStructP.I)
		}
	})

	t.Run("Int16", func(t *testing.T) {
		var testStruct struct {
			I int16
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, int16(num))
			require.NoError(t, err)
			require.Equal(t, int16(num), testStruct.I)
		}

		var testStructP struct {
			I *int16
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)

			n := int16(num)
			err := Set(&fieldValue, n)
			require.NoError(t, err)
			require.Equal(t, n, *testStructP.I)
		}
	})

	t.Run("Int32", func(t *testing.T) {
		var testStruct struct {
			I int32
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, int32(num))
			require.NoError(t, err)
			require.Equal(t, int32(num), testStruct.I)
		}

		var testStructP struct {
			I *int32
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)

			n := int32(num)
			err := Set(&fieldValue, n)
			require.NoError(t, err)
			require.Equal(t, n, *testStructP.I)
		}
	})

	t.Run("Int64", func(t *testing.T) {
		var testStruct struct {
			I int64
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, int64(num))
			require.NoError(t, err)
			require.Equal(t, int64(num), testStruct.I)
		}

		var testStructP struct {
			I *int64
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)

			n := int64(num)
			err := Set(&fieldValue, n)
			require.NoError(t, err)
			require.Equal(t, n, *testStructP.I)
		}
	})

	t.Run("Uint", func(t *testing.T) {
		var testStruct struct {
			I uint
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, num)
			require.NoError(t, err)
			require.Equal(t, uint(num), testStruct.I)
		}

		var testStructP struct {
			I *uint
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)

			n := uint(num)
			err := Set(&fieldValue, &n)
			require.NoError(t, err)
			require.Equal(t, n, *testStructP.I)
		}
	})

	t.Run("Uint8", func(t *testing.T) {
		var testStruct struct {
			I uint8
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, uint8(num))
			require.NoError(t, err)
			require.Equal(t, uint8(num), testStruct.I)
		}

		var testStructP struct {
			I *uint8
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)

			n := uint8(num)
			err := Set(&fieldValue, &n)
			require.NoError(t, err)
			require.Equal(t, n, *testStructP.I)
		}
	})

	t.Run("Uint16", func(t *testing.T) {
		var testStruct struct {
			I uint16
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, uint16(num))
			require.NoError(t, err)
			require.Equal(t, uint16(num), testStruct.I)
		}

		var testStructP struct {
			I *uint16
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)

			n := uint16(num)
			err := Set(&fieldValue, n)
			require.NoError(t, err)
			require.Equal(t, n, *testStructP.I)
		}
	})

	t.Run("Uint32", func(t *testing.T) {
		var testStruct struct {
			I uint32
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, uint32(num))
			require.NoError(t, err)
			require.Equal(t, uint32(num), testStruct.I)
		}

		var testStructP struct {
			I *uint32
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)

			n := uint32(num)
			err := Set(&fieldValue, n)
			require.NoError(t, err)
			require.Equal(t, n, *testStructP.I)
		}
	})

	t.Run("Uint64", func(t *testing.T) {
		var testStruct struct {
			I uint64
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, uint64(num))
			require.NoError(t, err)
			require.Equal(t, uint64(num), testStruct.I)
		}

		var testStructP struct {
			I *uint64
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)

			n := uint64(num)
			err := Set(&fieldValue, n)
			require.NoError(t, err)
			require.Equal(t, n, *testStructP.I)
		}
	})

	t.Run("Float32", func(t *testing.T) {
		var testStruct struct {
			F float32
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, float32(num))
			require.NoError(t, err)
			require.Equal(t, float32(num), testStruct.F)
		}

		var testStructP struct {
			F *float32
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)

			n := float32(num)
			err := Set(&fieldValue, n)
			require.NoError(t, err)
			require.Equal(t, n, *testStructP.F)
		}
	})

	t.Run("Float64", func(t *testing.T) {
		var testStruct struct {
			F float64
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, float64(num))
			require.NoError(t, err)
			require.Equal(t, float64(num), testStruct.F)
		}

		var testStructP struct {
			F *float64
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)

			n := float64(num)
			err := Set(&fieldValue, n)
			require.NoError(t, err)
			require.Equal(t, n, *testStructP.F)
		}
	})

	t.Run("Slice string", func(t *testing.T) {
		var testStruct struct {
			SL []string
		}

		sl := []string{str, str}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, sl)
			require.NoError(t, err)
			require.Equal(t, sl, testStruct.SL)
		}
	})

	t.Run("Unsupported", func(t *testing.T) {
		var testStruct struct {
			M map[string]string
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		m := map[string]string{
			str: str,
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, m)
			require.Error(t, err)
		}
	})
}
