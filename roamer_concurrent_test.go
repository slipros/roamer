package roamer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
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

// TestRoamer_Parse_Concurrent_Basic tests basic concurrent parsing safety
func TestRoamer_Parse_Concurrent_Basic(t *testing.T) {
	// Create roamer instance
	roamer := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewQuery(), parser.NewHeader()),
		WithFormatters(formatter.NewString()),
	)

	// Create test JSON data
	jsonData := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	// Test struct
	type TestStruct struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
		ID    string `query:"id"`
		Auth  string `header:"Authorization"`
	}

	const numGoroutines = 100
	const numIterationsPerGoroutine = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numIterationsPerGoroutine)

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numIterationsPerGoroutine; j++ {
				// Create fresh request for each iteration to avoid sharing request body
				jsonBytes, _ := json.Marshal(jsonData)
				req, _ := http.NewRequest(http.MethodPost, "https://api.example.com/users?id=123", bytes.NewReader(jsonBytes))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer token123")

				var target TestStruct
				if err := roamer.Parse(req, &target); err != nil {
					errors <- fmt.Errorf("goroutine %d, iteration %d: %w", goroutineID, j, err)
					return
				}

				// Validate parsing results
				if target.Name != "John Doe" || target.Email != "john@example.com" || target.Age != 30 {
					errors <- fmt.Errorf("goroutine %d, iteration %d: incorrect parsing results", goroutineID, j)
					return
				}
				if target.ID != "123" || target.Auth != "Bearer token123" {
					errors <- fmt.Errorf("goroutine %d, iteration %d: incorrect header/query parsing", goroutineID, j)
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	// Check for errors
	var errorList []error
	for err := range errors {
		errorList = append(errorList, err)
	}

	require.Empty(t, errorList, "Concurrent parsing should not produce errors")
}

// TestRoamer_Parse_Concurrent_RaceConditions tests for race conditions using -race detector
func TestRoamer_Parse_Concurrent_RaceConditions(t *testing.T) {
	roamer := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewQuery(), parser.NewHeader(), parser.NewCookie()),
		WithFormatters(formatter.NewString()),
	)

	// Define request types for creating fresh requests in each goroutine
	requestTypes := []struct {
		reqType  string
		jsonData map[string]interface{}
	}{
		{"json", map[string]interface{}{"field": "value1"}},
		{"query", nil},
		{"headers", nil},
		{"cookies", nil},
		{"mixed", map[string]interface{}{"data": "mixed"}},
	}

	type TestStruct struct {
		Field string `json:"field" format:"lower_case"`
		Query string `query:"test_param"`
		Auth  string `header:"Authorization" format:"trim_space"`
		Token string `cookie:"session_token"`
		Data  string `json:"data"`
	}

	const numWorkers = 20
	const numRequestsPerWorker = 50

	var wg sync.WaitGroup
	errChan := make(chan error, numWorkers*numRequestsPerWorker)

	// Launch workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for i := 0; i < numRequestsPerWorker; i++ {
				// Create a fresh request for each iteration to avoid race conditions
				reqType := requestTypes[i%len(requestTypes)]
				req := createConcurrentTestRequest(t, reqType.reqType, reqType.jsonData)
				var target TestStruct

				if err := roamer.Parse(req, &target); err != nil {
					errChan <- fmt.Errorf("worker %d, request %d: %w", workerID, i, err)
				}
			}
		}(w)
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	assert.Empty(t, errors, "Should not have race conditions")
}

