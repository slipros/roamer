package decoder

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFormURL(t *testing.T) {
	f := NewFormURL()
	require.NotNil(t, f)
	assert.Equal(t, ContentTypeFormURL, f.ContentType())
	assert.Equal(t, f.splitSymbol, SplitSymbol)

	f = NewFormURL(WithDisabledSplit())
	require.NotNil(t, f)
	assert.Equal(t, ContentTypeFormURL, f.ContentType())
	assert.False(t, f.split)

	f = NewFormURL(WithSplitSymbol("="))
	require.NotNil(t, f)
	assert.Equal(t, ContentTypeFormURL, f.ContentType())
	assert.Equal(t, "=", f.splitSymbol)

	f = NewFormURL(WithContentType[*FormURL]("test"))
	require.NotNil(t, f)
	assert.Equal(t, "test", f.ContentType())

	f = NewFormURL(WithSkipFilled[*FormURL](false))
	require.NotNil(t, f)
	assert.Equal(t, false, f.skipFilled)
}

func TestFormURL_Tag(t *testing.T) {
	f := NewFormURL()
	assert.Equal(t, TagForm, f.Tag())
}

var (
	str           = "string"
	integerString = fmt.Sprintf("%d", 1)
	floatString   = fmt.Sprintf("%f", float32(1))

	integer, _    = strconv.Atoi(integerString)
	integer8      = int8(integer)
	integer16     = int16(integer)
	integer32     = int32(integer)
	integer64     = int64(integer)
	floating32    = float32(integer)
	floating64    = float64(integer)
	complexNum128 = complex(floating64, 0)
	complexNum64  = complex64(complexNum128)
	complexString = fmt.Sprintf("%f", complexNum128)

	form = url.Values{
		"string":          {str},
		"string_ptr":      {str},
		"int":             {integerString},
		"int_ptr":         {integerString},
		"int_8":           {integerString},
		"int_8_ptr":       {integerString},
		"int_16":          {integerString},
		"int_16_ptr":      {integerString},
		"int_32":          {integerString},
		"int_32_ptr":      {integerString},
		"int_64":          {integerString},
		"int_64_ptr":      {integerString},
		"float_32":        {floatString},
		"float_32_ptr":    {floatString},
		"float_64":        {floatString},
		"float_64_ptr":    {floatString},
		"complex_64":      {complexString},
		"complex_64_ptr":  {complexString},
		"complex_128":     {complexString},
		"complex_128_ptr": {complexString},
		"slice_string":    {str, str, str},
	}
)

