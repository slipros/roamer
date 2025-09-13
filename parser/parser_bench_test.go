package parser

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/slipros/roamer/internal/cache"
)

// BenchmarkQueryParser tests performance of query parameter parsing
func BenchmarkQueryParser_Parse(b *testing.B) {
	tests := []struct {
		name        string
		paramCount  int
		splitValues bool
	}{
		{"Small_5Params", 5, false},
		{"Medium_20Params", 20, false},
		{"Large_100Params", 100, false},
		{"WithSplitting_10Params", 10, true},
		{"WithSplitting_50Params", 50, true},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			parser := NewQuery()
			if !tt.splitValues {
				parser = NewQuery(WithDisabledSplit())
			}

			req := createQueryTestRequest(b, tt.paramCount, tt.splitValues)
			tag := reflect.StructTag(`query:"test_param"`)
			c := make(Cache, 10)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = parser.Parse(req, tag, c)
			}
		})
	}
}

// BenchmarkHeaderParser tests performance of HTTP header parsing
func BenchmarkHeaderParser_Parse(b *testing.B) {
	tests := []struct {
		name        string
		headerCount int
	}{
		{"Small_5Headers", 5},
		{"Medium_20Headers", 20},
		{"Large_100Headers", 100},
		{"XLarge_500Headers", 500},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			parser := NewHeader()
			req := createHeaderTestRequest(b, tt.headerCount)
			tag := reflect.StructTag(`header:"X-Test-Header"`)
			c := make(Cache, 10)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = parser.Parse(req, tag, c)
			}
		})
	}
}

// BenchmarkCookieParser tests performance of cookie parsing
func BenchmarkCookieParser_Parse(b *testing.B) {
	tests := []struct {
		name        string
		cookieCount int
	}{
		{"Small_5Cookies", 5},
		{"Medium_20Cookies", 20},
		{"Large_100Cookies", 100},
		{"XLarge_500Cookies", 500},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			parser := NewCookie()
			req := createCookieTestRequest(b, tt.cookieCount)
			tag := reflect.StructTag(`cookie:"test_cookie"`)
			c := make(Cache, 10)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = parser.Parse(req, tag, c)
			}
		})
	}
}

// BenchmarkQueryParser_SplitBehavior compares performance with/without splitting
func BenchmarkQueryParser_SplitBehavior(b *testing.B) {
	req := createQueryTestRequestWithSplitValues(b, 50)
	tag := reflect.StructTag(`query:"split_param"`)
	c := make(Cache, 10)

	b.Run("WithSplitting", func(b *testing.B) {
		parser := NewQuery() // splitting enabled by default
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _ = parser.Parse(req, tag, c)
		}
	})

	b.Run("WithoutSplitting", func(b *testing.B) {
		parser := NewQuery(WithDisabledSplit())
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _ = parser.Parse(req, tag, c)
		}
	})
}

// BenchmarkQueryParser_CustomSeparator tests performance with custom separators
func BenchmarkQueryParser_CustomSeparator(b *testing.B) {
	separators := []string{",", ";", "|", ":", "&"}

	for _, sep := range separators {
		b.Run(fmt.Sprintf("Separator_%s", strings.ReplaceAll(sep, "&", "amp")), func(b *testing.B) {
			parser := NewQuery(WithSplitSymbol(sep))
			req := createQueryTestRequestWithCustomSeparator(b, 20, sep)
			tag := reflect.StructTag(`query:"custom_sep_param"`)
			c := make(Cache, 10)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = parser.Parse(req, tag, c)
			}
		})
	}
}

// BenchmarkHeaderParser_CaseInsensitive tests performance of case-insensitive header lookup
func BenchmarkHeaderParser_CaseInsensitive(b *testing.B) {
	headerVariations := []string{
		"content-type",
		"Content-Type",
		"CONTENT-TYPE",
		"Content-type",
		"cOnTeNt-TyPe",
	}

	parser := NewHeader()
	req := createHeaderTestRequestWithVariations(b)
	c := make(Cache, 10)

	for _, headerName := range headerVariations {
		b.Run(fmt.Sprintf("Header_%s", strings.ReplaceAll(headerName, "-", "_")), func(b *testing.B) {
			tag := reflect.StructTag(fmt.Sprintf(`header:"%s"`, headerName))

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = parser.Parse(req, tag, c)
			}
		})
	}
}

// BenchmarkCookieParser_FindInMany tests performance when searching for specific cookie among many
func BenchmarkCookieParser_FindInMany(b *testing.B) {
	tests := []struct {
		name         string
		totalCookies int
		targetIndex  int // Where in the list the target cookie is
	}{
		{"First_Of_100", 100, 1},
		{"Middle_Of_100", 100, 50},
		{"Last_Of_100", 100, 100},
		{"First_Of_500", 500, 1},
		{"Middle_Of_500", 500, 250},
		{"Last_Of_500", 500, 500},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			parser := NewCookie()
			req := createCookieTestRequestWithTarget(b, tt.totalCookies, tt.targetIndex, "target_cookie")
			tag := reflect.StructTag(`cookie:"target_cookie"`)
			c := make(Cache, 10)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = parser.Parse(req, tag, c)
			}
		})
	}
}

