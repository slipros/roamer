package value

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/constraints"
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
			S *string
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(&fieldValue, str)
			require.NoError(t, err)
			require.Equal(t, str, *testStructP.S)
		}
	})

	t.Run("Int", func(t *testing.T) {
		t.Parallel()

		testSetInt(t, num)
		testSetInt(t, int8(num))
		testSetInt(t, int16(num))
		testSetInt(t, int32(num))
		testSetInt(t, int64(num))
	})

	t.Run("Uint", func(t *testing.T) {
		t.Parallel()

		testSetUint(t, uint(num))
		testSetUint(t, uint8(num))
		testSetUint(t, uint16(num))
		testSetUint(t, uint32(num))
		testSetUint(t, uint64(num))
	})

	t.Run("Float", func(t *testing.T) {
		t.Parallel()

		testSetFloat(t, float32(num))
		testSetFloat(t, float64(num))
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

func testSetInt[T constraints.Integer](t *testing.T, integer T) {
	var testStruct struct {
		I T
	}

	v := reflect.Indirect(reflect.ValueOf(&testStruct))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		err := Set(&fieldValue, integer)
		require.NoError(t, err)
		require.Equal(t, integer, testStruct.I)
	}

	var testStructP struct {
		I *T
	}

	v = reflect.Indirect(reflect.ValueOf(&testStructP))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		err := Set(&fieldValue, &integer)
		require.NoError(t, err)
		require.Equal(t, integer, *testStructP.I)
	}
}

func testSetUint[T constraints.Unsigned](t *testing.T, integer T) {
	var testStruct struct {
		I T
	}

	v := reflect.Indirect(reflect.ValueOf(&testStruct))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		err := Set(&fieldValue, integer)
		require.NoError(t, err)
		require.Equal(t, integer, testStruct.I)
	}

	var testStructP struct {
		I *T
	}

	v = reflect.Indirect(reflect.ValueOf(&testStructP))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		err := Set(&fieldValue, &integer)
		require.NoError(t, err)
		require.Equal(t, integer, *testStructP.I)
	}
}

func testSetFloat[T constraints.Float](t *testing.T, float T) {
	var testStruct struct {
		F T
	}

	v := reflect.Indirect(reflect.ValueOf(&testStruct))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		err := Set(&fieldValue, float)
		require.NoError(t, err)
		require.Equal(t, float, testStruct.F)
	}

	var testStructP struct {
		F *T
	}

	v = reflect.Indirect(reflect.ValueOf(&testStructP))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		err := Set(&fieldValue, float)
		require.NoError(t, err)
		require.Equal(t, float, *testStructP.F)
	}
}
