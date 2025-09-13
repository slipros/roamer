package decoder

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type multipartFormDataTestData struct {
	Image      string  `multipart:"image_url"`
	ImageAsPtr *string `multipart:"image_url"`
	ImageAsURL url.URL `multipart:"image_url"`
	NoKey      string  `multipart:"no_key"`
	NoKeyAsPtr *string `multipart:"no_key"`

	TextFile      MultipartFile  `multipart:"text_file"`
	OtherTextFile *MultipartFile `multipart:"other_text_file"`
	NoFile        MultipartFile  `multipart:"no_file"`
	NoFileAsPtr   *MultipartFile `multipart:"no_file"`
	AllFiles      MultipartFiles `multipart:",allfiles"`
}

func TestNewMultipartFormData(t *testing.T) {
	m := injectStructureCache(NewMultipartFormData())
	require.NotNil(t, m)
	require.Equal(t, ContentTypeMultipartFormData, m.ContentType())

	m = injectStructureCache(NewMultipartFormData(WithMaxMemory(1000)))
	require.NotNil(t, m)
	require.Equal(t, ContentTypeMultipartFormData, m.ContentType())
	require.Equal(t, int64(1000), m.maxMemory)

	m = injectStructureCache(NewMultipartFormData(WithContentType[*MultipartFormData]("test")))
	require.NotNil(t, m)
	require.Equal(t, "test", m.ContentType())

	m = injectStructureCache(NewMultipartFormData(WithSkipFilled[*MultipartFormData](false)))
	require.NotNil(t, m)
	require.Equal(t, false, m.skipFilled)
}

func TestMultipartFormData_Decode(t *testing.T) {
	type args struct {
		r    *http.Request
		ptr  *multipartFormDataTestData
		want *multipartFormDataTestData
	}
	tests := []struct {
		name    string
		args    func() args
		wantErr bool
	}{
		{
			name: "success",
			args: func() args {
				r, ptr, want := prepareMultipartFormDataArgs()

				return args{
					r:    r,
					ptr:  ptr,
					want: want,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args()

			m := injectStructureCache(NewMultipartFormData())
			if err := m.Decode(args.r, args.ptr); !tt.wantErr && err != nil {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.Equal(t, "http://some.url", args.ptr.Image)
			require.NotNil(t, args.ptr.ImageAsPtr)
			require.Equal(t, "http://some.url", *args.ptr.ImageAsPtr)
			require.Equal(t, "http://some.url", args.ptr.ImageAsURL.String())

			require.Empty(t, args.ptr.NoKey)
			require.Nil(t, args.ptr.NoKeyAsPtr)

			require.Equal(t, "text_file", args.ptr.TextFile.Key)
			require.NotNil(t, args.ptr.TextFile.File)
			require.NotNil(t, args.ptr.TextFile.Header)

			require.NotNil(t, args.ptr.OtherTextFile)
			require.Equal(t, "other_text_file", args.ptr.OtherTextFile.Key)
			require.NotNil(t, args.ptr.OtherTextFile.File)
			require.NotNil(t, args.ptr.OtherTextFile.Header)

			require.Empty(t, args.ptr.NoFile)
			require.Nil(t, args.ptr.NoFileAsPtr)

			require.NotEmpty(t, args.ptr.AllFiles)
			require.Equal(t, 2, len(args.ptr.AllFiles))
		})
	}
}

func BenchmarkMultipartFormData_Decode(b *testing.B) {
	r, _, _ := prepareMultipartFormDataArgs()

	m := injectStructureCache(NewMultipartFormData(WithSkipFilled[*MultipartFormData](false)))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create a fresh target for each iteration to avoid state pollution
		freshTarget := &multipartFormDataTestData{}
		if err := m.Decode(r, freshTarget); err != nil {
			b.Fatal(err)
		}
	}
}

func prepareMultipartFormDataArgs() (r *http.Request, ptr *multipartFormDataTestData, want *multipartFormDataTestData) {
	imageValue := "http://some.url"

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	err := w.WriteField("image_url", imageValue)
	if err != nil {
		panic(err)
	}

	textFileWriter, err := w.CreateFormFile("text_file", "hello.txt")
	if err != nil {
		panic(err)
	}

	_, err = textFileWriter.Write([]byte("Hello"))
	if err != nil {
		panic(err)
	}

	otherTextFileWriter, err := w.CreateFormFile("other_text_file", "other_hello.txt")
	if err != nil {
		panic(err)
	}

	_, err = otherTextFileWriter.Write([]byte("Hello2"))
	if err != nil {
		panic(err)
	}

	err = w.Close()
	if err != nil {
		panic(err)
	}

	contentType := w.FormDataContentType()

	// prepare want
	dummyReq, err := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(b.Bytes()))
	if err != nil {
		panic(err)
	}

	dummyReq.Header.Add("Content-Type", contentType)

	err = dummyReq.ParseMultipartForm(defaultMultipartFormDataMaxMemory)
	if err != nil {
		panic(err)
	}

	file, header, err := dummyReq.FormFile("text_file")
	if err != nil {
		panic(err)
	}

	textFile := MultipartFile{
		Key:    "text_file",
		File:   file,
		Header: header,
	}

	file, header, err = dummyReq.FormFile("other_text_file")
	if err != nil {
		panic(err)
	}

	otherTextFile := MultipartFile{
		Key:    "other_text_file",
		File:   file,
		Header: header,
	}

	imageAsURL, err := url.Parse(imageValue)
	if err != nil {
		panic(err)
	}

	want = &multipartFormDataTestData{
		Image:         imageValue,
		ImageAsPtr:    &imageValue,
		ImageAsURL:    *imageAsURL,
		NoKey:         "",
		NoKeyAsPtr:    nil,
		TextFile:      textFile,
		OtherTextFile: &otherTextFile,
		NoFile:        MultipartFile{},
		NoFileAsPtr:   nil,
		AllFiles: MultipartFiles{
			textFile,
			otherTextFile,
		},
	}
	//

	r, err = http.NewRequest(http.MethodPost, requestURL, &b)
	if err != nil {
		panic(err)
	}

	r.Header.Add("Content-Type", contentType)

	return r, &multipartFormDataTestData{}, want
}

