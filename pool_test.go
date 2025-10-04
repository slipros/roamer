package roamer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/formatter"
	"github.com/slipros/roamer/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewParseWithPool_Successfully tests successful scenarios for NewParseWithPool
func TestNewParseWithPool_Successfully(t *testing.T) {
	t.Parallel()

	t.Run("basic parsing with pool - JSON body", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Age   int    `json:"age"`
		}

		jsonData := map[string]any{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
		}
		jsonBytes, _ := json.Marshal(jsonData)

		req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes)))

		r := NewRoamer(WithDecoders(decoder.NewJSON()))
		parseFunc := NewParseWithPool[TestStruct](r)

		err := parseFunc(req, func(data *TestStruct) error {
			assert.Equal(t, "John Doe", data.Name)
			assert.Equal(t, "john@example.com", data.Email)
			assert.Equal(t, 30, data.Age)
			return nil
		})

		require.NoError(t, err)
	})

	t.Run("parsing with pool - query parameters", func(t *testing.T) {
		t.Parallel()

		type QueryStruct struct {
			UserID string `query:"user_id"`
			Sort   string `query:"sort"`
			Limit  int    `query:"limit"`
		}

		req, _ := http.NewRequest(http.MethodGet, "http://example.com?user_id=123&sort=desc&limit=50", nil)
		r := NewRoamer(WithParsers(parser.NewQuery()))
		parseFunc := NewParseWithPool[QueryStruct](r)

		err := parseFunc(req, func(data *QueryStruct) error {
			assert.Equal(t, "123", data.UserID)
			assert.Equal(t, "desc", data.Sort)
			assert.Equal(t, 50, data.Limit)
			return nil
		})

		require.NoError(t, err)
	})

	t.Run("parsing with pool - headers", func(t *testing.T) {
		t.Parallel()

		type HeaderStruct struct {
			UserAgent     string `header:"User-Agent"`
			Authorization string `header:"Authorization"`
			ContentType   string `header:"Content-Type"`
		}

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		req.Header.Set("User-Agent", "TestAgent/1.0")
		req.Header.Set("Authorization", "Bearer token123")
		req.Header.Set("Content-Type", "application/json")

		r := NewRoamer(WithParsers(parser.NewHeader()))
		parseFunc := NewParseWithPool[HeaderStruct](r)

		err := parseFunc(req, func(data *HeaderStruct) error {
			assert.Equal(t, "TestAgent/1.0", data.UserAgent)
			assert.Equal(t, "Bearer token123", data.Authorization)
			assert.Equal(t, "application/json", data.ContentType)
			return nil
		})

		require.NoError(t, err)
	})

	t.Run("parsing with pool - multi-source data", func(t *testing.T) {
		t.Parallel()

		type MultiSourceStruct struct {
			Name   string `json:"name"`
			UserID string `query:"user_id"`
			Token  string `header:"Authorization"`
		}

		jsonData := map[string]any{"name": "Alice"}
		jsonBytes, _ := json.Marshal(jsonData)

		req, _ := http.NewRequest(http.MethodPost, "http://example.com?user_id=456", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes)))
		req.Header.Set("Authorization", "Bearer xyz")

		r := NewRoamer(
			WithDecoders(decoder.NewJSON()),
			WithParsers(parser.NewQuery(), parser.NewHeader()),
		)
		parseFunc := NewParseWithPool[MultiSourceStruct](r)

		err := parseFunc(req, func(data *MultiSourceStruct) error {
			assert.Equal(t, "Alice", data.Name)
			assert.Equal(t, "456", data.UserID)
			assert.Equal(t, "Bearer xyz", data.Token)
			return nil
		})

		require.NoError(t, err)
	})

	t.Run("parsing with pool - with formatters", func(t *testing.T) {
		t.Parallel()

		type FormatterStruct struct {
			Name  string `json:"name" string:"trim_space,lower"`
			Email string `json:"email" string:"trim_space"`
		}

		jsonData := map[string]any{
			"name":  "  JOHN DOE  ",
			"email": "  john@example.com  ",
		}
		jsonBytes, _ := json.Marshal(jsonData)

		req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes)))

		r := NewRoamer(
			WithDecoders(decoder.NewJSON()),
			WithFormatters(formatter.NewString()),
		)
		parseFunc := NewParseWithPool[FormatterStruct](r)

		err := parseFunc(req, func(data *FormatterStruct) error {
			assert.Equal(t, "john doe", data.Name)
			assert.Equal(t, "john@example.com", data.Email)
			return nil
		})

		require.NoError(t, err)
	})

	t.Run("callback modifies data successfully", func(t *testing.T) {
		t.Parallel()

		type ModifyStruct struct {
			Counter int `json:"counter"`
		}

		jsonData := map[string]any{"counter": 5}
		jsonBytes, _ := json.Marshal(jsonData)

		req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes)))

		r := NewRoamer(WithDecoders(decoder.NewJSON()))
		parseFunc := NewParseWithPool[ModifyStruct](r)

		var doubledValue int
		err := parseFunc(req, func(data *ModifyStruct) error {
			assert.Equal(t, 5, data.Counter)
			data.Counter *= 2
			doubledValue = data.Counter
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, 10, doubledValue)
	})
}

