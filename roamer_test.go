package roamer

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/formatter"
	"github.com/slipros/roamer/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errBigBad = errors.New("big bad error")

type testAfterParser struct {
	Value string
}

// Implement AfterParser for testAfterParser
func (p *testAfterParser) AfterParse(r *http.Request) error {
	p.Value = "processed"
	return nil
}

// Implement AfterParser with error
type errorAfterParser struct{}

func (p *errorAfterParser) AfterParse(r *http.Request) error {
	return errBigBad
}

// Common test types and utilities
type fields struct {
	opts []OptionsFunc
}

type args struct {
	req *http.Request
	ptr any
}

type testStruct struct {
	String string `json:"string" header:"X-String" query:"string"`
	Int    int    `json:"int" header:"X-Int" query:"int"`
}

// Multi-source struct with dedicated fields for each source
type multiSourceStruct struct {
	// JSON-specific fields
	JSONString string `json:"string"`
	JSONInt    int    `json:"int"`

	// Header-specific fields
	HeaderString string `header:"X-String"`
	HeaderInt    int    `header:"X-Int"`

	// Query-specific fields
	QueryString string `query:"string"`
	QueryInt    int    `query:"int"`
}

// TestRoamer_Parse_Successfully tests successful parsing scenarios
func TestRoamer_Parse_Successfully(t *testing.T) {
	// Create test JSON
	testJSON, _ := json.Marshal(testStruct{
		String: "test",
		Int:    123,
	})

	tests := []struct {
		name   string
		setup  func() (fields, args)
		verify func(*testing.T, any)
	}{
		{
			name: "parse struct success with JSON decoder",
			setup: func() (fields, args) {
				// Create request with JSON body
				req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(testJSON))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Content-Length", strconv.Itoa(len(testJSON)))

				// Create struct to fill
				target := &testStruct{}

				// Configure roamer with JSON decoder
				return fields{
						opts: []OptionsFunc{
							WithDecoders(decoder.NewJSON()),
						},
					}, args{
						req: req,
						ptr: target,
					}
			},
			verify: func(t *testing.T, result any) {
				target, ok := result.(*testStruct)
				require.True(t, ok)
				assert.Equal(t, "test", target.String)
				assert.Equal(t, 123, target.Int)
			},
		},
		{
			name: "parse struct with query parameters",
			setup: func() (fields, args) {
				// Create request with query parameters
				req, _ := http.NewRequest(http.MethodGet, "http://example.com?string=queryValue&int=456", nil)

				// Create struct to fill
				target := &testStruct{}

				// Configure roamer with query parser
				return fields{
						opts: []OptionsFunc{
							WithParsers(parser.NewQuery()),
						},
					}, args{
						req: req,
						ptr: target,
					}
			},
			verify: func(t *testing.T, result any) {
				target, ok := result.(*testStruct)
				require.True(t, ok)
				assert.Equal(t, "queryValue", target.String)
				assert.Equal(t, 456, target.Int)
			},
		},
		{
			name: "parse struct with headers",
			setup: func() (fields, args) {
				// Create request with headers
				req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
				req.Header.Set("X-String", "headerValue")
				req.Header.Set("X-Int", "789")

				// Create struct to fill
				target := &testStruct{}

				// Configure roamer with header parser
				return fields{
						opts: []OptionsFunc{
							WithParsers(parser.NewHeader()),
						},
					}, args{
						req: req,
						ptr: target,
					}
			},
			verify: func(t *testing.T, result any) {
				target, ok := result.(*testStruct)
				require.True(t, ok)
				assert.Equal(t, "headerValue", target.String)
				assert.Equal(t, 789, target.Int)
			},
		},
		{
			name: "parse struct with multiple sources",
			setup: func() (fields, args) {
				// Create test JSON for this specific test
				jsonData, _ := json.Marshal(struct {
					String string `json:"string"`
					Int    int    `json:"int"`
				}{
					String: "jsonValue",
					Int:    123,
				})

				// Create request with all data sources
				req, _ := http.NewRequest(http.MethodPost, "http://example.com?string=queryValue&int=456", bytes.NewReader(jsonData))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Content-Length", strconv.Itoa(len(jsonData)))
				req.Header.Set("X-String", "headerValue")
				req.Header.Set("X-Int", "789")

				// Create multi-source struct to fill
				target := &multiSourceStruct{}

				// Configure roamer with parsers and decoders
				return fields{
						opts: []OptionsFunc{
							WithDecoders(decoder.NewJSON()),
							WithParsers(parser.NewHeader(), parser.NewQuery()),
						},
					}, args{
						req: req,
						ptr: target,
					}
			},
			verify: func(t *testing.T, result any) {
				target, ok := result.(*multiSourceStruct)
				require.True(t, ok)

				// Each field should be filled from its specific source
				assert.Equal(t, "jsonValue", target.JSONString, "JSON field should be filled from JSON body")
				assert.Equal(t, 123, target.JSONInt, "JSON field should be filled from JSON body")

				assert.Equal(t, "headerValue", target.HeaderString, "Header field should be filled from HTTP header")
				assert.Equal(t, 789, target.HeaderInt, "Header field should be filled from HTTP header")

				assert.Equal(t, "queryValue", target.QueryString, "Query field should be filled from URL query")
				assert.Equal(t, 456, target.QueryInt, "Query field should be filled from URL query")
			},
		},
		{
			name: "parse map from JSON",
			setup: func() (fields, args) {
				// Create request with JSON body
				req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(testJSON))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Content-Length", strconv.Itoa(len(testJSON)))

				// Create map to fill
				target := make(map[string]interface{})

				// Configure roamer with JSON decoder
				return fields{
						opts: []OptionsFunc{
							WithDecoders(decoder.NewJSON()),
						},
					}, args{
						req: req,
						ptr: &target,
					}
			},
			verify: func(t *testing.T, result any) {
				target, ok := result.(*map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "test", (*target)["string"])
				assert.Equal(t, float64(123), (*target)["int"]) // JSON numbers are decoded to float64
			},
		},
		{
			name: "parse slice from JSON",
			setup: func() (fields, args) {
				// Create test slice and JSON
				testSlice := []string{"item1", "item2", "item3"}
				testSliceJSON, _ := json.Marshal(testSlice)

				// Create request with JSON body
				req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(testSliceJSON))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Content-Length", strconv.Itoa(len(testSliceJSON)))

				// Create slice to fill
				target := []string{}

				// Configure roamer with JSON decoder
				return fields{
						opts: []OptionsFunc{
							WithDecoders(decoder.NewJSON()),
						},
					}, args{
						req: req,
						ptr: &target,
					}
			},
			verify: func(t *testing.T, result any) {
				target, ok := result.(*[]string)
				require.True(t, ok)
				assert.Equal(t, []string{"item1", "item2", "item3"}, *target)
			},
		},
		{
			name: "parse with string formatter",
			setup: func() (fields, args) {
				// Create request with headers with spaces
				req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
				req.Header.Set("X-String", "  headerValue  ")

				// Create struct to fill
				target := &struct {
					String string `header:"X-String" string:"trim_space"`
				}{}

				// Configure roamer with header parser and string formatter
				return fields{
						opts: []OptionsFunc{
							WithParsers(parser.NewHeader()),
							WithFormatters(formatter.NewString()),
						},
					}, args{
						req: req,
						ptr: target,
					}
			},
			verify: func(t *testing.T, result any) {
				target, ok := result.(*struct {
					String string `header:"X-String" string:"trim_space"`
				})
				require.True(t, ok)
				assert.Equal(t, "headerValue", target.String) // spaces should be removed
			},
		},
		{
			name: "parse with AfterParser",
			setup: func() (fields, args) {
				// Create request
				req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

				// Create struct with AfterParser
				target := &testAfterParser{}

				return fields{}, args{
					req: req,
					ptr: target,
				}
			},
			verify: func(t *testing.T, result any) {
				target, ok := result.(*testAfterParser)
				require.True(t, ok)
				assert.Equal(t, "processed", target.Value)
			},
		},
		{
			name: "skip filled fields",
			setup: func() (fields, args) {
				// Create request with query parameters
				req, _ := http.NewRequest(http.MethodGet, "http://example.com?string=queryValue&int=456", nil)

				// Create struct to fill with prefilled field
				target := &testStruct{
					String: "prefilled", // This field should not be overwritten
				}

				// Configure roamer with query parser and skipFilled=true
				return fields{
						opts: []OptionsFunc{
							WithParsers(parser.NewQuery()),
							WithSkipFilled(true),
						},
					}, args{
						req: req,
						ptr: target,
					}
			},
			verify: func(t *testing.T, result any) {
				target, ok := result.(*testStruct)
				require.True(t, ok)
				assert.Equal(t, "prefilled", target.String) // Value should not change
				assert.Equal(t, 456, target.Int)
			},
		},
		{
			name: "don't skip filled fields",
			setup: func() (fields, args) {
				// Create request with query parameters
				req, _ := http.NewRequest(http.MethodGet, "http://example.com?string=queryValue&int=456", nil)

				// Create struct to fill with prefilled field
				target := &testStruct{
					String: "prefilled", // This field should be overwritten
				}

				// Configure roamer with query parser and skipFilled=false
				return fields{
						opts: []OptionsFunc{
							WithParsers(parser.NewQuery()),
							WithSkipFilled(false),
						},
					}, args{
						req: req,
						ptr: target,
					}
			},
			verify: func(t *testing.T, result any) {
				target, ok := result.(*testStruct)
				require.True(t, ok)
				assert.Equal(t, "queryValue", target.String) // Value should be overwritten
				assert.Equal(t, 456, target.Int)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test scenario
			f, a := tt.setup()

			r := NewRoamer(f.opts...)
			err := r.Parse(a.req, a.ptr)
			// In success tests we expect no errors
			require.NoError(t, err, "Parse() should not return an error")

			if tt.verify != nil {
				tt.verify(t, a.ptr)
			}
		})
	}
}