// TestMultipartFormData_ErrorScenarios tests error handling scenarios
func TestMultipartFormData_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() (*http.Request, any)
		expectError bool
		errorMsg    string
	}{
		{
			name: "invalid content type",
			setup: func() (*http.Request, any) {
				req, _ := http.NewRequest(http.MethodPost, requestURL, strings.NewReader("test"))
				req.Header.Set("Content-Type", "application/json") // Wrong content type
				target := &multipartFormDataTestData{}
				return req, target
			},
			expectError: true,
			errorMsg:    "multipart",
		},
		{
			name: "malformed multipart data",
			setup: func() (*http.Request, any) {
				// Create malformed multipart data
				body := "--boundary\r\nContent-Disposition: form-data; name=\"test\"\r\n\r\nvalue" // Missing closing boundary
				req, _ := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(body))
				req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
				target := &multipartFormDataTestData{}
				return req, target
			},
			expectError: true,
			errorMsg:    "EOF",
		},
		{
			name: "nil target",
			setup: func() (*http.Request, any) {
				req, _ := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(""))
				req.Header.Set("Content-Type", "multipart/form-data; boundary=test")
				return req, nil
			},
			expectError: true,
			errorMsg:    "not a ptr",
		},
		{
			name: "unsupported target type",
			setup: func() (*http.Request, any) {
				req, _ := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(""))
				req.Header.Set("Content-Type", "multipart/form-data; boundary=test")
				target := "not a struct or map"
				return req, &target
			},
			expectError: true,
			errorMsg:    "not supported type",
		},
		{
			name: "empty request body",
			setup: func() (*http.Request, any) {
				req, _ := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(""))
				req.Header.Set("Content-Type", "multipart/form-data; boundary=test")
				target := &multipartFormDataTestData{}
				return req, target
			},
			expectError: true,
			errorMsg:    "EOF",
		},
		{
			name: "maximum memory exceeded",
			setup: func() (*http.Request, any) {
				// Create a large multipart form (larger than max memory)
				var b bytes.Buffer
				w := multipart.NewWriter(&b)

				// Create a large field
				largeData := strings.Repeat("a", 1024*1024*5) // 5MB
				w.WriteField("large_field", largeData)
				w.Close()

				req, _ := http.NewRequest(http.MethodPost, requestURL, &b)
				req.Header.Set("Content-Type", w.FormDataContentType())
				target := make(map[string]any)
				return req, &target
			},
			expectError: true,
			errorMsg:    "supported", // This will fail with "not supported type" for the large map
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, target := tt.setup()

			m := injectStructureCache(NewMultipartFormData(WithMaxMemory(1024 * 1024))) // 1MB limit
			err := m.Decode(req, target)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errorMsg))
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestMultipartFormData_ComplexStructures tests complex multipart structures
func TestMultipartFormData_ComplexStructures(t *testing.T) {
	type NestedStruct struct {
		Name  string `multipart:"name"`
		Value int    `multipart:"value"`
	}

	type ComplexStruct struct {
		SimpleField  string            `multipart:"simple"`
		NumberField  int               `multipart:"number"`
		FloatField   float64           `multipart:"float"`
		BoolField    bool              `multipart:"bool"`
		FileField    MultipartFile     `multipart:"file"`
		OptionalFile *MultipartFile    `multipart:"optional_file"`
		AllFiles     MultipartFiles    `multipart:",allfiles"`
		URLField     url.URL           `multipart:"url"`
		ArrayField   []string          `multipart:"array"`
		MapField     map[string]string `multipart:"map"`
		Nested       NestedStruct      `multipart:"nested"`
		MissingField string            `multipart:"missing"`
		SkippedField string            `multipart:"-"`
		PointerField *string           `multipart:"pointer"`
	}

	t.Run("complex structure parsing", func(t *testing.T) {
		// Create multipart form with various field types
		var b bytes.Buffer
		w := multipart.NewWriter(&b)

		// Simple fields
		w.WriteField("simple", "test value")
		w.WriteField("number", "42")
		w.WriteField("float", "3.14159")
		w.WriteField("bool", "true")
		w.WriteField("url", "https://example.com/path?query=value")
		w.WriteField("pointer", "pointer value")

		// Array field
		w.WriteField("array", "item1,item2,item3")

		// Map field (JSON-like)
		w.WriteField("map", "{\"key1\":\"value1\",\"key2\":\"value2\"}")

		// Nested struct fields
		w.WriteField("nested.name", "nested name")
		w.WriteField("nested.value", "100")

		// File field
		fileWriter, _ := w.CreateFormFile("file", "test.txt")
		fileWriter.Write([]byte("file content"))

		// Optional file field
		optFileWriter, _ := w.CreateFormFile("optional_file", "optional.txt")
		optFileWriter.Write([]byte("optional content"))

		// Additional files for allfiles
		extraWriter, _ := w.CreateFormFile("extra", "extra.txt")
		extraWriter.Write([]byte("extra content"))

		w.Close()

		req, _ := http.NewRequest(http.MethodPost, requestURL, &b)
		req.Header.Set("Content-Type", w.FormDataContentType())

		target := &ComplexStruct{}
		m := injectStructureCache(NewMultipartFormData())
		err := m.Decode(req, target)

		require.NoError(t, err)

		// Verify parsed values
		assert.Equal(t, "test value", target.SimpleField)
		assert.Equal(t, 42, target.NumberField)
		assert.InDelta(t, 3.14159, target.FloatField, 0.00001)
		assert.True(t, target.BoolField)
		assert.Equal(t, "https://example.com/path?query=value", target.URLField.String())
		assert.Equal(t, "pointer value", *target.PointerField)
		assert.Equal(t, []string{"item1", "item2", "item3"}, target.ArrayField)

		// Verify file fields
		assert.Equal(t, "file", target.FileField.Key)
		assert.NotNil(t, target.FileField.File)
		assert.NotNil(t, target.FileField.Header)

		assert.NotNil(t, target.OptionalFile)
		assert.Equal(t, "optional_file", target.OptionalFile.Key)

		// Verify all files includes both specific files and extras
		assert.Len(t, target.AllFiles, 3) // file + optional_file + extra

		// Verify missing field remains empty
		assert.Empty(t, target.MissingField)

		// Verify skipped field is not processed
		assert.Empty(t, target.SkippedField)
	})
}

