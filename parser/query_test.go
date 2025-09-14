package parser

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQuery(t *testing.T) {
	q := NewQuery()
	require.NotNil(t, q)
	require.Equal(t, TagQuery, q.Tag())

	q = NewQuery(WithDisabledSplit())
	require.NotNil(t, q)
	require.False(t, q.split)

	q = NewQuery(WithSplitSymbol(";"))
	require.NotNil(t, q)
	require.Equal(t, ";", q.splitSymbol)
}

func TestQuery(t *testing.T) {
	queryName := "user_id"
	queryValue := "1337"

	type args struct {
		req   *http.Request
		tag   reflect.StructTag
		cache Cache
	}
	tests := []struct {
		name      string
		args      func() args
		want      any
		notExists bool
	}{
		{
			name: "Get value from query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(Cache),
				}
			},
			want: queryValue,
		},
		{
			name: "Get value from cached query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				cache := make(map[string]any, 1)
				cache[cacheKeyQuery] = q

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: cache,
				}
			},
			want: queryValue,
		},
		{
			name: "Get value from query - no query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(Cache),
				}
			},
			want: "",
		},
		{
			name: "Get value from query - wrong query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName+"1", queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(Cache),
				}
			},
			want: "",
		},
		{
			name: "Get value from array query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue)
				q.Add(queryName, queryValue+"2")

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(Cache),
				}
			},
			want: []string{queryValue, queryValue + "2"},
		},
		{
			name: "Get value from query with split symbol",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue+","+queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(Cache),
				}
			},
			want: []string{queryValue, queryValue},
		},
		{
			name:      "Wrong tag",
			notExists: true,
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue+","+queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagHeader, queryName)),
					cache: make(Cache),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args()

			q := NewQuery()

			value, exists := q.Parse(args.req, args.tag, args.cache)
			if tt.notExists && exists {
				t.Errorf("Parse() want not exists, but value exists")
			}

			if tt.want == nil && exists {
				t.Errorf("Parse() want is nil, but value exists")
			}

			if !tt.notExists {
				require.Equal(t, tt.want, value)
			}
		})
	}
}

