package decoder

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/textproto"
	"testing"

	rerr "github.com/slipros/roamer/err"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockMultipartFile creates a mock implementation of multipart.File
type mockMultipartFile struct {
	content    []byte
	closeError error
	readError  error
	closed     bool
}

func (m *mockMultipartFile) Read(p []byte) (n int, err error) {
	if m.readError != nil {
		return 0, m.readError
	}
	return bytes.NewReader(m.content).Read(p)
}

func (m *mockMultipartFile) Close() error {
	m.closed = true
	return m.closeError
}

func (m *mockMultipartFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (m *mockMultipartFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, nil
}

// createTestFileHeader creates a file header for testing
func createTestFileHeader(contentType string) *multipart.FileHeader {
	header := &multipart.FileHeader{
		Filename: "test.txt",
		Header:   make(textproto.MIMEHeader),
		Size:     int64(len("test content")),
	}
	header.Header.Set("Content-Type", contentType)
	return header
}

// TestMultipartFile_ContentType_Successfully tests successful ContentType method call
func TestMultipartFile_ContentType_Successfully(t *testing.T) {
	// Create a test header with known content type
	header := createTestFileHeader("text/plain")
	file := &mockMultipartFile{content: []byte("test content")}

	mf := MultipartFile{
		Key:    "test-key",
		File:   file,
		Header: header,
	}

	// Check that ContentType returns the correct value
	assert.Equal(t, "text/plain", mf.ContentType())
}

// TestMultipartFile_IsValid_Successfully tests valid file scenarios
func TestMultipartFile_IsValid_Successfully(t *testing.T) {
	// Define test cases for valid files
	tests := []struct {
		name   string
		key    string
		file   multipart.File
		header *multipart.FileHeader
	}{
		{
			name:   "Complete valid file",
			key:    "test-key",
			file:   &mockMultipartFile{content: []byte("test content")},
			header: createTestFileHeader("text/plain"),
		},
	}

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf := MultipartFile{
				Key:    tc.key,
				File:   tc.file,
				Header: tc.header,
			}

			// IsValid should return true for valid files with the new logic
			assert.True(t, mf.IsValid(), "File should be valid with all fields properly set")
		})
	}
}

// TestMultipartFile_IsValid_Failure tests invalid file scenarios
func TestMultipartFile_IsValid_Failure(t *testing.T) {
	// Define test cases for invalid files
	tests := []struct {
		name   string
		key    string
		file   multipart.File
		header *multipart.FileHeader
		reason string
	}{
		{
			name:   "Empty key",
			key:    "",
			file:   &mockMultipartFile{content: []byte("test content")},
			header: createTestFileHeader("text/plain"),
			reason: "File should be invalid with empty key",
		},
		{
			name:   "Nil file",
			key:    "test-key",
			file:   nil,
			header: createTestFileHeader("text/plain"),
			reason: "File should be invalid with nil file",
		},
		{
			name:   "Nil header",
			key:    "test-key",
			file:   &mockMultipartFile{content: []byte("test content")},
			header: nil,
			reason: "File should be invalid with nil header",
		},
	}

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mf := MultipartFile{
				Key:    tc.key,
				File:   tc.file,
				Header: tc.header,
			}

			// IsValid should return false for invalid files with the new logic
			assert.False(t, mf.IsValid(), tc.reason)
		})
	}
}

// TestMultipartFiles_Close_Successfully tests successful closing of files
func TestMultipartFiles_Close_Successfully(t *testing.T) {
	tests := []struct {
		name       string
		setupFiles func() (MultipartFiles, []*mockMultipartFile)
	}{
		{
			name: "Close all files successfully",
			setupFiles: func() (MultipartFiles, []*mockMultipartFile) {
				file1 := &mockMultipartFile{content: []byte("content1")}
				file2 := &mockMultipartFile{content: []byte("content2")}
				file3 := &mockMultipartFile{content: []byte("content3")}

				header1 := createTestFileHeader("text/plain")
				header2 := createTestFileHeader("text/html")
				header3 := createTestFileHeader("application/json")

				files := MultipartFiles{
					{Key: "file1", File: file1, Header: header1},
					{Key: "file2", File: file2, Header: header2},
					{Key: "file3", File: file3, Header: header3},
				}

				return files, []*mockMultipartFile{file1, file2, file3}
			},
		},
		{
			name: "Empty files slice",
			setupFiles: func() (MultipartFiles, []*mockMultipartFile) {
				return MultipartFiles{}, []*mockMultipartFile{}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			files, mockFiles := tc.setupFiles()

			// Close the files
			err := files.Close()

			// Check results
			require.NoError(t, err)

			// Check that all files are closed if there are any
			if len(mockFiles) > 0 {
				for i, file := range mockFiles {
					assert.True(t, file.closed, "File %d should be closed", i+1)
				}
			}
		})
	}
}