// TestMultipartFormData_MapDecoding tests that maps are not supported (limitation test)
func TestMultipartFormData_MapDecoding(t *testing.T) {
	t.Run("map decoding should fail with not supported error", func(t *testing.T) {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)

		w.WriteField("field1", "value1")
		w.WriteField("field2", "value2")

		w.Close()

		req, _ := http.NewRequest(http.MethodPost, requestURL, &b)
		req.Header.Set("Content-Type", w.FormDataContentType())

		// Test that maps are not supported
		target := make(map[string]any)
		m := injectStructureCache(NewMultipartFormData())
		err := m.Decode(req, &target)

		require.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "not supported")
	})
}

// TestMultipartFormData_SkipFilledOption tests skipFilled functionality
func TestMultipartFormData_SkipFilledOption(t *testing.T) {
	type TestStruct struct {
		Field1 string `multipart:"field1"`
		Field2 string `multipart:"field2"`
		Field3 int    `multipart:"field3"`
	}

	tests := []struct {
		name       string
		skipFilled bool
		initial    TestStruct
		expected   TestStruct
	}{
		{
			name:       "skip filled enabled - prefilled values not overwritten",
			skipFilled: true,
			initial: TestStruct{
				Field1: "prefilled1",
				Field2: "", // Empty, should be filled
				Field3: 42, // Non-zero, should not be overwritten
			},
			expected: TestStruct{
				Field1: "prefilled1", // Should not change
				Field2: "value2",     // Should be filled
				Field3: 42,           // Should not change
			},
		},
		{
			name:       "skip filled disabled - all values overwritten",
			skipFilled: false,
			initial: TestStruct{
				Field1: "prefilled1",
				Field2: "prefilled2",
				Field3: 42,
			},
			expected: TestStruct{
				Field1: "value1", // Should be overwritten
				Field2: "value2", // Should be overwritten
				Field3: 123,      // Should be overwritten
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create multipart form
			var b bytes.Buffer
			w := multipart.NewWriter(&b)

			w.WriteField("field1", "value1")
			w.WriteField("field2", "value2")
			w.WriteField("field3", "123")

			w.Close()

			req, _ := http.NewRequest(http.MethodPost, requestURL, &b)
			req.Header.Set("Content-Type", w.FormDataContentType())

			target := tt.initial
			m := injectStructureCache(NewMultipartFormData(WithSkipFilled[*MultipartFormData](tt.skipFilled)))
			err := m.Decode(req, &target)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, target)
		})
	}
}

