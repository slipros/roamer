package value

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/constraints"
)

var num = 1
var str = "test_string"

type implementsStringer struct {
}

func (i *implementsStringer) String() string {
	return "hello"
}

func TestSet(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		var testStruct struct {
			S string
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(fieldValue, str)
			require.NoError(t, err)
			require.Equal(t, str, testStruct.S)
		}

		var testStructP struct {
			S *string
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(fieldValue, str)
			require.NoError(t, err)
			require.Equal(t, str, *testStructP.S)
		}
	})

	t.Run("String ptr", func(t *testing.T) {
		var testStruct struct {
			S *string
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(fieldValue, str)
			require.NoError(t, err)
			require.Equal(t, str, *testStruct.S)
		}

		var testStructP struct {
			S *string
		}

		v = reflect.Indirect(reflect.ValueOf(&testStructP))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(fieldValue, str)
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

	t.Run("Int ptr", func(t *testing.T) {
		t.Parallel()

		testSetIntPointer(t, num)
		testSetIntPointer(t, int8(num))
		testSetIntPointer(t, int16(num))
		testSetIntPointer(t, int32(num))
		testSetIntPointer(t, int64(num))
	})

	t.Run("Uint", func(t *testing.T) {
		t.Parallel()

		testSetUint(t, uint(num))
		testSetUint(t, uint8(num))
		testSetUint(t, uint16(num))
		testSetUint(t, uint32(num))
		testSetUint(t, uint64(num))
	})

	t.Run("Uint ptr", func(t *testing.T) {
		t.Parallel()

		testSetUintPointer(t, uint(num))
		testSetUintPointer(t, uint8(num))
		testSetUintPointer(t, uint16(num))
		testSetUintPointer(t, uint32(num))
		testSetUintPointer(t, uint64(num))
	})

	t.Run("Float", func(t *testing.T) {
		t.Parallel()

		testSetFloat(t, float32(num))
		testSetFloat(t, float64(num))
	})

	t.Run("Float ptr", func(t *testing.T) {
		t.Parallel()

		testSetFloatPointer(t, float32(num))
		testSetFloatPointer(t, float64(num))
	})

	t.Run("Slice string", func(t *testing.T) {
		var testStruct struct {
			SL []string
		}

		sl := []string{str, str}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(fieldValue, sl)
			require.NoError(t, err)
			require.Equal(t, sl, testStruct.SL)
		}
	})

	t.Run("Same type", func(t *testing.T) {
		var testStruct struct {
			M map[string]string
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		m := map[string]string{
			str: str,
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(fieldValue, m)
			require.NoError(t, err)
		}
	})

	t.Run("Stringer", func(t *testing.T) {
		var testStruct struct {
			M string
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		m := &implementsStringer{}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(fieldValue, m)
			require.NoError(t, err)
		}
	})

	t.Run("Unsupported", func(t *testing.T) {
		var testStruct struct {
			M struct {
			}
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		m := map[string]string{
			str: str,
		}

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := Set(fieldValue, m)
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
		err := Set(fieldValue, integer)
		require.NoError(t, err)
		require.Equal(t, integer, testStruct.I)
	}

	var testStructP struct {
		I *T
	}

	v = reflect.Indirect(reflect.ValueOf(&testStructP))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		err := Set(fieldValue, &integer)
		require.NoError(t, err)
		require.Equal(t, integer, *testStructP.I)
	}
}

func testSetIntPointer[T constraints.Integer](t *testing.T, integer T) {
	var testStruct struct {
		I *T
	}

	v := reflect.Indirect(reflect.ValueOf(&testStruct))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		err := Set(fieldValue, integer)
		require.NoError(t, err)
		require.Equal(t, integer, *testStruct.I)
	}

	var testStructP struct {
		I *T
	}

	v = reflect.Indirect(reflect.ValueOf(&testStructP))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		err := Set(fieldValue, &integer)
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
		err := Set(fieldValue, integer)
		require.NoError(t, err)
		require.Equal(t, integer, testStruct.I)
	}

	var testStructP struct {
		I *T
	}

	v = reflect.Indirect(reflect.ValueOf(&testStructP))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		err := Set(fieldValue, &integer)
		require.NoError(t, err)
		require.Equal(t, integer, *testStructP.I)
	}
}

func testSetUintPointer[T constraints.Unsigned](t *testing.T, integer T) {
	var testStruct struct {
		I *T
	}

	v := reflect.Indirect(reflect.ValueOf(&testStruct))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		err := Set(fieldValue, integer)
		require.NoError(t, err)
		require.Equal(t, integer, *testStruct.I)
	}

	var testStructP struct {
		I *T
	}

	v = reflect.Indirect(reflect.ValueOf(&testStructP))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		err := Set(fieldValue, &integer)
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
		err := Set(fieldValue, float)
		require.NoError(t, err)
		require.Equal(t, float, testStruct.F)
	}

	var testStructP struct {
		F *T
	}

	v = reflect.Indirect(reflect.ValueOf(&testStructP))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		err := Set(fieldValue, float)
		require.NoError(t, err)
		require.Equal(t, float, *testStructP.F)
	}
}

func testSetFloatPointer[T constraints.Float](t *testing.T, float T) {
	var testStruct struct {
		F *T
	}

	v := reflect.Indirect(reflect.ValueOf(&testStruct))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		err := Set(fieldValue, float)
		require.NoError(t, err)
		require.Equal(t, float, *testStruct.F)
	}

	var testStructP struct {
		F *T
	}

	v = reflect.Indirect(reflect.ValueOf(&testStructP))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		err := Set(fieldValue, float)
		require.NoError(t, err)
		require.Equal(t, float, *testStructP.F)
	}
}