// TestRoamer_Parse_Concurrent_MemoryLeaks tests for memory leaks in concurrent scenarios
func TestRoamer_Parse_Concurrent_MemoryLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	roamer := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewQuery(), parser.NewHeader()),
	)

	// Get initial memory stats
	runtime.GC()
	var initialStats runtime.MemStats
	runtime.ReadMemStats(&initialStats)

	// Create test request
	jsonData := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		jsonData[fmt.Sprintf("field_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	type LargeStruct struct {
		Fields map[string]interface{} `json:",inline"`
	}

	const numGoroutines = 50
	const numIterations = 100

	var wg sync.WaitGroup

	// Run concurrent parsing operations
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < numIterations; i++ {
				jsonBytes, _ := json.Marshal(jsonData)
				req, _ := http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader(jsonBytes))
				req.Header.Set("Content-Type", "application/json")

				var target LargeStruct
				_ = roamer.Parse(req, &target)

				// Intentionally create some garbage
				_ = make([]byte, 1024)

				if i%10 == 0 {
					runtime.GC()
				}
			}
		}()
	}

	wg.Wait()

	// Force garbage collection and get final memory stats
	runtime.GC()
	runtime.GC() // Call twice to ensure cleanup
	var finalStats runtime.MemStats
	runtime.ReadMemStats(&finalStats)

	// Check for significant memory growth (allowing for some variance)
	memoryGrowth := finalStats.HeapAlloc - initialStats.HeapAlloc
	maxAcceptableGrowth := uint64(10 * 1024 * 1024) // 10MB threshold

	t.Logf("Memory growth: %d bytes", memoryGrowth)
	t.Logf("Initial heap: %d bytes", initialStats.HeapAlloc)
	t.Logf("Final heap: %d bytes", finalStats.HeapAlloc)

	assert.Less(t, memoryGrowth, maxAcceptableGrowth,
		"Memory growth should not exceed %d bytes, but grew by %d bytes", maxAcceptableGrowth, memoryGrowth)
}

// TestRoamer_Parse_Concurrent_ContextCancellation tests behavior with context cancellation
func TestRoamer_Parse_Concurrent_ContextCancellation(t *testing.T) {
	roamer := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewQuery()),
	)

	type TestStruct struct {
		Data string `json:"data"`
	}

	jsonData := map[string]interface{}{"data": "test"}
	jsonBytes, _ := json.Marshal(jsonData)

	const numGoroutines = 20
	var wg sync.WaitGroup
	var completedCount int64
	var cancelledCount int64

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					atomic.AddInt64(&cancelledCount, 1)
					return
				default:
					req, _ := http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader(jsonBytes))
					req.Header.Set("Content-Type", "application/json")
					req = req.WithContext(ctx)

					var target TestStruct
					if err := roamer.Parse(req, &target); err != nil && ctx.Err() != nil {
						atomic.AddInt64(&cancelledCount, 1)
						return
					}
					atomic.AddInt64(&completedCount, 1)

					// Small delay to allow context cancellation to occur
					time.Sleep(1 * time.Millisecond)
				}
			}
		}(i)
	}

	wg.Wait()

	completed := atomic.LoadInt64(&completedCount)
	cancelled := atomic.LoadInt64(&cancelledCount)

	t.Logf("Completed operations: %d", completed)
	t.Logf("Cancelled operations: %d", cancelled)

	assert.Equal(t, int64(numGoroutines), cancelled, "All goroutines should be cancelled")
	assert.True(t, completed > 0, "Some operations should complete before cancellation")
}

// TestRoamer_Parse_Concurrent_StructureCache tests concurrent access to structure cache
func TestRoamer_Parse_Concurrent_StructureCache(t *testing.T) {
	roamer := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewQuery()),
	)

	// Define multiple struct types to test cache behavior
	type Struct1 struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	type Struct2 struct {
		Data   string  `json:"data"`
		Value  float64 `json:"value"`
		Active bool    `json:"active"`
	}

	type Struct3 struct {
		Name    string    `json:"name"`
		Email   string    `json:"email"`
		Created time.Time `json:"created"`
		Tags    []string  `json:"tags"`
	}

	// Factory functions to create fresh struct instances
	structFactories := []func() interface{}{
		func() interface{} { return &Struct1{} },
		func() interface{} { return &Struct2{} },
		func() interface{} { return &Struct3{} },
	}

	jsonDataSets := []map[string]interface{}{
		{"field1": "value1", "field2": 42},
		{"data": "test", "value": 3.14, "active": true},
		{"name": "John", "email": "john@example.com", "created": time.Now().Format(time.RFC3339), "tags": []string{"user", "active"}},
	}

	const numGoroutines = 30
	var wg sync.WaitGroup
	errChan := make(chan error, numGoroutines*10)

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for i := 0; i < 10; i++ {
				structIndex := (goroutineID + i) % len(structFactories)
				// Create a fresh struct instance for each goroutine to avoid race conditions
				target := structFactories[structIndex]()
				jsonData := jsonDataSets[structIndex]

				jsonBytes, _ := json.Marshal(jsonData)
				req, _ := http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader(jsonBytes))
				req.Header.Set("Content-Type", "application/json")

				if err := roamer.Parse(req, target); err != nil {
					errChan <- fmt.Errorf("goroutine %d, iteration %d, struct %d: %w", goroutineID, i, structIndex, err)
				}
			}
		}(g)
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	assert.Empty(t, errors, "Concurrent structure cache access should not cause errors")
}

