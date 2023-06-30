package decoder

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFormURL(t *testing.T) {
	formURLDecoder := NewFormURL()
	require.NotNil(t, formURLDecoder)
	require.Equal(t, ContentTypeFormURL, formURLDecoder.ContentType())
	require.Equal(t, formURLDecoder.splitSymbol, SplitSymbol)

	formURLDecoderWithDisabledSplit := NewFormURL(WithDisabledSplit())
	require.NotNil(t, formURLDecoder)
	require.Equal(t, ContentTypeFormURL, formURLDecoder.ContentType())
	require.False(t, formURLDecoderWithDisabledSplit.split)

	formURLDecoderWithSplitSymbol := NewFormURL(WithSplitSymbol("="))
	require.NotNil(t, formURLDecoder)
	require.Equal(t, ContentTypeFormURL, formURLDecoder.ContentType())
	require.Equal(t, formURLDecoderWithSplitSymbol.splitSymbol, "=")
}

func TestFormURL_Decode(t *testing.T) {
	var (
		str           = "string"
		integerString = fmt.Sprintf("%d", 1)
		floatString   = fmt.Sprintf("%f", float32(1))
	)

	integer, err := strconv.Atoi(integerString)
	require.NoError(t, err)

	integer8 := int8(integer)
	integer16 := int16(integer)
	integer32 := int32(integer)
	integer64 := int64(integer)
	floating32 := float32(integer)
	floating64 := float64(integer)
	complexNum128 := complex(floating64, 0)
	complexNum64 := complex64(complexNum128)
	complexString := fmt.Sprintf("%f", complexNum128)

	form := make(url.Values, 20)
	form.Add("string", str)
	form.Add("string_ptr", str)
	form.Add("int", integerString)
	form.Add("int_ptr", integerString)
	form.Add("int_8", integerString)
	form.Add("int_8_ptr", integerString)
	form.Add("int_16", integerString)
	form.Add("int_16_ptr", integerString)
	form.Add("int_32", integerString)
	form.Add("int_32_ptr", integerString)
	form.Add("int_64", integerString)
	form.Add("int_64_ptr", integerString)
	form.Add("float_32", floatString)
	form.Add("float_32_ptr", floatString)
	form.Add("float_64", floatString)
	form.Add("float_64_ptr", floatString)
	form.Add("complex_64", complexString)
	form.Add("complex_64_ptr", complexString)
	form.Add("complex_128", complexString)
	form.Add("complex_128_ptr", complexString)

	type args struct {
		req  *http.Request
		ptr  any
		want any
	}
	tests := []struct {
		name         string
		args         func() args
		wantNotEqual bool
		wantErr      bool
	}{
		{
			name: "Fill struct fields",
			args: func() args {
				type Data struct {
					String        string      `form:"string"`
					StringPtr     *string     `form:"string_ptr"`
					Integer       int         `form:"int"`
					IntegerPtr    *int        `form:"int_ptr"`
					Integer8      int8        `form:"int_8"`
					Integer8Ptr   *int8       `form:"int_8_ptr"`
					Integer16     int16       `form:"int_16"`
					Integer16Ptr  *int16      `form:"int_16_ptr"`
					Integer32     int32       `form:"int_32"`
					Integer32Ptr  *int32      `form:"int_32_ptr"`
					Integer64     int64       `form:"int_64"`
					Integer64Ptr  *int64      `form:"int_64_ptr"`
					Float32       float32     `form:"float_32"`
					Float32Ptr    *float32    `form:"float_32_ptr"`
					Float64       float64     `form:"float_64"`
					Float64Ptr    *float64    `form:"float_64_ptr"`
					Complex64     complex64   `form:"complex_64"`
					Complex64Ptr  *complex64  `form:"complex_64_ptr"`
					Complex128    complex128  `form:"complex_128"`
					Complex128Ptr *complex128 `form:"complex_128_ptr"`
				}

				data := Data{
					String:        str,
					StringPtr:     &str,
					Integer:       integer,
					IntegerPtr:    &integer,
					Integer8:      integer8,
					Integer8Ptr:   &integer8,
					Integer16:     integer16,
					Integer16Ptr:  &integer16,
					Integer32:     integer32,
					Integer32Ptr:  &integer32,
					Integer64:     integer64,
					Integer64Ptr:  &integer64,
					Float32:       floating32,
					Float32Ptr:    &floating32,
					Float64:       floating64,
					Float64Ptr:    &floating64,
					Complex64:     complexNum64,
					Complex64Ptr:  &complexNum64,
					Complex128:    complexNum128,
					Complex128Ptr: &complexNum128,
				}

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				return args{
					req:  req,
					ptr:  &Data{},
					want: &data,
				}
			},
		},
		{
			name: "Fill map any fields",
			args: func() args {
				data := map[string]any{
					"string":          str,
					"string_ptr":      str,
					"int":             integerString,
					"int_ptr":         integerString,
					"int_8":           integerString,
					"int_8_ptr":       integerString,
					"int_16":          integerString,
					"int_16_ptr":      integerString,
					"int_32":          integerString,
					"int_32_ptr":      integerString,
					"int_64":          integerString,
					"int_64_ptr":      integerString,
					"float_32":        floatString,
					"float_32_ptr":    floatString,
					"float_64":        floatString,
					"float_64_ptr":    floatString,
					"complex_64":      complexString,
					"complex_64_ptr":  complexString,
					"complex_128":     complexString,
					"complex_128_ptr": complexString,
				}

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				emptyMap := make(map[string]any, len(data))

				return args{
					req:  req,
					ptr:  &emptyMap,
					want: &data,
				}
			},
		},
		{
			name: "Fill string map fields",
			args: func() args {
				data := map[string]string{
					"string":          str,
					"string_ptr":      str,
					"int":             integerString,
					"int_ptr":         integerString,
					"int_8":           integerString,
					"int_8_ptr":       integerString,
					"int_16":          integerString,
					"int_16_ptr":      integerString,
					"int_32":          integerString,
					"int_32_ptr":      integerString,
					"int_64":          integerString,
					"int_64_ptr":      integerString,
					"float_32":        floatString,
					"float_32_ptr":    floatString,
					"float_64":        floatString,
					"float_64_ptr":    floatString,
					"complex_64":      complexString,
					"complex_64_ptr":  complexString,
					"complex_128":     complexString,
					"complex_128_ptr": complexString,
				}

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				emptyMap := make(map[string]string, len(data))

				return args{
					req:  req,
					ptr:  &emptyMap,
					want: &data,
				}
			},
		},
		{
			name: "Fill empty struct",
			args: func() args {
				type Data struct{}
				data := Data{}

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				return args{
					req:  req,
					ptr:  &Data{},
					want: &data,
				}
			},
		},
		{
			name: "Fill struct fields - empty form",
			args: func() args {
				type Data struct {
					String        string      `form:"string"`
					StringPtr     *string     `form:"string_ptr"`
					Integer       int         `form:"int"`
					IntegerPtr    *int        `form:"int_ptr"`
					Integer8      int8        `form:"int_8"`
					Integer8Ptr   *int8       `form:"int_8_ptr"`
					Integer16     int16       `form:"int_16"`
					Integer16Ptr  *int16      `form:"int_16_ptr"`
					Integer32     int32       `form:"int_32"`
					Integer32Ptr  *int32      `form:"int_32_ptr"`
					Integer64     int64       `form:"int_64"`
					Integer64Ptr  *int64      `form:"int_64_ptr"`
					Float32       float32     `form:"float_32"`
					Float32Ptr    *float32    `form:"float_32_ptr"`
					Float64       float64     `form:"float_64"`
					Float64Ptr    *float64    `form:"float_64_ptr"`
					Complex64     complex64   `form:"complex_64"`
					Complex64Ptr  *complex64  `form:"complex_64_ptr"`
					Complex128    complex128  `form:"complex_128"`
					Complex128Ptr *complex128 `form:"complex_128_ptr"`
				}

				data := Data{}

				form := make(url.Values)

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				return args{
					req:  req,
					ptr:  &Data{},
					want: &data,
				}
			},
		},
		{
			name: "Fill map fields - empty form",
			args: func() args {
				data := map[string]any{}

				form := make(url.Values)

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				emptyMap := make(map[string]any, len(data))

				return args{
					req:  req,
					ptr:  &emptyMap,
					want: &data,
				}
			},
		},
		{
			name:         "Fill struct fields - wrong form value",
			wantNotEqual: true,
			args: func() args {
				type Data struct {
					String        string      `form:"string"`
					StringPtr     *string     `form:"string_ptr"`
					Integer       int         `form:"int"`
					IntegerPtr    *int        `form:"int_ptr"`
					Integer8      int8        `form:"int_8"`
					Integer8Ptr   *int8       `form:"int_8_ptr"`
					Integer16     int16       `form:"int_16"`
					Integer16Ptr  *int16      `form:"int_16_ptr"`
					Integer32     int32       `form:"int_32"`
					Integer32Ptr  *int32      `form:"int_32_ptr"`
					Integer64     int64       `form:"int_64"`
					Integer64Ptr  *int64      `form:"int_64_ptr"`
					Float32       float32     `form:"float_32"`
					Float32Ptr    *float32    `form:"float_32_ptr"`
					Float64       float64     `form:"float_64"`
					Float64Ptr    *float64    `form:"float_64_ptr"`
					Complex64     complex64   `form:"complex_64"`
					Complex64Ptr  *complex64  `form:"complex_64_ptr"`
					Complex128    complex128  `form:"complex_128"`
					Complex128Ptr *complex128 `form:"complex_128_ptr"`
				}

				data := Data{
					String:        str,
					StringPtr:     &str,
					Integer:       integer,
					IntegerPtr:    &integer,
					Integer8:      integer8,
					Integer8Ptr:   &integer8,
					Integer16:     integer16,
					Integer16Ptr:  &integer16,
					Integer32:     integer32,
					Integer32Ptr:  &integer32,
					Integer64:     integer64,
					Integer64Ptr:  &integer64,
					Float32:       floating32,
					Float32Ptr:    &floating32,
					Float64:       floating64,
					Float64Ptr:    &floating64,
					Complex64:     complexNum64,
					Complex64Ptr:  &complexNum64,
					Complex128:    complexNum128,
					Complex128Ptr: &complexNum128,
				}

				var (
					str           = "string1"
					integerString = fmt.Sprintf("%d", 2)
					floatString   = fmt.Sprintf("%f", float32(2))
					complexString = fmt.Sprintf("%f", complex(2, 0))
				)

				form := make(url.Values, 20)
				form.Add("string", str)
				form.Add("string_ptr", str)
				form.Add("int", integerString)
				form.Add("int_ptr", integerString)
				form.Add("int_8", integerString)
				form.Add("int_8_ptr", integerString)
				form.Add("int_16", integerString)
				form.Add("int_16_ptr", integerString)
				form.Add("int_32", integerString)
				form.Add("int_32_ptr", integerString)
				form.Add("int_64", integerString)
				form.Add("int_64_ptr", integerString)
				form.Add("float_32", floatString)
				form.Add("float_32_ptr", floatString)
				form.Add("float_64", floatString)
				form.Add("float_64_ptr", floatString)
				form.Add("complex_64", complexString)
				form.Add("complex_64_ptr", complexString)
				form.Add("complex_128", complexString)
				form.Add("complex_128_ptr", complexString)

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				return args{
					req:  req,
					ptr:  &Data{},
					want: &data,
				}
			},
		},
		{
			name:         "Fill map fields - wrong form value",
			wantNotEqual: true,
			args: func() args {
				var (
					str           = "string1"
					integerString = fmt.Sprintf("%d", 2)
					floatString   = fmt.Sprintf("%f", float32(2))
					complexString = fmt.Sprintf("%f", complex(2, 0))
				)

				data := map[string]any{
					"string":          str,
					"string_ptr":      str,
					"int":             integerString,
					"int_ptr":         integerString,
					"int_8":           integerString,
					"int_8_ptr":       integerString,
					"int_16":          integerString,
					"int_16_ptr":      integerString,
					"int_32":          integerString,
					"int_32_ptr":      integerString,
					"int_64":          integerString,
					"int_64_ptr":      integerString,
					"float_32":        floatString,
					"float_32_ptr":    floatString,
					"float_64":        floatString,
					"float_64_ptr":    floatString,
					"complex_64":      complexString,
					"complex_64_ptr":  complexString,
					"complex_128":     complexString,
					"complex_128_ptr": complexString,
				}

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				emptyMap := make(map[string]any, len(data))

				return args{
					req:  req,
					ptr:  &emptyMap,
					want: &data,
				}
			},
		},
		{
			name:    "Unsupported ptr",
			wantErr: true,
			args: func() args {
				var data []string

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				return args{
					req:  req,
					ptr:  data,
					want: &data,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args()
			f := NewFormURL()

			err := f.Decode(args.req, args.ptr)
			if !tt.wantErr && err != nil {
				t.Errorf("Decode() error = %v", err)
			}

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			if tt.wantNotEqual {
				require.NotEqualValues(t, args.want, args.ptr)
				return
			}

			require.EqualValues(t, args.want, args.ptr)
		})
	}
}
