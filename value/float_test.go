package value

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/constraints"
)

var num = 1

func TestSetFloat(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		t.Parallel()

		testSetFloatString(t, float32(num))
		testSetFloatString(t, float64(num))
	})

	t.Run("Boolean", func(t *testing.T) {
		t.Parallel()

		testSetFloatBoolean(t, float32(1), true)
		testSetFloatBoolean(t, float32(0), false)
		testSetFloatBoolean(t, float64(1), true)
		testSetFloatBoolean(t, float64(0), false)
	})

	t.Run("Int", func(t *testing.T) {
		t.Parallel()

		testSetFloatInt(t, float32(num))
		testSetFloatInt(t, float64(num))
	})

	t.Run("Uint", func(t *testing.T) {
		t.Parallel()

		testSetFloatUint(t, float32(num))
		testSetFloatUint(t, float64(num))
	})

	t.Run("Float", func(t *testing.T) {
		t.Parallel()

		testSetFloatFloat(t, float32(num))
		testSetFloatFloat(t, float64(num))
	})

	t.Run("Unsupported", func(t *testing.T) {
		t.Parallel()

		testSetFloatUnsupported(t, float32(num))
		testSetFloatUnsupported(t, float64(num))
	})
}

func testSetFloatString[T constraints.Float](t *testing.T, float T) {
	var testStruct struct {
		S string
	}

	v := reflect.Indirect(reflect.ValueOf(&testStruct))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		err := SetFloat(&fieldValue, float)
		require.NoError(t, err)
		require.Equal(t, strconv.FormatFloat(float64(num), 'E', -1, 64), testStruct.S)
	}
}

func testSetFloatBoolean[T constraints.Float](t *testing.T, float T, want bool) {
	var testStruct struct {
		B bool
	}

	v := reflect.Indirect(reflect.ValueOf(&testStruct))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		err := SetFloat(&fieldValue, float)
		require.NoError(t, err)

		require.Equal(t, want, testStruct.B)
	}
}

func testSetFloatInt[T constraints.Float](t *testing.T, float T) {
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
		err := SetFloat(&fieldValue, float)
		require.NoError(t, err)
		require.Equal(t, int64(float), fieldValue.Int())
	}
}

func testSetFloatUint[T constraints.Float](t *testing.T, float T) {
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
		err := SetFloat(&fieldValue, float)
		require.NoError(t, err)
		require.Equal(t, uint64(float), fieldValue.Uint())
	}
}

func testSetFloatFloat[T constraints.Float](t *testing.T, float T) {
	var testStruct struct {
		FL32 float32
		FL64 float64
	}

	v := reflect.Indirect(reflect.ValueOf(&testStruct))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		err := SetFloat(&fieldValue, float)
		require.NoError(t, err)
		require.Equal(t, float64(num), fieldValue.Float())
	}
}

func testSetFloatUnsupported[T constraints.Float](t *testing.T, float T) {
	var testStruct struct {
		SL []string
	}

	v := reflect.Indirect(reflect.ValueOf(&testStruct))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		err := SetFloat(&fieldValue, float)
		require.Error(t, err)
	}
}