// TestRoamer_Parse_Failure tests parsing scenarios that result in errors
func TestRoamer_Parse_Failure(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() (fields, args)
		expectedError error // Optional: to check specific error types
	}{
		{
			name: "nil pointer",
			setup: func() (fields, args) {
				return fields{}, args{
					req: &http.Request{},
					ptr: nil,
				}
			},
		},
		{
			name: "not a pointer",
			setup: func() (fields, args) {
				return fields{}, args{
					req: &http.Request{},
					ptr: testStruct{},
				}
			},
		},
		{
			name: "unsupported target type",
			setup: func() (fields, args) {
				var unsupportedType int
				return fields{}, args{
					req: &http.Request{},
					ptr: &unsupportedType,
				}
			},
		},
		{
			name: "AfterParser returns error",
			setup: func() (fields, args) {
				// Create request
				req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)

				// Create struct with AfterParser that returns error
				target := &errorAfterParser{}

				return fields{}, args{
					req: req,
					ptr: target,
				}
			},
			expectedError: errBigBad,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test scenario
			f, a := tt.setup()

			r := NewRoamer(f.opts...)
			err := r.Parse(a.req, a.ptr)
			// In failure tests we expect errors
			require.Error(t, err, "Parse() should return an error")

			// If a specific error is expected, check it
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError, "Wrong error type returned")
			}
		})
	}
}