// TestNewParseWithPool_Failure tests failure scenarios for NewParseWithPool
func TestNewParseWithPool_Failure(t *testing.T) {
	t.Parallel()

	t.Run("nil roamer instance", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			Name string `json:"name"`
		}

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		parseFunc := NewParseWithPool[TestStruct](nil)
		callback := func(data *TestStruct) error { return nil }

		err := parseFunc(req, callback)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "roamer")
	})

	t.Run("nil callback function", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			Name string `json:"name"`
		}

		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		r := NewRoamer()
		parseFunc := NewParseWithPool[TestStruct](r)

		err := parseFunc(req, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "callback")
	})

	t.Run("callback returns error", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			Name string `json:"name"`
		}

		jsonData := map[string]any{"name": "test"}
		jsonBytes, _ := json.Marshal(jsonData)

		req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes)))

		r := NewRoamer(WithDecoders(decoder.NewJSON()))
		parseFunc := NewParseWithPool[TestStruct](r)

		expectedErr := errors.New("callback error")
		err := parseFunc(req, func(data *TestStruct) error {
			return expectedErr
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "callback error")
	})

	t.Run("parsing error propagates", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			Name string `json:"name"`
		}

		// Invalid JSON
		req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte("{invalid json}")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", "14")

		r := NewRoamer(WithDecoders(decoder.NewJSON()))
		parseFunc := NewParseWithPool[TestStruct](r)

		err := parseFunc(req, func(data *TestStruct) error {
			return nil
		})

		require.Error(t, err)
	})
}

// TestNewParseWithPool_FieldsZeroed tests that fields are properly zeroed before returning to pool
func TestNewParseWithPool_FieldsZeroed(t *testing.T) {
	type TestStruct struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Active bool   `json:"active"`
		Count  int    `json:"count"`
	}

	r := NewRoamer(WithDecoders(decoder.NewJSON()))
	parseFunc := NewParseWithPool[TestStruct](r)

	// First request with full data
	jsonData1 := map[string]any{
		"name":   "John Doe",
		"email":  "john@example.com",
		"active": true,
		"count":  42,
	}
	jsonBytes1, _ := json.Marshal(jsonData1)
	req1, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes1)))

	err := parseFunc(req1, func(data *TestStruct) error {
		assert.Equal(t, "John Doe", data.Name)
		assert.Equal(t, "john@example.com", data.Email)
		assert.True(t, data.Active)
		assert.Equal(t, 42, data.Count)
		return nil
	})
	require.NoError(t, err)

	// Second request with partial data - should not have remnants from first request
	jsonData2 := map[string]any{
		"name": "Jane Smith",
	}
	jsonBytes2, _ := json.Marshal(jsonData2)
	req2, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes2)))

	err = parseFunc(req2, func(data *TestStruct) error {
		assert.Equal(t, "Jane Smith", data.Name)
		assert.Empty(t, data.Email, "Email should be zero value, not remnant from previous use")
		assert.False(t, data.Active, "Active should be zero value, not remnant from previous use")
		assert.Zero(t, data.Count, "Count should be zero value, not remnant from previous use")
		return nil
	})
	require.NoError(t, err)
}

// TestNewParseWithPool_Concurrent tests concurrent safety of pool-based parsing
func TestNewParseWithPool_Concurrent(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		ID    int    `json:"id"`
		Email string `json:"email"`
	}

	r := NewRoamer(WithDecoders(decoder.NewJSON()))
	parseFunc := NewParseWithPool[TestStruct](r)

	const numGoroutines = 50
	const numIterationsPerGoroutine = 20

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numIterationsPerGoroutine)
	successCount := new(atomic.Int64)

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < numIterationsPerGoroutine; i++ {
				jsonData := map[string]any{
					"name":  fmt.Sprintf("User%d_%d", goroutineID, i),
					"id":    goroutineID*1000 + i,
					"email": fmt.Sprintf("user%d_%d@example.com", goroutineID, i),
				}
				jsonBytes, _ := json.Marshal(jsonData)

				req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes)))

				err := parseFunc(req, func(data *TestStruct) error {
					// Verify data integrity
					expectedName := fmt.Sprintf("User%d_%d", goroutineID, i)
					expectedID := goroutineID*1000 + i
					expectedEmail := fmt.Sprintf("user%d_%d@example.com", goroutineID, i)

					if data.Name != expectedName || data.ID != expectedID || data.Email != expectedEmail {
						return fmt.Errorf("data mismatch: got Name=%s ID=%d Email=%s, want Name=%s ID=%d Email=%s",
							data.Name, data.ID, data.Email, expectedName, expectedID, expectedEmail)
					}

					successCount.Add(1)
					return nil
				})

				if err != nil {
					errors <- fmt.Errorf("goroutine %d, iteration %d: %w", goroutineID, i, err)
					return
				}
			}
		}(g)
	}

	wg.Wait()
	close(errors)

	var errorList []error
	for err := range errors {
		errorList = append(errorList, err)
	}

	require.Empty(t, errorList, "Concurrent parsing should not produce errors")

	expectedSuccessCount := int64(numGoroutines * numIterationsPerGoroutine)
	assert.Equal(t, expectedSuccessCount, successCount.Load(), "All operations should succeed")
}