func TestFormURL_Decode_Successfully(t *testing.T) {
	type fields struct {
		disabledSplit bool
	}
	type args struct {
		req  *http.Request
		ptr  any
		want any
	}
	tests := []struct {
		name         string
		fields       fields
		args         func() args
		wantNotEqual bool
	}{
		{
			name: "Fill struct",
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

				want := Data{
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
					want: &want,
				}
			},
		},
		{
			name: "Fill struct slice field",
			args: func() args {
				type Data struct {
					SliceStr []string `form:"slice_string"`
				}

				form := make(url.Values, 3)
				form.Add("slice_string", str)
				form.Add("slice_string", str)
				form.Add("slice_string", str)

				want := Data{
					SliceStr: []string{str, str, str},
				}

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				return args{
					req:  req,
					ptr:  &Data{},
					want: &want,
				}
			},
		},
		{
			name: "Fill map[string]any ",
			args: func() args {
				want := map[string]any{
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
					"slice_string":    []string{str, str, str},
				}

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				emptyMap := make(map[string]any, len(want))

				return args{
					req:  req,
					ptr:  &emptyMap,
					want: &want,
				}
			},
		},
		{
			name: "Fill map[string]string",
			args: func() args {
				want := map[string]string{
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
					"slice_string":    strings.Join([]string{str, str, str}, SplitSymbol),
				}

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				emptyMap := make(map[string]string, len(want))

				return args{
					req:  req,
					ptr:  &emptyMap,
					want: &want,
				}
			},
		},
		{
			name: "Fill map[string][]string",
			args: func() args {
				form := make(url.Values, 3)
				form.Add("slice_string", str)
				form.Add("slice_string", str)
				form.Add("slice_string", str)

				want := map[string][]string{
					"slice_string": {str, str, str},
				}

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				emptyMap := make(map[string][]string, len(want))

				return args{
					req:  req,
					ptr:  &emptyMap,
					want: &want,
				}
			},
		},
		{
			name: "Fill struct with comma-separated string",
			args: func() args {
				type Data struct {
					Tags []string `form:"tags"`
				}
				form := make(url.Values)
				form.Add("tags", "foo,bar,baz")
				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)
				req.Header.Add("Content-Type", ContentTypeFormURL)
				return args{
					req:  req,
					ptr:  &Data{},
					want: &Data{Tags: []string{"foo", "bar", "baz"}},
				}
			},
		},
		{
			name:   "Fill struct with comma-separated string and split disabled",
			fields: fields{disabledSplit: true},
			args: func() args {
				type Data struct {
					Tags string `form:"tags"`
				}
				form := make(url.Values)
				form.Add("tags", "foo,bar,baz")
				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)
				req.Header.Add("Content-Type", ContentTypeFormURL)
				return args{
					req:  req,
					ptr:  &Data{},
					want: &Data{Tags: "foo,bar,baz"},
				}
			},
		},
		{
			name: "Fill empty struct",
			args: func() args {
				type Data struct{}
				var want Data

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				return args{
					req:  req,
					ptr:  &Data{},
					want: &want,
				}
			},
		},
		{
			name: "Fill struct without form tag",
			args: func() args {
				type Data struct {
					S string `json:"s"`
				}
				var want Data

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				return args{
					req:  req,
					ptr:  &Data{},
					want: &want,
				}
			},
		},
		{
			name: "Fill struct with non exported field",
			args: func() args {
				type Data struct {
					data string
				}
				var want Data

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				return args{
					req:  req,
					ptr:  &Data{},
					want: &want,
				}
			},
		},
		{
			name: "Fill struct without tag",
			args: func() args {
				type Data struct {
					Data string
				}
				var want Data

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				return args{
					req:  req,
					ptr:  &Data{},
					want: &want,
				}
			},
		},
		{
			name: "Fill struct - empty form",
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

				var want Data

				form := make(url.Values)
				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				return args{
					req:  req,
					ptr:  &Data{},
					want: &want,
				}
			},
		},
		{
			name: "Fill map[string]any - empty form",
			args: func() args {
				want := make(map[string]any)

				form := make(url.Values)
				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				emptyMap := make(map[string]any, len(want))

				return args{
					req:  req,
					ptr:  &emptyMap,
					want: &want,
				}
			},
		},
		{
			name:         "Fill struct - wrong form value",
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

				want := Data{
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
					want: &want,
				}
			},
		},
		{
			name:         "Fill map[string]any - wrong form value",
			wantNotEqual: true,
			args: func() args {
				var (
					str           = "string1"
					integerString = fmt.Sprintf("%d", 2)
					floatString   = fmt.Sprintf("%f", float32(2))
					complexString = fmt.Sprintf("%f", complex(2, 0))
				)

				want := map[string]any{
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

				emptyMap := make(map[string]any, len(want))

				return args{
					req: req,
					ptr: &emptyMap,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			args := tt.args()
			var f *FormURL
			if tt.fields.disabledSplit {
				f = NewFormURL(WithDisabledSplit())
			} else {
				f = NewFormURL()
			}

			err := f.Decode(args.req, args.ptr)
			assert.NoError(t, err)

			if tt.wantNotEqual {
				assert.NotEqualValues(t, args.want, args.ptr)
				return
			}

			assert.EqualValues(t, args.want, args.ptr)
		})
	}
}

func TestFormURL_Decode_Failure(t *testing.T) {
	type args struct {
		req *http.Request
		ptr any
	}
	tests := []struct {
		name string
		args func() args
	}{
		{
			name: "Fill map[int]any",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				emptyMap := make(map[int]any)

				return args{
					req: req,
					ptr: &emptyMap,
				}
			},
		},
		{
			name: "Fill map[string]int",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				emptyMap := make(map[string]int)

				return args{
					req: req,
					ptr: &emptyMap,
				}
			},
		},
		{
			name: "Fill struct inside struct",
			args: func() args {
				type Data struct {
					S struct {
						Str string
					} `form:"string"`
				}

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				return args{
					req: req,
					ptr: &Data{},
				}
			},
		},
		{
			name: "Invalid value in request url query",
			args: func() args {
				type Data struct {
					SliceStr []string `form:"slice_string"`
				}

				form := make(url.Values, 3)
				form.Add("slice_string", str)
				form.Add("slice_string", str)
				form.Add("slice_string", str)

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)
				req.URL.RawQuery = "test;test;test"

				req.Header.Add("Content-Type", ContentTypeFormURL)

				return args{
					req: req,
					ptr: &Data{},
				}
			},
		},
		{
			name: "Unsupported ptr",
			args: func() args {

				req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
				require.NoError(t, err)

				req.Header.Add("Content-Type", ContentTypeFormURL)

				var sl []string

				return args{
					req: req,
					ptr: &sl,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			args := tt.args()
			f := NewFormURL()

			err := f.Decode(args.req, args.ptr)
			assert.Error(t, err)
			if err != nil {
				assert.NotEmpty(t, err.Error(), "Error message should not be empty")
			}
		})
	}
}

func BenchmarkFormURL_Decode(b *testing.B) {
	var (
		str           = "string"
		integerString = fmt.Sprintf("%d", 1)
		floatString   = fmt.Sprintf("%f", float32(1))
	)

	integer, err := strconv.Atoi(integerString)
	if err != nil {
		b.Fatal(err)
	}

	floating64 := float64(integer)
	complexNum128 := complex(floating64, 0)
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

	req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
	if err != nil {
		b.Fatal(err)
	}

	req.Header.Add("Content-Type", ContentTypeFormURL)

	f := NewFormURL(WithSkipFilled[*FormURL](false))
	var d Data

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := f.Decode(req, &d); err != nil {
			b.Fatal(err)
		}
	}
}