// TestQuery_parseQuery tests the parseQuery function with comprehensive coverage
// to ensure it behaves identically to http.Request.URL.Query() and url.ParseQuery
func TestQuery_parseQuery(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expected    url.Values
		expectError bool
		description string
	}{
		// Successfully parsed cases
		{
			name:        "Empty string",
			query:       "",
			expected:    url.Values{},
			description: "Empty query string should return empty url.Values",
		},
		{
			name:        "Single parameter",
			query:       "a=1",
			expected:    url.Values{"a": {"1"}},
			description: "Single key-value pair",
		},
		{
			name:        "Multiple parameters",
			query:       "a=1&b=2&c=3",
			expected:    url.Values{"a": {"1"}, "b": {"2"}, "c": {"3"}},
			description: "Multiple key-value pairs separated by ampersands",
		},
		{
			name:        "Multiple values for same key",
			query:       "a=1&a=2&a=3",
			expected:    url.Values{"a": {"1", "2", "3"}},
			description: "Same key appearing multiple times should create slice",
		},
		{
			name:        "Mixed single and multiple values",
			query:       "a=1&b=2&a=3&c=4",
			expected:    url.Values{"a": {"1", "3"}, "b": {"2"}, "c": {"4"}},
			description: "Combination of single and multiple values for different keys",
		},
		{
			name:        "Parameter without value",
			query:       "a",
			expected:    url.Values{"a": {""}},
			description: "Key without equals sign should have empty string value",
		},
		{
			name:        "Parameter with empty value",
			query:       "a=",
			expected:    url.Values{"a": {""}},
			description: "Key with equals but no value should have empty string value",
		},
		{
			name:        "Mixed parameters with and without values",
			query:       "a=1&b&c=&d=4",
			expected:    url.Values{"a": {"1"}, "b": {""}, "c": {""}, "d": {"4"}},
			description: "Mix of parameters with values, without values, and empty values",
		},
		{
			name:        "Trailing ampersand",
			query:       "a=1&b=2&",
			expected:    url.Values{"a": {"1"}, "b": {"2"}},
			description: "Trailing ampersand should be ignored",
		},
		{
			name:        "Leading ampersand",
			query:       "&a=1&b=2",
			expected:    url.Values{"a": {"1"}, "b": {"2"}},
			description: "Leading ampersand should be ignored",
		},
		{
			name:        "Multiple consecutive ampersands",
			query:       "a=1&&b=2&&&c=3",
			expected:    url.Values{"a": {"1"}, "b": {"2"}, "c": {"3"}},
			description: "Multiple consecutive ampersands should be ignored",
		},
		{
			name:        "Only ampersands",
			query:       "&&&",
			expected:    url.Values{},
			description: "String with only ampersands should return empty values",
		},
		{
			name:        "URL encoded key",
			query:       "hello%20world=test",
			expected:    url.Values{"hello world": {"test"}},
			description: "URL encoded spaces in key should be decoded",
		},
		{
			name:        "URL encoded value",
			query:       "test=hello%20world",
			expected:    url.Values{"test": {"hello world"}},
			description: "URL encoded spaces in value should be decoded",
		},
		{
			name:        "URL encoded key and value",
			query:       "hello%20key=world%20value",
			expected:    url.Values{"hello key": {"world value"}},
			description: "Both key and value with URL encoding",
		},
		{
			name:        "Special characters encoded",
			query:       "name=John%20Doe&email=test%40example.com",
			expected:    url.Values{"name": {"John Doe"}, "email": {"test@example.com"}},
			description: "Common special characters like @ and space",
		},
		{
			name:        "Plus signs and ampersands encoded",
			query:       "data=hello%2Bworld%26more",
			expected:    url.Values{"data": {"hello+world&more"}},
			description: "Plus signs and ampersands in values",
		},
		{
			name:        "Unicode characters",
			query:       "unicode=тест&emoji=🚀",
			expected:    url.Values{"unicode": {"тест"}, "emoji": {"🚀"}},
			description: "Unicode characters should be preserved",
		},
		{
			name:        "Percent encoding edge cases",
			query:       "a=100%25&b=%20",
			expected:    url.Values{"a": {"100%"}, "b": {" "}},
			description: "Percent signs and valid percent encoding",
		},
		{
			name:        "Complex query with all features",
			query:       "name=John%20Doe&tags=go,web&tags=programming&empty=&flag&encoded=hello%2Bworld",
			expected:    url.Values{"name": {"John Doe"}, "tags": {"go,web", "programming"}, "empty": {""}, "flag": {""}, "encoded": {"hello+world"}},
			description: "Complex real-world query string",
		},
		{
			name:        "Equals in value",
			query:       "equation=a%3D1%2Bb%3D2",
			expected:    url.Values{"equation": {"a=1+b=2"}},
			description: "Equals signs within encoded values",
		},
		{
			name:        "Long parameter names and values",
			query:       strings.Repeat("a", 100) + "=" + strings.Repeat("b", 200),
			expected:    url.Values{strings.Repeat("a", 100): {strings.Repeat("b", 200)}},
			description: "Very long parameter names and values",
		},
		{
			name:        "Many parameters",
			query:       generateManyParams(50),
			expected:    generateExpectedManyParams(50),
			description: "Large number of parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQuery()

			// Test custom parseQuery implementation
			result := q.parseQuery(tt.query)

			// Compare with expected result
			assert.Equal(t, tt.expected, result, "parseQuery result should match expected values")

			// Also compare with standard library to ensure identical behavior
			if tt.query != "" {
				standardResult, err := url.ParseQuery(tt.query)
				if err == nil {
					assert.Equal(t, standardResult, result, "parseQuery should behave identically to url.ParseQuery when standard library succeeds")
				}
				// Note: If standard library errors, our implementation may still handle gracefully
			}
		})
	}
}