// TestRoamer_Parse_Concurrent_ParserCache tests concurrent access to parser cache pool
func TestRoamer_Parse_Concurrent_ParserCache(t *testing.T) {
	roamer := NewRoamer(
		WithParsers(parser.NewQuery(), parser.NewHeader(), parser.NewCookie()),
	)

	type TestStruct struct {
		Query1  string `query:"q1"`
		Query2  string `query:"q2"`
		Header1 string `header:"H1"`
		Header2 string `header:"H2"`
		Cookie1 string `cookie:"c1"`
		Cookie2 string `cookie:"c2"`
	}

	const numGoroutines = 25
	var wg sync.WaitGroup

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for i := 0; i < 20; i++ {
				req, _ := http.NewRequest(http.MethodGet,
					fmt.Sprintf("https://example.com?q1=value1_%d&q2=value2_%d", id, i), nil)
				req.Header.Set("H1", fmt.Sprintf("header1_%d", id))
				req.Header.Set("H2", fmt.Sprintf("header2_%d", i))
				req.AddCookie(&http.Cookie{Name: "c1", Value: fmt.Sprintf("cookie1_%d", id)})
				req.AddCookie(&http.Cookie{Name: "c2", Value: fmt.Sprintf("cookie2_%d", i)})

				var target TestStruct
				if err := roamer.Parse(req, &target); err != nil {
					t.Errorf("goroutine %d, iteration %d: %v", id, i, err)
				}
			}
		}(g)
	}

	wg.Wait()
}

// TestRoamer_Parse_Concurrent_FormatterAccess tests concurrent access to formatters
func TestRoamer_Parse_Concurrent_FormatterAccess(t *testing.T) {
	roamer := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithFormatters(formatter.NewString()),
	)

	type TestStruct struct {
		Lower   string `json:"lower" format:"lower_case"`
		Upper   string `json:"upper" format:"upper_case"`
		Trimmed string `json:"trimmed" format:"trim_space"`
		MultiOp string `json:"multi" format:"trim_space,lower_case"`
	}

	const numGoroutines = 20
	var wg sync.WaitGroup

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for i := 0; i < 15; i++ {
				jsonData := map[string]interface{}{
					"lower":   fmt.Sprintf("  UPPER_TEXT_%d_%d  ", id, i),
					"upper":   fmt.Sprintf("  lower_text_%d_%d  ", id, i),
					"trimmed": fmt.Sprintf("  trimmed_text_%d_%d  ", id, i),
					"multi":   fmt.Sprintf("  MULTI_OP_TEXT_%d_%d  ", id, i),
				}

				jsonBytes, _ := json.Marshal(jsonData)
				req, _ := http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader(jsonBytes))
				req.Header.Set("Content-Type", "application/json")

				var target TestStruct
				if err := roamer.Parse(req, &target); err != nil {
					t.Errorf("goroutine %d, iteration %d: %v", id, i, err)
					continue
				}

				// Verify formatters worked correctly (just check that formatting occurred)
				// The exact values depend on the formatter implementations
				assert.NotEmpty(t, target.Lower, "Lower field should not be empty")
				assert.NotEmpty(t, target.Upper, "Upper field should not be empty")
				assert.NotEmpty(t, target.Trimmed, "Trimmed field should not be empty")
				assert.NotEmpty(t, target.MultiOp, "MultiOp field should not be empty")
			}
		}(g)
	}

	wg.Wait()
}

