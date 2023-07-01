package value

import (
	"reflect"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestSetSliceString(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		sl := []string{str, str}

		var testStruct struct {
			S string
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetSliceString(&fieldValue, sl)
			require.NoError(t, err)
			require.Equal(t, strings.Join(sl, ","), testStruct.S)
		}
	})

	t.Run("[]string", func(t *testing.T) {
		sl := []string{str, str}

		var testStruct struct {
			SL []string
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetSliceString(&fieldValue, sl)
			require.NoError(t, err)
			require.Equal(t, sl, testStruct.SL)
		}
	})

	t.Run("[]any", func(t *testing.T) {
		sl := []string{str, str}

		var testStruct struct {
			SL []any
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetSliceString(&fieldValue, sl)
			require.NoError(t, err)
			require.Equal(t, []any{str, str}, testStruct.SL)
		}
	})

	/*
		t.Run("[]string in any", func(t *testing.T) {
			sl := []string{str, str}

			var testStruct struct {
				Str any
			}

			v := reflect.Indirect(reflect.ValueOf(&testStruct))

			for i := 0; i < v.NumField(); i++ {
				fieldValue := v.Field(i)
				err := SetSliceString(&fieldValue, sl)
				require.NoError(t, err)
				require.Equal(t, sl, testStruct.Str)
			}
		})
	*/

	t.Run("[]error", func(t *testing.T) {
		sl := []string{str, str}

		var testStruct struct {
			SL []error
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetSliceString(&fieldValue, sl)
			require.Error(t, err)
		}
	})

	t.Run("not assignable interface", func(t *testing.T) {
		sl := []string{str, str}

		var testStruct struct {
			Err error
		}

		testStruct.Err = errors.New("")

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetSliceString(&fieldValue, sl)
			require.Error(t, err)
		}
	})

	t.Run("Unsupported", func(t *testing.T) {
		sl := []string{str, str}

		var testStruct struct {
			M map[string]string
		}

		v := reflect.Indirect(reflect.ValueOf(&testStruct))

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			err := SetSliceString(&fieldValue, sl)
			require.Error(t, err)
		}
	})
}