// TestNewParseWithPool_HighContention tests pool behavior under high contention
func TestNewParseWithPool_HighContention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high contention test in short mode")
	}

	type ComplexStruct struct {
		Field1 string  `json:"field1"`
		Field2 int     `json:"field2"`
		Field3 bool    `json:"field3"`
		Field4 string  `query:"param1"`
		Field5 string  `header:"X-Custom"`
		Field6 float64 `json:"field6"`
		Field7 string  `json:"field7"`
		Field8 int64   `json:"field8"`
	}

	r := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewQuery(), parser.NewHeader()),
	)
	parseFunc := NewParseWithPool[ComplexStruct](r)

	const numGoroutines = 100
	const numIterationsPerGoroutine = 50

	var wg sync.WaitGroup
	startBarrier := make(chan struct{})

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Wait for all goroutines to be ready
			<-startBarrier

			for i := 0; i < numIterationsPerGoroutine; i++ {
				jsonData := map[string]any{
					"field1": fmt.Sprintf("value%d", i),
					"field2": i,
					"field3": i%2 == 0,
					"field6": float64(i) * 1.5,
					"field7": "test",
					"field8": int64(i * 1000),
				}
				jsonBytes, _ := json.Marshal(jsonData)

				req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://example.com?param1=query%d", i), bytes.NewReader(jsonBytes))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes)))
				req.Header.Set("X-Custom", fmt.Sprintf("header%d", i))

				_ = parseFunc(req, func(data *ComplexStruct) error {
					// Verify some fields
					assert.Equal(t, fmt.Sprintf("value%d", i), data.Field1)
					assert.Equal(t, i, data.Field2)
					return nil
				})
			}
		}(g)
	}

	// Start all goroutines simultaneously to create maximum contention
	close(startBarrier)
	wg.Wait()
}

// TestNewParseWithPool_NoDataLeakage tests that data doesn't leak between requests
func TestNewParseWithPool_NoDataLeakage(t *testing.T) {
	type SensitiveStruct struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Token    string `json:"token"`
		Secret   string `json:"secret"`
	}

	r := NewRoamer(WithDecoders(decoder.NewJSON()))
	parseFunc := NewParseWithPool[SensitiveStruct](r)

	// First request with sensitive data
	sensitiveData := map[string]any{
		"username": "admin",
		"password": "super_secret_password",
		"token":    "sensitive_token_12345",
		"secret":   "top_secret_data",
	}
	jsonBytes1, _ := json.Marshal(sensitiveData)
	req1, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes1)))

	err := parseFunc(req1, func(data *SensitiveStruct) error {
		assert.Equal(t, "admin", data.Username)
		assert.Equal(t, "super_secret_password", data.Password)
		return nil
	})
	require.NoError(t, err)

	// Second request with different data - ensure no leakage
	normalData := map[string]any{
		"username": "user123",
	}
	jsonBytes2, _ := json.Marshal(normalData)
	req2, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes2)))

	err = parseFunc(req2, func(data *SensitiveStruct) error {
		assert.Equal(t, "user123", data.Username)
		assert.Empty(t, data.Password, "Password should be zeroed, not leaked from previous request")
		assert.Empty(t, data.Token, "Token should be zeroed, not leaked from previous request")
		assert.Empty(t, data.Secret, "Secret should be zeroed, not leaked from previous request")
		return nil
	})
	require.NoError(t, err)
}

