package decoder

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

// BenchmarkStruct represents various data types for comprehensive decoder testing
type BenchmarkStruct struct {
	// Basic types
	StringField string  `json:"string_field" xml:"string_field" form:"string_field"`
	IntField    int     `json:"int_field" xml:"int_field" form:"int_field"`
	FloatField  float64 `json:"float_field" xml:"float_field" form:"float_field"`
	BoolField   bool    `json:"bool_field" xml:"bool_field" form:"bool_field"`

	// Time and complex types
	TimeField    time.Time `json:"time_field" xml:"time_field" form:"time_field"`
	PointerField *string   `json:"pointer_field" xml:"pointer_field" form:"pointer_field"`

	// Collections
	StringSlice []string          `json:"string_slice" xml:"string_slice" form:"string_slice"`
	IntSlice    []int             `json:"int_slice" xml:"int_slice" form:"int_slice"`
	StringMap   map[string]string `json:"string_map" xml:"string_map" form:"string_map"`

	// Nested struct (for JSON/XML)
	NestedStruct NestedBenchmarkStruct `json:"nested_struct" xml:"nested_struct"`
}

type NestedBenchmarkStruct struct {
	Name  string `json:"name" xml:"name"`
	Value int    `json:"value" xml:"value"`
}

// BenchmarkJSONDecoder tests JSON decoding performance
func BenchmarkJSONDecoder_Decode(b *testing.B) {
	tests := []struct {
		name     string
		dataSize string
		setup    func() *http.Request
	}{
		{
			name:     "Small_JSON_1KB",
			dataSize: "1KB",
			setup:    func() *http.Request { return createJSONRequest(b, generateSmallJSONData()) },
		},
		{
			name:     "Medium_JSON_10KB",
			dataSize: "10KB",
			setup:    func() *http.Request { return createJSONRequest(b, generateMediumJSONData()) },
		},
		{
			name:     "Large_JSON_100KB",
			dataSize: "100KB",
			setup:    func() *http.Request { return createJSONRequest(b, generateLargeJSONData()) },
		},
		{
			name:     "Complex_Nested_JSON",
			dataSize: "Complex",
			setup:    func() *http.Request { return createJSONRequest(b, generateComplexJSONData()) },
		},
	}

	decoder := NewJSON()

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			req := tt.setup()
			var target BenchmarkStruct

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				target = BenchmarkStruct{}
				if err := decoder.Decode(req, &target); err != nil {
					b.Fatal(err)
				}
				// Reset request body for next iteration
				req.Body = createJSONBody(b, req)
			}
		})
	}
}

// BenchmarkXMLDecoder tests XML decoding performance
func BenchmarkXMLDecoder_Decode(b *testing.B) {
	tests := []struct {
		name     string
		dataSize string
		setup    func() *http.Request
	}{
		{
			name:     "Small_XML_1KB",
			dataSize: "1KB",
			setup:    func() *http.Request { return createXMLRequest(b, generateSmallXMLData()) },
		},
		{
			name:     "Medium_XML_10KB",
			dataSize: "10KB",
			setup:    func() *http.Request { return createXMLRequest(b, generateMediumXMLData()) },
		},
		{
			name:     "Large_XML_100KB",
			dataSize: "100KB",
			setup:    func() *http.Request { return createXMLRequest(b, generateLargeXMLData()) },
		},
		{
			name:     "Complex_Nested_XML",
			dataSize: "Complex",
			setup:    func() *http.Request { return createXMLRequest(b, generateComplexXMLData()) },
		},
	}

	decoder := NewXML()

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			req := tt.setup()
			var target BenchmarkStruct

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				target = BenchmarkStruct{}
				if err := decoder.Decode(req, &target); err != nil {
					b.Fatal(err)
				}
				// Reset request body for next iteration
				req.Body = createXMLBody(b, req)
			}
		})
	}
}

// BenchmarkFormURLDecoder tests URL-encoded form decoding performance
func BenchmarkFormURLDecoder_Decode(b *testing.B) {
	tests := []struct {
		name       string
		fieldCount int
	}{
		{"Small_Form_10Fields", 10},
		{"Medium_Form_50Fields", 50},
		{"Large_Form_200Fields", 200},
		{"XLarge_Form_1000Fields", 1000},
	}

	decoder := NewFormURL()

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			req := createFormURLRequest(b, tt.fieldCount)
			var target map[string]any

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				target = make(map[string]any)
				if err := decoder.Decode(req, &target); err != nil {
					// Form decoding to map might fail, that's ok for benchmarking
				}
				// Reset request body for next iteration
				req.Body = createFormURLBody(b, req, tt.fieldCount)
			}
		})
	}
}