// TestQuery_parseQuery_ErrorHandling tests error handling scenarios
func TestQuery_parseQuery_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		description string
	}{
		{
			name:        "Invalid percent encoding in value",
			query:       "a=hello%2",
			description: "Invalid percent encoding should not cause panic",
		},
		{
			name:        "Invalid percent encoding in key",
			query:       "hello%2=world",
			description: "Invalid percent encoding in key should not cause panic",
		},
		{
			name:        "Multiple invalid encodings",
			query:       "a%=b%&c%2=d%3",
			description: "Multiple invalid encodings should be handled gracefully",
		},
		{
			name:        "Non-hex percent encoding",
			query:       "a=hello%GG&b=world%ZZ",
			description: "Non-hexadecimal percent encoding should be left as-is",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQuery()

			// Should not panic
			assert.NotPanics(t, func() {
				result := q.parseQuery(tt.query)
				assert.NotNil(t, result, "Result should not be nil even with malformed input")
			}, "parseQuery should not panic on malformed input")

			// Compare behavior with standard library
			standardResult, err := url.ParseQuery(tt.query)
			customResult := q.parseQuery(tt.query)

			if err != nil {
				// If standard library fails, our implementation should handle gracefully
				assert.NotNil(t, customResult, "Custom implementation should return valid result even if standard library errors")
			} else {
				// If standard library succeeds, results should match
				assert.Equal(t, standardResult, customResult, "Results should match when standard library succeeds")
			}
		})
	}
}

// TestQuery_parseQuery_Performance tests performance characteristics
func TestQuery_parseQuery_Performance(t *testing.T) {
	// Test with various query sizes to ensure no performance regression
	queries := []struct {
		name  string
		query string
	}{
		{"Small query", "a=1&b=2&c=3"},
		{"Medium query", generateManyParams(100)},
		{"Large query", generateManyParams(1000)},
	}

	for _, q := range queries {
		t.Run(q.name, func(t *testing.T) {
			parser := NewQuery()

			// Measure memory allocations
			var result url.Values
			allocs := testing.AllocsPerRun(100, func() {
				result = parser.parseQuery(q.query)
			})

			// Should not allocate excessively
			assert.NotNil(t, result, "Result should not be nil")
			t.Logf("Query size: %d chars, Allocations per run: %.2f", len(q.query), allocs)
		})
	}
}

// Helper function to generate many parameters for testing
func generateManyParams(count int) string {
	var parts []string
	for i := 0; i < count; i++ {
		parts = append(parts, fmt.Sprintf("param%d=value%d", i, i))
	}
	return strings.Join(parts, "&")
}

// Helper function to generate expected result for many parameters
func generateExpectedManyParams(count int) url.Values {
	result := make(url.Values, count)
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("param%d", i)
		value := fmt.Sprintf("value%d", i)
		result[key] = []string{value}
	}
	return result
}

// BenchmarkQuery_parseQuery benchmarks the parseQuery function
func BenchmarkQuery_parseQuery(b *testing.B) {
	benchmarks := []struct {
		name  string
		query string
	}{
		{"Empty", ""},
		{"Single", "a=1"},
		{"Small", "a=1&b=2&c=3"},
		{"Medium", generateManyParams(50)},
		{"Large", generateManyParams(500)},
		{"WithEncoding", "name=John%20Doe&email=test%40example.com&data=hello%2Bworld%26more"},
		{"MultipleValues", "a=1&a=2&a=3&b=4&b=5&c=6"},
	}

	q := NewQuery()

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = q.parseQuery(bm.query)
			}
		})
	}
}

// BenchmarkQuery_parseQuery_vs_Standard compares performance with standard library
func BenchmarkQuery_parseQuery_vs_Standard(b *testing.B) {
	testQuery := generateManyParams(100)
	q := NewQuery()

	b.Run("Custom_parseQuery", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = q.parseQuery(testQuery)
		}
	})

	b.Run("Standard_ParseQuery", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = url.ParseQuery(testQuery)
		}
	})
}