// BenchmarkParsers_Concurrent tests concurrent access to parsers (thread safety)
func BenchmarkParsers_Concurrent(b *testing.B) {
	queryParser := NewQuery()
	headerParser := NewHeader()
	cookieParser := NewCookie()

	queryReq := createQueryTestRequest(b, 20, false)
	headerReq := createHeaderTestRequest(b, 20)
	cookieReq := createCookieTestRequest(b, 20)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		c := make(Cache, 10)
		queryTag := reflect.StructTag(`query:"test_param"`)
		headerTag := reflect.StructTag(`header:"X-Test-Header"`)
		cookieTag := reflect.StructTag(`cookie:"test_cookie"`)

		for pb.Next() {
			// Simulate concurrent parsing with different parsers
			_, _ = queryParser.Parse(queryReq, queryTag, c)
			_, _ = headerParser.Parse(headerReq, headerTag, c)
			_, _ = cookieParser.Parse(cookieReq, cookieTag, c)
		}
	})
}

// BenchmarkStructureCache_Impact compares parsing with and without structure cache
func BenchmarkStructureCache_Impact(b *testing.B) {
	// This benchmark tests how structure cache affects parsing performance
	// Note: In real usage, structure cache is used by Roamer, not individual parsers

	structureCache := &cache.Structure{}

	// Define a complex struct type for cache testing
	complexStructType := reflect.TypeOf(struct {
		QueryField1  string `query:"q1"`
		QueryField2  string `query:"q2"`
		QueryField3  string `query:"q3"`
		HeaderField1 string `header:"H1"`
		HeaderField2 string `header:"H2"`
		CookieField1 string `cookie:"c1"`
		CookieField2 string `cookie:"c2"`
	}{})

	b.Run("FirstAccess_CacheMiss", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// Fresh cache each time to simulate cache miss
			freshCache := &cache.Structure{}
			freshCache.Fields(complexStructType)
		}
	})

	b.Run("SubsequentAccess_CacheHit", func(b *testing.B) {
		// Warm up the cache
		structureCache.Fields(complexStructType)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			structureCache.Fields(complexStructType)
		}
	})
}

// Helper functions for creating test requests

func createQueryTestRequest(b *testing.B, paramCount int, withSplitValues bool) *http.Request {
	b.Helper()

	u, _ := url.Parse("https://example.com/test")
	q := u.Query()

	for i := 0; i < paramCount; i++ {
		paramName := fmt.Sprintf("param_%d", i)
		if withSplitValues && i%3 == 0 { // Every 3rd param has split values
			q.Add(paramName, "value1,value2,value3")
		} else {
			q.Add(paramName, fmt.Sprintf("value_%d", i))
		}
	}

	// Add the specific test parameter that benchmarks will look for
	q.Add("test_param", "test_value")

	u.RawQuery = q.Encode()
	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	return req
}

func createQueryTestRequestWithSplitValues(b *testing.B, paramCount int) *http.Request {
	b.Helper()

	u, _ := url.Parse("https://example.com/test")
	q := u.Query()

	// Add many params with split values
	for i := 0; i < paramCount; i++ {
		q.Add(fmt.Sprintf("param_%d", i), fmt.Sprintf("val_%d_1,val_%d_2,val_%d_3", i, i, i))
	}

	// Add the specific parameter that will be parsed
	q.Add("split_param", "split1,split2,split3,split4,split5")

	u.RawQuery = q.Encode()
	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	return req
}

func createQueryTestRequestWithCustomSeparator(b *testing.B, paramCount int, separator string) *http.Request {
	b.Helper()

	u, _ := url.Parse("https://example.com/test")
	q := u.Query()

	for i := 0; i < paramCount; i++ {
		q.Add(fmt.Sprintf("param_%d", i), fmt.Sprintf("value_%d", i))
	}

	// Add parameter with custom separator
	values := []string{"val1", "val2", "val3", "val4", "val5"}
	q.Add("custom_sep_param", strings.Join(values, separator))

	u.RawQuery = q.Encode()
	req, _ := http.NewRequest(http.MethodGet, u.String(), nil)
	return req
}

func createHeaderTestRequest(b *testing.B, headerCount int) *http.Request {
	b.Helper()

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/test", nil)

	for i := 0; i < headerCount; i++ {
		req.Header.Set(fmt.Sprintf("X-Header-%d", i), fmt.Sprintf("header_value_%d", i))
	}

	// Add the specific test header that benchmarks will look for
	req.Header.Set("X-Test-Header", "test_header_value")

	return req
}

func createHeaderTestRequestWithVariations(b *testing.B) *http.Request {
	b.Helper()

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/test", nil)

	// Add the content-type header (will be tested with different cases)
	req.Header.Set("Content-Type", "application/json")

	// Add some other headers for noise
	for i := 0; i < 10; i++ {
		req.Header.Set(fmt.Sprintf("X-Other-Header-%d", i), fmt.Sprintf("value_%d", i))
	}

	return req
}

func createCookieTestRequest(b *testing.B, cookieCount int) *http.Request {
	b.Helper()

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/test", nil)

	for i := 0; i < cookieCount; i++ {
		cookie := &http.Cookie{
			Name:  fmt.Sprintf("cookie_%d", i),
			Value: fmt.Sprintf("cookie_value_%d", i),
		}
		req.AddCookie(cookie)
	}

	// Add the specific test cookie that benchmarks will look for
	testCookie := &http.Cookie{
		Name:  "test_cookie",
		Value: "test_cookie_value",
	}
	req.AddCookie(testCookie)

	return req
}

func createCookieTestRequestWithTarget(b *testing.B, totalCookies, targetIndex int, targetName string) *http.Request {
	b.Helper()

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/test", nil)

	for i := 1; i <= totalCookies; i++ {
		var cookie *http.Cookie
		if i == targetIndex {
			cookie = &http.Cookie{
				Name:  targetName,
				Value: "target_value",
			}
		} else {
			cookie = &http.Cookie{
				Name:  fmt.Sprintf("cookie_%d", i),
				Value: fmt.Sprintf("value_%d", i),
			}
		}
		req.AddCookie(cookie)
	}

	return req
}
