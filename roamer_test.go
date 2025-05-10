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
	skipFilled bool
	decoders   Decoders
	parsers    Parsers
	formatters Formatters
}

type args struct {
	req *http.Request
	ptr any
}

type testStruct struct {
	String string `json:"string" header:"X-String" query:"string"`
	Int    int    `json:"int" header:"X-Int" query:"int"`
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
						decoders: Decoders{
							"application/json": decoder.NewJSON(),
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
						parsers: Parsers{
							"query": parser.NewQuery(),
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
						parsers: Parsers{
							"header": parser.NewHeader(),
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
				// Create request with all data sources
				req, _ := http.NewRequest(http.MethodPost, "http://example.com?string=queryValue&int=456", bytes.NewReader(testJSON))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Content-Length", strconv.Itoa(len(testJSON)))
				req.Header.Set("X-String", "headerValue")
				req.Header.Set("X-Int", "789")

				// Create struct to fill
				target := &testStruct{}

				// Configure roamer with parsers and decoders
				return fields{
						decoders: Decoders{
							"application/json": decoder.NewJSON(),
						},
						parsers: Parsers{
							"header": parser.NewHeader(),
							"query":  parser.NewQuery(),
						},
					}, args{
						req: req,
						ptr: target,
					}
			},
			verify: func(t *testing.T, result any) {
				target, ok := result.(*testStruct)
				require.True(t, ok)
				// Priority should be given to the first successful parser
				// Check actual priority order
				assert.Equal(t, "headerValue", target.String)
				assert.Equal(t, 789, target.Int)
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
						decoders: Decoders{
							"application/json": decoder.NewJSON(),
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
						decoders: Decoders{
							"application/json": decoder.NewJSON(),
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
						parsers: Parsers{
							"header": parser.NewHeader(),
						},
						formatters: Formatters{
							"string": formatter.NewString(),
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
						skipFilled: true,
						parsers: Parsers{
							"query": parser.NewQuery(),
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
						skipFilled: false,
						parsers: Parsers{
							"query": parser.NewQuery(),
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

			r := &Roamer{
				skipFilled:    f.skipFilled,
				decoders:      f.decoders,
				parsers:       f.parsers,
				formatters:    f.formatters,
				hasParsers:    len(f.parsers) > 0,
				hasDecoders:   len(f.decoders) > 0,
				hasFormatters: len(f.formatters) > 0,
			}

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

			r := &Roamer{
				skipFilled:    f.skipFilled,
				decoders:      f.decoders,
				parsers:       f.parsers,
				formatters:    f.formatters,
				hasParsers:    len(f.parsers) > 0,
				hasDecoders:   len(f.decoders) > 0,
				hasFormatters: len(f.formatters) > 0,
			}

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

func BenchmarkParse_With_Body_Header_Query(b *testing.B) {
	toJSON := func(v any) (int, io.Reader, error) {
		var buffer bytes.Buffer
		if err := json.NewEncoder(&buffer).Encode(&v); err != nil {
			return 0, nil, err
		}

		return buffer.Len(), &buffer, nil
	}

	query := make(url.Values)
	query.Add("int", "9223372036854775807")
	query.Add("int8", "127")
	query.Add("int16", "32767")
	query.Add("int32", "2147483647")
	query.Add("int64", "9223372036854775807")
	query.Add("time", "2002-10-02T15:00:00.05Z")
	query.Add("url", "http://google.com")

	header := make(http.Header)
	header.Add("User-Agent", "agent 1337")

	bodyLen, body, err := toJSON(
		struct {
			Strings []string `json:"strings"`
		}{
			Strings: []string{"1", "2"},
		},
	)
	if err != nil {
		b.Fatal(err)
	}

	header.Add("Content-Type", decoder.ContentTypeJSON)
	header.Add("Content-Length", strconv.Itoa(bodyLen))

	type Data struct {
		Int   int       `query:"int"`
		Int8  int8      `query:"int8"`
		Int32 int32     `query:"int32"`
		Int64 int64     `query:"int64"`
		Time  time.Time `query:"time"`
		Url   url.URL   `query:"url"`

		UserAgent string   `header:"User-Agent"`
		Strings   []string `json:"strings"`
	}

	req := http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			RawQuery: query.Encode(),
		},
		Header:        header,
		ContentLength: int64(bodyLen),
		Body:          io.NopCloser(body),
	}

	r := NewRoamer(
		WithSkipFilled(false),
		WithParsers(parser.NewHeader(), parser.NewQuery()),
	)

	var d Data

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := r.Parse(&req, &d); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParse_With_Body_Header_Query_FastStructFieldParser(b *testing.B) {
	toJSON := func(v any) (int, io.Reader, error) {
		var buffer bytes.Buffer
		if err := json.NewEncoder(&buffer).Encode(&v); err != nil {
			return 0, nil, err
		}

		return buffer.Len(), &buffer, nil
	}

	query := make(url.Values, 7)
	query.Add("int", "9223372036854775807")
	query.Add("int8", "127")
	query.Add("int16", "32767")
	query.Add("int32", "2147483647")
	query.Add("int64", "9223372036854775807")
	query.Add("time", "2002-10-02T15:00:00.05Z")
	query.Add("url", "http://google.com")

	header := make(http.Header)
	header.Add("User-Agent", "agent 1337")

	bodyLen, body, err := toJSON(
		struct {
			Strings []string `json:"strings"`
		}{
			Strings: []string{"1", "2"},
		},
	)
	if err != nil {
		b.Fatal(err)
	}

	header.Add("Content-Type", decoder.ContentTypeJSON)
	header.Add("Content-Length", strconv.Itoa(bodyLen))

	type Data struct {
		Int   int       `query:"int"`
		Int8  int8      `query:"int8"`
		Int32 int32     `query:"int32"`
		Int64 int64     `query:"int64"`
		Time  time.Time `query:"time"`
		Url   url.URL   `query:"url"`

		UserAgent string   `header:"User-Agent"`
		Strings   []string `json:"strings"`
	}

	req := http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			RawQuery: query.Encode(),
		},
		Header:        header,
		ContentLength: int64(bodyLen),
		Body:          io.NopCloser(body),
	}

	r := NewRoamer(
		WithSkipFilled(false),
		WithParsers(parser.NewHeader(), parser.NewQuery()),
		WithExperimentalFastStructFieldParser(),
	)

	var d Data

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := r.Parse(&req, &d); err != nil {
			b.Fatal(err)
		}
	}
}