// Benchmark helper types and functions
// RequestPayload represents a standard payload for benchmark requests
type RequestPayload struct {
	Strings []string               `json:"strings"`
	Numbers []int                  `json:"numbers"`
	Map     map[string]interface{} `json:"map"`
}

// SmallStruct for small payload benchmarks
type SmallStruct struct {
	String string `json:"string" header:"X-String" query:"string"`
	Int    int    `json:"int" header:"X-Int" query:"int"`
}

// MediumStruct for medium payload benchmarks
type MediumStruct struct {
	Int       int       `query:"int"`
	Int8      int8      `query:"int8"`
	Int32     int32     `query:"int32"`
	Int64     int64     `query:"int64"`
	Time      time.Time `query:"time"`
	Url       url.URL   `query:"url"`
	UserAgent string    `header:"User-Agent"`
	Strings   []string  `json:"strings"`
}

// LargeStruct for large payload benchmarks
type LargeStruct struct {
	String        string                 `json:"string" header:"X-String" query:"string" string:"trim_space"`
	Int           int                    `json:"int" header:"X-Int" query:"int"`
	Int8          int8                   `query:"int8"`
	Int16         int16                  `query:"int16"`
	Int32         int32                  `query:"int32"`
	Int64         int64                  `query:"int64"`
	Uint          uint                   `query:"uint"`
	Uint8         uint8                  `query:"uint8"`
	Uint16        uint16                 `query:"uint16"`
	Uint32        uint32                 `query:"uint32"`
	Uint64        uint64                 `query:"uint64"`
	Float32       float32                `query:"float32"`
	Float64       float64                `query:"float64"`
	Bool          bool                   `query:"bool"`
	Time          time.Time              `query:"time"`
	Url           url.URL                `query:"url"`
	UserAgent     string                 `header:"User-Agent"`
	Accept        string                 `header:"Accept"`
	RefererHeader string                 `header:"Referer"`
	CustomHeader  string                 `header:"X-Custom-Header"`
	StringsArray  []string               `json:"strings"`
	NumbersArray  []int                  `json:"numbers"`
	MapData       map[string]interface{} `json:"map"`
}