// BenchmarkNewParseWithPool_vs_Parse compares performance of pooled vs non-pooled parsing
func BenchmarkNewParseWithPool_vs_Parse(b *testing.B) {
	type BenchStruct struct {
		Name   string `json:"name"`
		Email  string `json:"email"`
		Age    int    `json:"age"`
		Active bool   `json:"active"`
	}

	jsonData := map[string]any{
		"name":   "John Doe",
		"email":  "john@example.com",
		"age":    30,
		"active": true,
	}
	jsonBytes, _ := json.Marshal(jsonData)

	r := NewRoamer(WithDecoders(decoder.NewJSON()))

	b.Run("WithPool", func(b *testing.B) {
		parseFunc := NewParseWithPool[BenchStruct](r)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes)))

			_ = parseFunc(req, func(data *BenchStruct) error {
				return nil
			})
		}
	})

	b.Run("WithoutPool", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes)))

			var data BenchStruct
			_ = r.Parse(req, &data)
		}
	})

	b.Run("GenericParse", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes)))

			_, _ = Parse[BenchStruct](r, req)
		}
	})
}

// BenchmarkNewParseWithPool_Concurrent benchmarks concurrent pool usage
func BenchmarkNewParseWithPool_Concurrent(b *testing.B) {
	type BenchStruct struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	jsonData := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}
	jsonBytes, _ := json.Marshal(jsonData)

	r := NewRoamer(WithDecoders(decoder.NewJSON()))
	parseFunc := NewParseWithPool[BenchStruct](r)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes)))

			_ = parseFunc(req, func(data *BenchStruct) error {
				return nil
			})
		}
	})
}

// BenchmarkNewParseWithPool_SmallStruct benchmarks pool with small structs
func BenchmarkNewParseWithPool_SmallStruct(b *testing.B) {
	type SmallStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	jsonData := map[string]any{"id": 123, "name": "test"}
	jsonBytes, _ := json.Marshal(jsonData)

	r := NewRoamer(WithDecoders(decoder.NewJSON()))
	parseFunc := NewParseWithPool[SmallStruct](r)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes)))

		_ = parseFunc(req, func(data *SmallStruct) error {
			return nil
		})
	}
}

// BenchmarkNewParseWithPool_LargeStruct benchmarks pool with large structs
func BenchmarkNewParseWithPool_LargeStruct(b *testing.B) {
	type LargeStruct struct {
		Field1  string         `json:"field1"`
		Field2  string         `json:"field2"`
		Field3  string         `json:"field3"`
		Field4  int            `json:"field4"`
		Field5  int            `json:"field5"`
		Field6  int64          `json:"field6"`
		Field7  float64        `json:"field7"`
		Field8  bool           `json:"field8"`
		Field9  time.Time      `json:"field9"`
		Field10 []string       `json:"field10"`
		Field11 map[string]any `json:"field11"`
	}

	jsonData := map[string]any{
		"field1":  "value1",
		"field2":  "value2",
		"field3":  "value3",
		"field4":  123,
		"field5":  456,
		"field6":  int64(789),
		"field7":  3.14159,
		"field8":  true,
		"field9":  time.Now().Format(time.RFC3339),
		"field10": []string{"a", "b", "c"},
		"field11": map[string]any{"key": "value"},
	}
	jsonBytes, _ := json.Marshal(jsonData)

	r := NewRoamer(WithDecoders(decoder.NewJSON()))
	parseFunc := NewParseWithPool[LargeStruct](r)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes)))

		_ = parseFunc(req, func(data *LargeStruct) error {
			return nil
		})
	}
}

// BenchmarkNewParseWithPool_MemoryPressure benchmarks pool under memory pressure
func BenchmarkNewParseWithPool_MemoryPressure(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping memory pressure test in short mode")
	}

	type MemoryStruct struct {
		Data   string            `json:"data"`
		Values []int             `json:"values"`
		Meta   map[string]string `json:"meta"`
	}

	// Create a large payload
	largeValues := make([]int, 1000)
	for i := range largeValues {
		largeValues[i] = i
	}

	jsonData := map[string]any{
		"data":   strings.Repeat("x", 10000),
		"values": largeValues,
		"meta": map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
	}
	jsonBytes, _ := json.Marshal(jsonData)

	r := NewRoamer(WithDecoders(decoder.NewJSON()))
	parseFunc := NewParseWithPool[MemoryStruct](r)

	// Get baseline memory
	runtime.GC()
	var startMem runtime.MemStats
	runtime.ReadMemStats(&startMem)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", strconv.Itoa(len(jsonBytes)))

		_ = parseFunc(req, func(data *MemoryStruct) error {
			return nil
		})

		// Trigger GC periodically
		if i%100 == 0 {
			runtime.GC()
		}
	}

	b.StopTimer()

	// Check final memory
	runtime.GC()
	var endMem runtime.MemStats
	runtime.ReadMemStats(&endMem)

	b.ReportMetric(float64(endMem.HeapAlloc-startMem.HeapAlloc)/float64(b.N), "B/op-heap")
}
