package value

import (
	"reflect"
	"strings"
	"testing"

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

	t.Run("String slice", func(t *testing.T) {
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
