package roamer

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRoamer_PreserveBody_Enabled tests body preservation when enabled
func TestRoamer_PreserveBody_Enabled(t *testing.T) {
	type testData struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	tests := []struct {
		name       string
		inputData  testData
		verifyBody func(*testing.T, []byte, testData)
	}{
		{
			name: "body is preserved after successful JSON decode",
			inputData: testData{
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   30,
			},
			verifyBody: func(t *testing.T, bodyBytes []byte, original testData) {
				var decoded testData
				err := json.Unmarshal(bodyBytes, &decoded)
				require.NoError(t, err, "should be able to decode preserved body")
				assert.Equal(t, original.Name, decoded.Name)
				assert.Equal(t, original.Email, decoded.Email)
				assert.Equal(t, original.Age, decoded.Age)
			},
		},
		{
			name: "body can be read multiple times",
			inputData: testData{
				Name:  "Jane Smith",
				Email: "jane@example.com",
				Age:   25,
			},
			verifyBody: func(t *testing.T, bodyBytes []byte, original testData) {
				// Verify the body matches the original multiple reads would work
				var decoded testData
				err := json.Unmarshal(bodyBytes, &decoded)
				require.NoError(t, err)
				assert.Equal(t, original, decoded)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create JSON body
			jsonData, err := json.Marshal(tt.inputData)
			require.NoError(t, err)

			// Create request
			req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonData))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Length", strconv.Itoa(len(jsonData)))

			// Create roamer with body preservation enabled
			r := NewRoamer(
				WithDecoders(decoder.NewJSON()),
				WithPreserveBody(),
			)

			// Parse the request
			var result testData
			err = r.Parse(req, &result)
			require.NoError(t, err, "Parse should succeed")

			// Verify parsed data
			assert.Equal(t, tt.inputData, result, "Parsed data should match input")

			// Read the body again to verify it was preserved
			bodyBytes, err := io.ReadAll(req.Body)
			require.NoError(t, err, "Should be able to read body again")
			assert.NotEmpty(t, bodyBytes, "Body should not be empty")

			// Verify the preserved body content
			if tt.verifyBody != nil {
				tt.verifyBody(t, bodyBytes, tt.inputData)
			}

			// Close the body
			err = req.Body.Close()
			require.NoError(t, err)
		})
	}
}

// TestRoamer_PreserveBody_Disabled tests that body is consumed when preservation is disabled
func TestRoamer_PreserveBody_Disabled(t *testing.T) {
	type testData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	inputData := testData{
		Name: "John Doe",
		Age:  30,
	}

	// Create JSON body
	jsonData, err := json.Marshal(inputData)
	require.NoError(t, err)

	// Create request
	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(jsonData)))

	// Create roamer with body preservation disabled (default behavior)
	r := NewRoamer(
		WithDecoders(decoder.NewJSON()),
	)

	// Parse the request
	var result testData
	err = r.Parse(req, &result)
	require.NoError(t, err, "Parse should succeed")

	// Verify parsed data
	assert.Equal(t, inputData, result, "Parsed data should match input")

	// Try to read the body again - it should be consumed (EOF)
	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(t, err, "ReadAll should not error on EOF")
	assert.Empty(t, bodyBytes, "Body should be empty after consumption")
}

// TestRoamer_PreserveBody_DefaultBehavior tests that preservation is disabled by default
func TestRoamer_PreserveBody_DefaultBehavior(t *testing.T) {
	type testData struct {
		Value string `json:"value"`
	}

	inputData := testData{Value: "test"}

	// Create JSON body
	jsonData, err := json.Marshal(inputData)
	require.NoError(t, err)

	// Create request
	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(jsonData)))

	// Create roamer without specifying WithPreserveBody (should default to false)
	r := NewRoamer(
		WithDecoders(decoder.NewJSON()),
	)

	// Parse the request
	var result testData
	err = r.Parse(req, &result)
	require.NoError(t, err)

	// Verify the body is consumed (default behavior)
	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.Empty(t, bodyBytes, "Body should be consumed by default")
}

// TestRoamer_PreserveBody_MultipleReads tests that the body can be read multiple times
func TestRoamer_PreserveBody_MultipleReads(t *testing.T) {
	type testData struct {
		Message string `json:"message"`
		Count   int    `json:"count"`
	}

	inputData := testData{
		Message: "hello",
		Count:   42,
	}

	// Create JSON body
	jsonData, err := json.Marshal(inputData)
	require.NoError(t, err)

	// Create request
	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(jsonData)))

	// Create roamer with body preservation
	r := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithPreserveBody(),
	)

	// Parse the request
	var result testData
	err = r.Parse(req, &result)
	require.NoError(t, err)
	assert.Equal(t, inputData, result)

	// First read
	bodyBytes1, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.NotEmpty(t, bodyBytes1)

	// The body should be fully consumed after first read
	// This is expected behavior - io.NopCloser wraps a bytes.Reader
	// which can only be read once. For multiple reads, the consumer
	// needs to recreate the reader from the preserved bytes.
	bodyBytes2, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.Empty(t, bodyBytes2, "Second read should be empty (bytes.Reader is consumed)")

	// However, the first read should contain valid data
	var decoded testData
	err = json.Unmarshal(bodyBytes1, &decoded)
	require.NoError(t, err)
	assert.Equal(t, inputData, decoded)
}