// TestMultipartFiles_Close_Failure tests failure scenarios for closing files
func TestMultipartFiles_Close_Failure(t *testing.T) {
	tests := []struct {
		name         string
		setupFiles   func() (MultipartFiles, []*mockMultipartFile)
		errorIndex   int
		expectedErr  error
		checkClosing func(t *testing.T, mockFiles []*mockMultipartFile)
	}{
		{
			name: "Error on closing a file",
			setupFiles: func() (MultipartFiles, []*mockMultipartFile) {
				file1 := &mockMultipartFile{content: []byte("content1")}
				file2 := &mockMultipartFile{content: []byte("content2"), closeError: io.ErrClosedPipe}
				file3 := &mockMultipartFile{content: []byte("content3")}

				header1 := createTestFileHeader("text/plain")
				header2 := createTestFileHeader("text/html")
				header3 := createTestFileHeader("application/json")

				files := MultipartFiles{
					{Key: "file1", File: file1, Header: header1},
					{Key: "file2", File: file2, Header: header2}, // this file will return an error when closed
					{Key: "file3", File: file3, Header: header3},
				}

				return files, []*mockMultipartFile{file1, file2, file3}
			},
			errorIndex:  1,
			expectedErr: io.ErrClosedPipe,
			checkClosing: func(t *testing.T, mockFiles []*mockMultipartFile) {
				assert.True(t, mockFiles[0].closed, "File 1 should be closed")
				assert.True(t, mockFiles[1].closed, "File 2 should be closed")
				assert.False(t, mockFiles[2].closed, "File 3 should not be closed due to error on file 2")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			files, mockFiles := tc.setupFiles()

			// Close the files
			err := files.Close()

			// Check error
			require.Error(t, err)

			var sliceErr rerr.SliceIterationError
			assert.ErrorAs(t, err, &sliceErr)
			assert.Equal(t, tc.errorIndex, sliceErr.Index, "Error should occur at the expected index")
			assert.ErrorIs(t, sliceErr.Err, tc.expectedErr)

			// Check the state of files
			if tc.checkClosing != nil {
				tc.checkClosing(t, mockFiles)
			}
		})
	}
}

// TestMultipartFile_Copy_Successfully tests successful copying of MultipartFile
func TestMultipartFile_Copy_Successfully(t *testing.T) {
	// Create a real multipart file header with actual file content
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create a form file
	fileWriter, err := writer.CreateFormFile("testfile", "test.txt")
	require.NoError(t, err)

	testContent := []byte("test file content")
	_, err = fileWriter.Write(testContent)
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	// Parse the multipart form
	reader := multipart.NewReader(&buf, writer.Boundary())
	form, err := reader.ReadForm(1024 * 1024)
	require.NoError(t, err)
	defer form.RemoveAll()

	// Get the file header
	require.NotEmpty(t, form.File["testfile"])
	fileHeader := form.File["testfile"][0]

	// Open the original file
	originalFile, err := fileHeader.Open()
	require.NoError(t, err)
	defer originalFile.Close()

	// Create MultipartFile
	mf := MultipartFile{
		Key:    "testfile",
		File:   originalFile,
		Header: fileHeader,
	}

	// Copy the file
	copy, err := mf.Copy()
	require.NoError(t, err)
	require.NotNil(t, copy.File)
	defer copy.File.Close()

	// Verify the copy has the same metadata
	assert.Equal(t, mf.Key, copy.Key)
	assert.Equal(t, mf.Header, copy.Header)

	// Verify the copy can read the content
	copiedContent, err := io.ReadAll(copy.File)
	require.NoError(t, err)
	assert.Equal(t, testContent, copiedContent)
}

// TestMultipartFile_Copy_Failure tests failure scenarios when copying MultipartFile
func TestMultipartFile_Copy_Failure(t *testing.T) {
	// Create a file header that will fail to open
	header := &multipart.FileHeader{
		Filename: "nonexistent.txt",
		Header:   make(textproto.MIMEHeader),
		Size:     0,
	}

	// Create MultipartFile with a header that can't be reopened
	mf := MultipartFile{
		Key:    "testfile",
		File:   &mockMultipartFile{content: []byte("test")},
		Header: header,
	}

	// Attempt to copy - should fail because Header.Open() will fail
	copy, err := mf.Copy()
	assert.Error(t, err)
	assert.Empty(t, copy.Key)
	assert.Nil(t, copy.File)
}