// Helper function to generate JSON body
func toJSON(v any) (int, io.Reader, error) {
	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(&v); err != nil {
		return 0, nil, err
	}

	return buffer.Len(), &buffer, nil
}

// Helper function to prepare a standard HTTP request for benchmarks
func prepareTestHTTPRequest(b *testing.B, method string, withJSON, withHeaders, withQuery bool) (*http.Request, int64) {
	var bodyLen int64
	var body io.Reader

	// Prepare query parameters
	query := make(url.Values)
	if withQuery {
		query.Add("string", "valueFromQuery")
		query.Add("int", "9223372036854775807")
		query.Add("int8", "127")
		query.Add("int16", "32767")
		query.Add("int32", "2147483647")
		query.Add("int64", "9223372036854775807")
		query.Add("uint", "9223372036854775807")
		query.Add("uint8", "255")
		query.Add("uint16", "65535")
		query.Add("uint32", "4294967295")
		query.Add("uint64", "18446744073709551615")
		query.Add("float32", "3.14159")
		query.Add("float64", "3.141592653589793")
		query.Add("bool", "true")
		query.Add("time", "2002-10-02T15:00:00.05Z")
		query.Add("url", "http://google.com")
	}

	// Prepare headers
	header := make(http.Header)
	if withHeaders {
		header.Add("X-String", "valueFromHeader")
		header.Add("X-Int", "42")
		header.Add("User-Agent", "BenchmarkAgent/1.0")
		header.Add("Accept", "application/json")
		header.Add("Referer", "http://example.com")
		header.Add("X-Custom-Header", "CustomValue")
	}

	// Prepare JSON body
	if withJSON {
		var err error
		bLen, bReader, err := toJSON(RequestPayload{
			Strings: []string{"string1", "string2", "string3"},
			Numbers: []int{1, 2, 3, 4, 5},
			Map: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			},
		})
		if err != nil {
			b.Fatal(err)
		}

		bodyLen = int64(bLen)
		body = bReader
		header.Add("Content-Type", decoder.ContentTypeJSON)
		header.Add("Content-Length", strconv.Itoa(bLen))
	}

	// Create request
	req := &http.Request{
		Method: method,
		URL: &url.URL{
			RawQuery: query.Encode(),
		},
		Header:        header,
		ContentLength: bodyLen,
		Body:          io.NopCloser(body),
	}

	return req, bodyLen
}

// Benchmark for JSON body parsing with small struct
func BenchmarkParse_JSONSmall(b *testing.B) {
	req, _ := prepareTestHTTPRequest(b, http.MethodPost, true, false, false)

	r := NewRoamer(WithDecoders(decoder.NewJSON()))
	var data SmallStruct

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := r.Parse(req, &data); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for JSON body parsing with large struct
func BenchmarkParse_JSONLarge(b *testing.B) {
	req, _ := prepareTestHTTPRequest(b, http.MethodPost, true, false, false)

	r := NewRoamer(WithDecoders(decoder.NewJSON()))
	var data LargeStruct

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := r.Parse(req, &data); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for query parameters parsing
func BenchmarkParse_QueryParams(b *testing.B) {
	req, _ := prepareTestHTTPRequest(b, http.MethodGet, false, false, true)

	r := NewRoamer(WithParsers(parser.NewQuery()))
	var data MediumStruct

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := r.Parse(req, &data); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for header parsing
func BenchmarkParse_Headers(b *testing.B) {
	req, _ := prepareTestHTTPRequest(b, http.MethodGet, false, true, false)

	r := NewRoamer(WithParsers(parser.NewHeader()))
	var data MediumStruct

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := r.Parse(req, &data); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for all parsers combined
func BenchmarkParse_AllParsers(b *testing.B) {
	req, _ := prepareTestHTTPRequest(b, http.MethodPost, true, true, true)

	r := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewHeader(), parser.NewQuery()),
	)
	var data LargeStruct

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := r.Parse(req, &data); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for all parsers with string formatter
func BenchmarkParse_WithFormatter(b *testing.B) {
	req, _ := prepareTestHTTPRequest(b, http.MethodPost, true, true, true)

	r := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewHeader(), parser.NewQuery()),
		WithFormatters(formatter.NewString()),
	)
	var data LargeStruct

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := r.Parse(req, &data); err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark for all parsers with skipFilled=true
func BenchmarkParse_SkipFilled(b *testing.B) {
	req, _ := prepareTestHTTPRequest(b, http.MethodPost, true, true, true)

	r := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewHeader(), parser.NewQuery()),
		WithSkipFilled(true),
	)

	// Pre-fill some fields
	data := LargeStruct{
		String: "prefilled",
		Int:    999,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := r.Parse(req, &data); err != nil {
			b.Fatal(err)
		}
	}
}