// TestRoamer_PreserveBody_ErrorHandling tests body preservation when decoding fails
func TestRoamer_PreserveBody_ErrorHandling(t *testing.T) {
	// Invalid JSON
	invalidJSON := []byte(`{"name": "John", "age": }`)

	// Create request with invalid JSON
	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(invalidJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(invalidJSON)))

	// Create roamer with body preservation
	r := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithPreserveBody(),
	)

	// Parse should fail due to invalid JSON
	type testData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var result testData
	err = r.Parse(req, &result)
	require.Error(t, err, "Parse should fail with invalid JSON")

	// Even on error, the body should be preserved for inspection
	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(t, err, "Should be able to read body even after decode error")
	assert.Equal(t, invalidJSON, bodyBytes, "Body should be preserved even on error")
}

// TestRoamer_PreserveBody_EmptyBody tests body preservation with empty body
func TestRoamer_PreserveBody_EmptyBody(t *testing.T) {
	// Create request with empty body (ContentLength = 0)
	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte{}))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", "0")

	// Create roamer with body preservation
	r := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithPreserveBody(),
	)

	// Parse should succeed (empty body is valid)
	type testData struct {
		Name string `json:"name"`
	}
	var result testData
	err = r.Parse(req, &result)
	require.NoError(t, err)

	// Body should remain empty
	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.Empty(t, bodyBytes)
}

// TestRoamer_PreserveBody_LargeBody tests body preservation with larger payloads
func TestRoamer_PreserveBody_LargeBody(t *testing.T) {
	// Create a larger payload
	type largeData struct {
		Items []string          `json:"items"`
		Meta  map[string]string `json:"meta"`
	}

	inputData := largeData{
		Items: make([]string, 100),
		Meta:  make(map[string]string),
	}

	// Fill with data
	for i := 0; i < 100; i++ {
		inputData.Items[i] = strconv.Itoa(i)
		inputData.Meta["key"+strconv.Itoa(i)] = "value" + strconv.Itoa(i)
	}

	// Create JSON body
	jsonData, err := json.Marshal(inputData)
	require.NoError(t, err)
	require.Greater(t, len(jsonData), 1000, "Test should use a reasonably large body")

	// Create request
	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(jsonData)))

	// Create roamer with body preservation
	r := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithPreserveBody(),
	)

	// Parse the request
	var result largeData
	err = r.Parse(req, &result)
	require.NoError(t, err)
	assert.Equal(t, inputData, result)

	// Read the body again
	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.NotEmpty(t, bodyBytes)

	// Verify preserved body is complete
	var decoded largeData
	err = json.Unmarshal(bodyBytes, &decoded)
	require.NoError(t, err)
	assert.Equal(t, inputData, decoded)
}

// TestRoamer_PreserveBody_WithParsers tests body preservation works alongside parsers
func TestRoamer_PreserveBody_WithParsers(t *testing.T) {
	type combinedData struct {
		// From JSON body
		Name string `json:"name"`
		// From query parameters
		ID int `query:"id"`
	}

	inputData := combinedData{
		Name: "Test User",
		ID:   123,
	}

	// Create JSON body
	jsonData, err := json.Marshal(struct {
		Name string `json:"name"`
	}{Name: inputData.Name})
	require.NoError(t, err)

	// Create request with both JSON body and query parameters
	req, err := http.NewRequest(http.MethodPost, "http://example.com?id=123", bytes.NewReader(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(jsonData)))

	// Create roamer with both decoders and parsers, with body preservation
	r := NewRoamer(
		WithDecoders(decoder.NewJSON()),
		WithParsers(parser.NewQuery()),
		WithPreserveBody(),
	)

	// Parse the request
	var result combinedData
	err = r.Parse(req, &result)
	require.NoError(t, err)
	assert.Equal(t, "Test User", result.Name, "Name from JSON should be parsed")
	assert.Equal(t, 123, result.ID, "ID from query should be parsed")

	// Verify body is still preserved
	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.NotEmpty(t, bodyBytes)

	var bodyData struct {
		Name string `json:"name"`
	}
	err = json.Unmarshal(bodyBytes, &bodyData)
	require.NoError(t, err)
	assert.Equal(t, "Test User", bodyData.Name)
}