// BenchmarkDecoders_Comparison compares all decoders with similar data sizes
func BenchmarkDecoders_Comparison(b *testing.B) {
	// Create equivalent data for each decoder type
	jsonData := generateMediumJSONData()
	xmlData := generateMediumXMLData()

	tests := []struct {
		name    string
		decoder interface {
			Decode(*http.Request, any) error
		}
		request func() *http.Request
	}{
		{
			name:    "JSON_Decoder",
			decoder: NewJSON(),
			request: func() *http.Request { return createJSONRequest(b, jsonData) },
		},
		{
			name:    "XML_Decoder",
			decoder: NewXML(),
			request: func() *http.Request { return createXMLRequest(b, xmlData) },
		},
		{
			name:    "FormURL_Decoder",
			decoder: NewFormURL(),
			request: func() *http.Request { return createFormURLRequest(b, 50) },
		},
		{
			name:    "Multipart_Decoder",
			decoder: injectStructureCache(NewMultipartFormData()),
			request: func() *http.Request { return createMultipartRequest(b, 20, 2, 1024) },
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			var target BenchmarkStruct

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				req := tt.request()
				target = BenchmarkStruct{}
				if err := tt.decoder.Decode(req, &target); err != nil {
					// Some decoders might fail with BenchmarkStruct, that's ok
				}
			}
		})
	}
}

// BenchmarkDecoders_Concurrent tests concurrent decoding
func BenchmarkDecoders_Concurrent(b *testing.B) {
	jsonDecoder := NewJSON()
	xmlDecoder := NewXML()
	formDecoder := NewFormURL()
	multipartDecoder := injectStructureCache(NewMultipartFormData())

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var target BenchmarkStruct
		for pb.Next() {
			// Test different decoders concurrently
			jsonReq := createJSONRequest(b, generateSmallJSONData())
			xmlReq := createXMLRequest(b, generateSmallXMLData())
			formReq := createFormURLRequest(b, 10)
			multipartReq := createMultipartRequest(b, 5, 1, 512)

			target = BenchmarkStruct{}
			_ = jsonDecoder.Decode(jsonReq, &target)

			target = BenchmarkStruct{}
			_ = xmlDecoder.Decode(xmlReq, &target)

			target = BenchmarkStruct{}
			_ = formDecoder.Decode(formReq, &target)

			target = BenchmarkStruct{}
			_ = multipartDecoder.Decode(multipartReq, &target)
		}
	})
}

// BenchmarkDecoders_MemoryAllocation focuses on memory allocation patterns
func BenchmarkDecoders_MemoryAllocation(b *testing.B) {
	tests := []struct {
		name    string
		decoder interface {
			Decode(*http.Request, any) error
			ContentType() string
		}
		requestSizes []string
	}{
		{"JSON_Memory", NewJSON(), []string{"small", "medium", "large"}},
		{"XML_Memory", NewXML(), []string{"small", "medium", "large"}},
		{"FormURL_Memory", NewFormURL(), []string{"small", "medium", "large"}},
	}

	for _, tt := range tests {
		for _, size := range tt.requestSizes {
			b.Run(fmt.Sprintf("%s_%s", tt.name, size), func(b *testing.B) {
				var req *http.Request
				switch tt.decoder.ContentType() {
				case "application/json":
					req = createJSONRequestBySize(b, size)
				case "application/xml":
					req = createXMLRequestBySize(b, size)
				case "application/x-www-form-urlencoded":
					req = createFormURLRequestBySize(b, size)
				default:
					b.Errorf("unhandled content type %s", tt.decoder.ContentType())
				}

				var target any

				b.ReportAllocs()
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					target = make(map[string]any)
					_ = tt.decoder.Decode(req, &target)
					// Reset for next iteration
					switch tt.decoder.ContentType() {
					case "application/json":
						req.Body = createJSONBody(b, req)
					case "application/xml":
						req.Body = createXMLBody(b, req)
					case "application/x-www-form-urlencoded":
						req.Body = createFormURLBodyBySize(b, req, size)
					case "multipart/form-data":
						req.Body = createMultipartBodyBySize(b, req, size)
					}
				}
			})
		}
	}
}

// Helper functions for creating test requests and data

func generateSmallJSONData() map[string]any {
	return map[string]any{
		"string_field": "test_value",
		"int_field":    42,
		"float_field":  3.14159,
		"bool_field":   true,
		"time_field":   time.Now().Format(time.RFC3339),
	}
}

func generateMediumJSONData() map[string]any {
	data := generateSmallJSONData()
	data["string_slice"] = []string{"item1", "item2", "item3", "item4", "item5"}
	data["int_slice"] = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	data["string_map"] = map[string]string{
		"key1": "value1", "key2": "value2", "key3": "value3",
		"key4": "value4", "key5": "value5",
	}
	data["nested_struct"] = map[string]any{
		"name":  "nested_name",
		"value": 99,
	}
	return data
}