func TestRoamer_Parse_DefaultValue_Successfully(t *testing.T) {
	type validDefaultStruct struct {
		String      string   `query:"s" default:"default string"`
		Int         int      `query:"i" default:"123"`
		Float       float64  `query:"f" default:"123.45"`
		Bool        bool     `query:"b" default:"true"`
		StringSlice []string `query:"ss" default:"a,b,c"`
		IntSlice    []int    `query:"is" default:"1,2,3"`
		NoTag       string   `default:"no tag"`
		IntPtr      *int     `query:"iptr" default:"42"`
	}

	tests := []struct {
		name string
		url  string
		ptr  any
		want any
	}{
		{
			name: "all values from default",
			url:  "http://example.com",
			ptr:  &validDefaultStruct{},
			want: &validDefaultStruct{
				String:      "default string",
				Int:         123,
				Float:       123.45,
				Bool:        true,
				StringSlice: []string{"a", "b", "c"},
				IntSlice:    []int{1, 2, 3},
				NoTag:       "no tag",
				IntPtr:      func() *int { i := 42; return &i }(),
			},
		},
		{
			name: "all values from query, default ignored",
			url:  "http://example.com?s=query&i=999&f=9.99&b=false&ss=x,y&is=8,9",
			ptr:  &validDefaultStruct{},
			want: &validDefaultStruct{
				String:      "query",
				Int:         999,
				Float:       9.99,
				Bool:        false,
				StringSlice: []string{"x", "y"},
				IntSlice:    []int{8, 9},
				NoTag:       "no tag",                                 // No parser, so default is used
				IntPtr:      func() *int { i := 42; return &i }(), // No query value, so default is used
			},
		},
		{
			name: "partial from query, partial from default",
			url:  "http://example.com?s=query&b=false",
			ptr:  &validDefaultStruct{},
			want: &validDefaultStruct{
				String:      "query",
				Int:         123,
				Float:       123.45,
				Bool:        false,
				StringSlice: []string{"a", "b", "c"},
				IntSlice:    []int{1, 2, 3},
				NoTag:       "no tag",
				IntPtr:      func() *int { i := 42; return &i }(),
			},
		},
		{
			name: "pre-filled value is not overwritten by default",
			url:  "http://example.com",
			ptr: &validDefaultStruct{
				Int: 999, // Pre-filled
			},
			want: &validDefaultStruct{
				String:      "default string",
				Int:         999, // Should not be changed from 999 to 123
				Float:       123.45,
				Bool:        true,
				StringSlice: []string{"a", "b", "c"},
				IntSlice:    []int{1, 2, 3},
				NoTag:       "no tag",
				IntPtr:      func() *int { i := 42; return &i }(),
			},
		},
	}

	r := NewRoamer(WithParsers(parser.NewQuery()))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.url, nil)
			require.NoError(t, err)

			err = r.Parse(req, tt.ptr)

			require.NoError(t, err)
			require.Equal(t, tt.want, tt.ptr)
		})
	}
}

func TestRoamer_Parse_DefaultValue_Failure(t *testing.T) {
	type invalidDefaultStruct struct {
		InvalidInt int `query:"invalid" default:"abc"`
	}

	tests := []struct {
		name string
		url  string
		ptr  any
	}{
		{
			name: "invalid default value returns error",
			url:  "http://example.com",
			ptr:  &invalidDefaultStruct{},
		},
	}

	r := NewRoamer(WithParsers(parser.NewQuery()))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.url, nil)
			require.NoError(t, err)

			err = r.Parse(req, tt.ptr)

			require.Error(t, err)
		})
	}
}