// TestMultipartFormData_LargeFile tests handling of large files
func TestMultipartFormData_LargeFile(t *testing.T) {
	t.Run("large file handling", func(t *testing.T) {
		type FileStruct struct {
			LargeFile MultipartFile `multipart:"large_file"`
		}

		// Create a reasonably large file (1MB)
		largeContent := strings.Repeat("This is a test line for large file content.\n", 20000) // ~1MB

		var b bytes.Buffer
		w := multipart.NewWriter(&b)

		fileWriter, _ := w.CreateFormFile("large_file", "large.txt")
		fileWriter.Write([]byte(largeContent))

		w.Close()

		req, _ := http.NewRequest(http.MethodPost, requestURL, &b)
		req.Header.Set("Content-Type", w.FormDataContentType())

		target := &FileStruct{}
		m := injectStructureCache(NewMultipartFormData(WithMaxMemory(1024 * 1024 * 10))) // 10MB limit
		err := m.Decode(req, target)

		require.NoError(t, err)
		assert.Equal(t, "large_file", target.LargeFile.Key)
		assert.NotNil(t, target.LargeFile.File)
		assert.NotNil(t, target.LargeFile.Header)
		assert.Equal(t, "large.txt", target.LargeFile.Header.Filename)

		// Verify file content can be read
		content, err := io.ReadAll(target.LargeFile.File)
		require.NoError(t, err)
		assert.Equal(t, len(largeContent), len(content))

		// Reset file position and read again to test reusability
		target.LargeFile.File.Seek(0, 0)
		content2, err := io.ReadAll(target.LargeFile.File)
		require.NoError(t, err)
		assert.Equal(t, content, content2)
	})
}

// TestMultipartFormData_TagParsing tests struct tag parsing
func TestMultipartFormData_TagParsing(t *testing.T) {
	type TagTestStruct struct {
		Simple         string `multipart:"simple"`
		WithUnderScore string `multipart:"with_underscore"`
		EmptyTag       string `multipart:""`
		NoTag          string
		Skipped        string         `multipart:"-"`
		AllFiles       MultipartFiles `multipart:",allfiles"`
		SpecialChars   string         `multipart:"field-with_special.chars"`
	}

	t.Run("tag parsing variations", func(t *testing.T) {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)

		w.WriteField("simple", "simple_value")
		w.WriteField("with_underscore", "underscore_value")
		w.WriteField("field-with_special.chars", "special_value")

		// Add files for allfiles
		fileWriter1, _ := w.CreateFormFile("file1", "file1.txt")
		fileWriter1.Write([]byte("file1 content"))

		fileWriter2, _ := w.CreateFormFile("file2", "file2.txt")
		fileWriter2.Write([]byte("file2 content"))

		w.Close()

		req, _ := http.NewRequest(http.MethodPost, requestURL, &b)
		req.Header.Set("Content-Type", w.FormDataContentType())

		target := &TagTestStruct{}
		m := injectStructureCache(NewMultipartFormData())
		err := m.Decode(req, target)

		require.NoError(t, err)

		assert.Equal(t, "simple_value", target.Simple)
		assert.Equal(t, "underscore_value", target.WithUnderScore)
		assert.Empty(t, target.EmptyTag) // Empty tag should not match anything
		assert.Empty(t, target.NoTag)    // No tag should not match anything
		assert.Empty(t, target.Skipped)  // Skipped field should not be processed
		assert.Equal(t, "special_value", target.SpecialChars)

		// Verify all files were collected
		assert.Len(t, target.AllFiles, 2)

		fileNames := make([]string, len(target.AllFiles))
		for i, f := range target.AllFiles {
			fileNames[i] = f.Header.Filename
		}
		assert.Contains(t, fileNames, "file1.txt")
		assert.Contains(t, fileNames, "file2.txt")
	})
}