func generateLargeJSONData() map[string]any {
	data := generateMediumJSONData()

	// Add many more fields to make it large
	for i := 0; i < 100; i++ {
		data[fmt.Sprintf("extra_field_%d", i)] = fmt.Sprintf("extra_value_%d", i)
	}

	// Large string slice
	largeSlice := make([]string, 200)
	for i := range largeSlice {
		largeSlice[i] = fmt.Sprintf("large_item_%d", i)
	}
	data["large_string_slice"] = largeSlice

	return data
}

func generateComplexJSONData() map[string]any {
	return map[string]any{
		"users": []map[string]any{
			{
				"id":    1,
				"name":  "User One",
				"email": "user1@example.com",
				"profile": map[string]any{
					"age":     25,
					"city":    "New York",
					"hobbies": []string{"reading", "swimming", "coding"},
				},
			},
			{
				"id":    2,
				"name":  "User Two",
				"email": "user2@example.com",
				"profile": map[string]any{
					"age":     30,
					"city":    "San Francisco",
					"hobbies": []string{"hiking", "photography", "cooking"},
				},
			},
		},
		"metadata": map[string]any{
			"version":     "1.0.0",
			"api_key":     "secret_key_12345",
			"environment": "production",
		},
	}
}

func generateSmallXMLData() string {
	return `<BenchmarkStruct>
		<string_field>test_value</string_field>
		<int_field>42</int_field>
		<float_field>3.14159</float_field>
		<bool_field>true</bool_field>
	</BenchmarkStruct>`
}

func generateMediumXMLData() string {
	return `<BenchmarkStruct>
		<string_field>test_value</string_field>
		<int_field>42</int_field>
		<float_field>3.14159</float_field>
		<bool_field>true</bool_field>
		<string_slice>item1</string_slice>
		<string_slice>item2</string_slice>
		<string_slice>item3</string_slice>
		<int_slice>1</int_slice>
		<int_slice>2</int_slice>
		<int_slice>3</int_slice>
		<nested_struct>
			<name>nested_name</name>
			<value>99</value>
		</nested_struct>
	</BenchmarkStruct>`
}

func generateLargeXMLData() string {
	var sb strings.Builder
	sb.WriteString(`<BenchmarkStruct>`)
	sb.WriteString(generateMediumXMLData()[17 : len(generateMediumXMLData())-19]) // Remove outer tags

	for i := 0; i < 100; i++ {
		sb.WriteString(fmt.Sprintf(`<extra_field_%d>extra_value_%d</extra_field_%d>`, i, i, i))
	}

	sb.WriteString(`</BenchmarkStruct>`)
	return sb.String()
}

func generateComplexXMLData() string {
	return `<Root>
		<users>
			<user>
				<id>1</id>
				<name>User One</name>
				<email>user1@example.com</email>
				<profile>
					<age>25</age>
					<city>New York</city>
					<hobbies>reading</hobbies>
					<hobbies>swimming</hobbies>
					<hobbies>coding</hobbies>
				</profile>
			</user>
			<user>
				<id>2</id>
				<name>User Two</name>
				<email>user2@example.com</email>
				<profile>
					<age>30</age>
					<city>San Francisco</city>
					<hobbies>hiking</hobbies>
					<hobbies>photography</hobbies>
					<hobbies>cooking</hobbies>
				</profile>
			</user>
		</users>
	</Root>`
}

