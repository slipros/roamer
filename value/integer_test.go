package value

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/constraints"
)

func TestSetInteger(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		t.Parallel()

		testSetIntegerString(t, num)
		testSetIntegerString(t, int8(num))
		testSetIntegerString(t, int16(num))
		testSetIntegerString(t, int32(num))
		testSetIntegerString(t, int64(num))
	})

	t.Run("Boolean", func(t *testing.T) {
		t.Parallel()

		testSetIntegerBoolean(t, 1, true)
		testSetIntegerBoolean(t, 0, false)
		testSetIntegerBoolean(t, int8(1), true)
		testSetIntegerBoolean(t, int8(0), false)
		testSetIntegerBoolean(t, int16(1), true)
		testSetIntegerBoolean(t, int16(0), false)
		testSetIntegerBoolean(t, int32(1), true)
		testSetIntegerBoolean(t, int32(0), false)
		testSetIntegerBoolean(t, int64(1), true)
		testSetIntegerBoolean(t, int64(0), false)
	})

	t.Run("Int", func(t *testing.T) {
		t.Parallel()

		testSetIntegerInt(t, num)
		testSetIntegerInt(t, int8(num))
		testSetIntegerInt(t, int16(num))
		testSetIntegerInt(t, int32(num))
		testSetIntegerInt(t, int64(num))
	})

	t.Run("Uint", func(t *testing.T) {
		t.Parallel()

		testSetIntegerUint(t, num)
		testSetIntegerUint(t, uint8(num))
		testSetIntegerUint(t, uint16(num))
		testSetIntegerUint(t, uint32(num))
		testSetIntegerUint(t, uint64(num))

	})

	t.Run("Float", func(t *testing.T) {
		t.Parallel()

		testSetIntegerFloat(t, num)
		testSetIntegerFloat(t, int8(num))
		testSetIntegerFloat(t, int16(num))
		testSetIntegerFloat(t, int32(num))
		testSetIntegerFloat(t, int64(num))
	})

	t.Run("Unsupported", func(t *testing.T) {
		t.Parallel()

		testSetIntegerUnsupported(t, num)
		testSetIntegerUnsupported(t, int8(num))
		testSetIntegerUnsupported(t, int16(num))
		testSetIntegerUnsupported(t, int32(num))
		testSetIntegerUnsupported(t, int64(num))
	})
}

func testSetIntegerString[T constraints.Integer](t *testing.T, integer T) {
	var testStruct struct {
		S string
	}

	v := reflect.Indirect(reflect.ValueOf(&testStruct))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		err := SetInteger(fieldValue, integer)
		require.NoError(t, err)
		require.Equal(t, strconv.Itoa(int(integer)), testStruct.S)
	}
}

func testSetIntegerBoolean[T constraints.Integer](t *testing.T, integer T, want bool) {
	var testStruct struct {
		B bool
	}

	v := reflect.Indirect(reflect.ValueOf(&testStruct))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		err := SetInteger(fieldValue, integer)
		require.NoError(t, err)

		require.Equal(t, want, testStruct.B)
	}
}

func testSetIntegerInt[T constraints.Integer](t *testing.T, integer T) {
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
		err := SetInteger(fieldValue, integer)
		require.NoError(t, err)
		require.Equal(t, int64(integer), fieldValue.Int())
	}
}

func testSetIntegerUint[T constraints.Integer](t *testing.T, integer T) {
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
		err := SetInteger(fieldValue, integer)
		require.NoError(t, err)
		require.Equal(t, uint64(integer), fieldValue.Uint())
	}
}

func testSetIntegerFloat[T constraints.Integer](t *testing.T, integer T) {
	var testStruct struct {
		FL32 float32
		FL64 float64
	}

	v := reflect.Indirect(reflect.ValueOf(&testStruct))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		err := SetInteger(fieldValue, integer)
		require.NoError(t, err)
		require.Equal(t, float64(num), fieldValue.Float())
	}
}

func testSetIntegerUnsupported[T constraints.Integer](t *testing.T, integer T) {
	var testStruct struct {
		SL []string
	}

	v := reflect.Indirect(reflect.ValueOf(&testStruct))

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		err := SetInteger(fieldValue, integer)
		require.Error(t, err)
	}
}