// BenchmarkMultipartFormData_Decode_Complex benchmarks complex multipart decoding
func BenchmarkMultipartFormData_Decode_Complex(b *testing.B) {
	type BenchStruct struct {
		Field1   string         `multipart:"field1"`
		Field2   string         `multipart:"field2"`
		Field3   int            `multipart:"field3"`
		Field4   float64        `multipart:"field4"`
		Field5   bool           `multipart:"field5"`
		File1    MultipartFile  `multipart:"file1"`
		File2    MultipartFile  `multipart:"file2"`
		AllFiles MultipartFiles `multipart:",allfiles"`
		Array    []string       `multipart:"array"`
		URL      url.URL        `multipart:"url"`
	}

	// Create complex multipart form
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	w.WriteField("field1", "benchmark value 1")
	w.WriteField("field2", "benchmark value 2")
	w.WriteField("field3", "42")
	w.WriteField("field4", "3.14159")
	w.WriteField("field5", "true")
	w.WriteField("array", "item1,item2,item3,item4,item5")
	w.WriteField("url", "https://example.com/path?query=value&another=param")

	// Add files
	fileContent := strings.Repeat("benchmark file content line\n", 100)
	file1Writer, _ := w.CreateFormFile("file1", "bench1.txt")
	file1Writer.Write([]byte(fileContent))

	file2Writer, _ := w.CreateFormFile("file2", "bench2.txt")
	file2Writer.Write([]byte(fileContent))

	extraFileWriter, _ := w.CreateFormFile("extra", "extra.txt")
	extraFileWriter.Write([]byte(fileContent))

	w.Close()

	// Convert to byte slice for reusability in benchmark
	multipartData := buf.Bytes()
	contentType := w.FormDataContentType()

	m := injectStructureCache(NewMultipartFormData())

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(multipartData))
		req.Header.Set("Content-Type", contentType)

		target := &BenchStruct{}
		if err := m.Decode(req, target); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMultipartFormData_Decode_MapVsStruct compares performance of map vs struct decoding
func BenchmarkMultipartFormData_Decode_MapVsStruct(b *testing.B) {
	// Prepare multipart data
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	for i := 0; i < 20; i++ {
		w.WriteField(fmt.Sprintf("field%d", i), fmt.Sprintf("value%d", i))
	}

	w.Close()
	multipartData := buf.Bytes()
	contentType := w.FormDataContentType()

	m := injectStructureCache(NewMultipartFormData())

	b.Run("struct decoding", func(b *testing.B) {
		type StructTarget struct {
			Field0  string `multipart:"field0"`
			Field1  string `multipart:"field1"`
			Field2  string `multipart:"field2"`
			Field3  string `multipart:"field3"`
			Field4  string `multipart:"field4"`
			Field5  string `multipart:"field5"`
			Field6  string `multipart:"field6"`
			Field7  string `multipart:"field7"`
			Field8  string `multipart:"field8"`
			Field9  string `multipart:"field9"`
			Field10 string `multipart:"field10"`
			Field11 string `multipart:"field11"`
			Field12 string `multipart:"field12"`
			Field13 string `multipart:"field13"`
			Field14 string `multipart:"field14"`
			Field15 string `multipart:"field15"`
			Field16 string `multipart:"field16"`
			Field17 string `multipart:"field17"`
			Field18 string `multipart:"field18"`
			Field19 string `multipart:"field19"`
		}

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(multipartData))
			req.Header.Set("Content-Type", contentType)

			target := &StructTarget{}
			if err := m.Decode(req, target); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("map decoding", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest(http.MethodPost, requestURL, bytes.NewReader(multipartData))
			req.Header.Set("Content-Type", contentType)

			target := make(map[string]any)
			if err := m.Decode(req, &target); err != nil {
				b.Fatal(err)
			}
		}
	})
}