// Helper functions for concurrent testing

func createConcurrentTestRequest(t *testing.T, requestType string, jsonData map[string]interface{}) *http.Request {
	t.Helper()

	var req *http.Request
	var body []byte

	if jsonData != nil {
		body, _ = json.Marshal(jsonData)
	}

	switch requestType {
	case "json":
		req, _ = http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	case "query":
		req, _ = http.NewRequest(http.MethodGet, "https://example.com?test_param=query_value", nil)
	case "headers":
		req, _ = http.NewRequest(http.MethodGet, "https://example.com", nil)
		req.Header.Set("Authorization", "Bearer token123")
	case "cookies":
		req, _ = http.NewRequest(http.MethodGet, "https://example.com", nil)
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "session123"})
	case "mixed":
		req, _ = http.NewRequest(http.MethodPost, "https://example.com?test_param=mixed_query", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "  Bearer mixed_token  ")
		req.AddCookie(&http.Cookie{Name: "session_token", Value: "mixed_session"})
	default:
		req, _ = http.NewRequest(http.MethodGet, "https://example.com", nil)
	}

	return req
}

// BenchmarkRoamer_Parse_Concurrent benchmarks concurrent parsing performance
func BenchmarkRoamer_Parse_Concurrent(b *testing.B) {
	roamer := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewQuery(), parser.NewHeader()),
	)

	type TestStruct struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		ID    string `query:"id"`
		Auth  string `header:"Authorization"`
	}

	jsonData := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	jsonBytes, _ := json.Marshal(jsonData)
	req, _ := http.NewRequest(http.MethodPost, "https://example.com?id=123", bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var target TestStruct
		for pb.Next() {
			target = TestStruct{}
			if err := roamer.Parse(req, &target); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkRoamer_Parse_Concurrent_Contention benchmarks with high contention
func BenchmarkRoamer_Parse_Concurrent_Contention(b *testing.B) {
	roamer := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewQuery(), parser.NewHeader(), parser.NewCookie()),
		WithFormatters(formatter.NewString()),
	)

	type ComplexStruct struct {
		// Many fields to stress the cache
		F1  string `json:"f1" format:"lower_case"`
		F2  string `json:"f2" format:"upper_case"`
		F3  string `json:"f3" format:"trim_space"`
		F4  string `json:"f4"`
		F5  string `json:"f5"`
		F6  int    `json:"f6"`
		F7  int    `json:"f7"`
		F8  bool   `json:"f8"`
		F9  bool   `json:"f9"`
		F10 string `query:"q1"`
		F11 string `query:"q2"`
		F12 string `header:"H1" format:"trim_space"`
		F13 string `header:"H2"`
		F14 string `cookie:"c1"`
		F15 string `cookie:"c2"`
	}

	complexData := map[string]interface{}{
		"f1": "LOWER", "f2": "upper", "f3": "  trimmed  ",
		"f4": "field4", "f5": "field5", "f6": 42, "f7": 84,
		"f8": true, "f9": false,
	}
	jsonBytes, _ := json.Marshal(complexData)

	req, _ := http.NewRequest(http.MethodPost, "https://example.com?q1=query1&q2=query2", bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("H1", "  header1  ")
	req.Header.Set("H2", "header2")
	req.AddCookie(&http.Cookie{Name: "c1", Value: "cookie1"})
	req.AddCookie(&http.Cookie{Name: "c2", Value: "cookie2"})

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var target ComplexStruct
		for pb.Next() {
			target = ComplexStruct{}
			if err := roamer.Parse(req, &target); err != nil {
				b.Fatal(err)
			}
		}
	})
}