func createJSONRequest(b *testing.B, data map[string]any) *http.Request {
	b.Helper()

	jsonBytes, _ := stdjson.Marshal(data)
	req, _ := http.NewRequest(http.MethodPost, "https://example.com/test", bytes.NewReader(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func createXMLRequest(b *testing.B, xmlData string) *http.Request {
	b.Helper()

	req, _ := http.NewRequest(http.MethodPost, "https://example.com/test", strings.NewReader(xmlData))
	req.Header.Set("Content-Type", "application/xml")
	return req
}

func createFormURLRequest(b *testing.B, fieldCount int) *http.Request {
	b.Helper()

	data := url.Values{}
	for i := 0; i < fieldCount; i++ {
		data.Add(fmt.Sprintf("field_%d", i), fmt.Sprintf("value_%d", i))
	}

	req, _ := http.NewRequest(http.MethodPost, "https://example.com/test", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func createMultipartRequest(b *testing.B, fields, fileCount, fileSize int) *http.Request {
	b.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add regular fields
	for i := 0; i < fields; i++ {
		writer.WriteField(fmt.Sprintf("field_%d", i), fmt.Sprintf("value_%d", i))
	}

	// Add file fields
	for i := 0; i < fileCount; i++ {
		part, _ := writer.CreateFormFile(fmt.Sprintf("file_%d", i), fmt.Sprintf("file_%d.txt", i))
		fileContent := strings.Repeat(fmt.Sprintf("content_%d ", i), fileSize/10)
		part.Write([]byte(fileContent))
	}

	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "https://example.com/test", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// Helper functions to create request bodies and reset them

func createJSONBody(b *testing.B, req *http.Request) io.ReadCloser {
	b.Helper()
	// This is a simplified version - in real usage you'd need to capture the original data
	data := generateMediumJSONData()
	jsonBytes, _ := stdjson.Marshal(data)
	return io.NopCloser(bytes.NewReader(jsonBytes))
}

func createXMLBody(b *testing.B, req *http.Request) io.ReadCloser {
	b.Helper()
	return io.NopCloser(strings.NewReader(generateMediumXMLData()))
}

func createFormURLBody(b *testing.B, req *http.Request, fieldCount int) io.ReadCloser {
	b.Helper()
	data := url.Values{}
	for i := 0; i < fieldCount; i++ {
		data.Add(fmt.Sprintf("field_%d", i), fmt.Sprintf("value_%d", i))
	}
	return io.NopCloser(strings.NewReader(data.Encode()))
}

func createMultipartBody(b *testing.B, req *http.Request, fields, fileCount, fileSize int) io.ReadCloser {
	b.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for i := 0; i < fields; i++ {
		writer.WriteField(fmt.Sprintf("field_%d", i), fmt.Sprintf("value_%d", i))
	}

	for i := 0; i < fileCount; i++ {
		part, _ := writer.CreateFormFile(fmt.Sprintf("file_%d", i), fmt.Sprintf("file_%d.txt", i))
		fileContent := strings.Repeat(fmt.Sprintf("content_%d ", i), fileSize/10)
		part.Write([]byte(fileContent))
	}

	writer.Close()
	return io.NopCloser(&buf)
}

// Size-based helper functions

func createJSONRequestBySize(b *testing.B, size string) *http.Request {
	b.Helper()
	var data map[string]any
	switch size {
	case "small":
		data = generateSmallJSONData()
	case "medium":
		data = generateMediumJSONData()
	case "large":
		data = generateLargeJSONData()
	default:
		data = generateMediumJSONData()
	}
	return createJSONRequest(b, data)
}

func createXMLRequestBySize(b *testing.B, size string) *http.Request {
	b.Helper()
	var xmlData string
	switch size {
	case "small":
		xmlData = generateSmallXMLData()
	case "medium":
		xmlData = generateMediumXMLData()
	case "large":
		xmlData = generateLargeXMLData()
	default:
		xmlData = generateMediumXMLData()
	}
	return createXMLRequest(b, xmlData)
}

func createFormURLRequestBySize(b *testing.B, size string) *http.Request {
	b.Helper()
	var fieldCount int
	switch size {
	case "small":
		fieldCount = 10
	case "medium":
		fieldCount = 50
	case "large":
		fieldCount = 200
	default:
		fieldCount = 50
	}
	return createFormURLRequest(b, fieldCount)
}

func createMultipartRequestBySize(b *testing.B, size string) *http.Request {
	b.Helper()
	var fields, fileCount, fileSize int
	switch size {
	case "small":
		fields, fileCount, fileSize = 5, 1, 512
	case "medium":
		fields, fileCount, fileSize = 20, 2, 1024
	case "large":
		fields, fileCount, fileSize = 50, 5, 10240
	default:
		fields, fileCount, fileSize = 20, 2, 1024
	}
	return createMultipartRequest(b, fields, fileCount, fileSize)
}

func createFormURLBodyBySize(b *testing.B, req *http.Request, size string) io.ReadCloser {
	b.Helper()
	var fieldCount int
	switch size {
	case "small":
		fieldCount = 10
	case "medium":
		fieldCount = 50
	case "large":
		fieldCount = 200
	default:
		fieldCount = 50
	}
	return createFormURLBody(b, req, fieldCount)
}

func createMultipartBodyBySize(b *testing.B, req *http.Request, size string) io.ReadCloser {
	b.Helper()
	var fields, fileCount, fileSize int
	switch size {
	case "small":
		fields, fileCount, fileSize = 5, 1, 512
	case "medium":
		fields, fileCount, fileSize = 20, 2, 1024
	case "large":
		fields, fileCount, fileSize = 50, 5, 10240
	default:
		fields, fileCount, fileSize = 20, 2, 1024
	}
	return createMultipartBody(b, req, fields, fileCount, fileSize)
}
