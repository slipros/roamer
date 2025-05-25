package decoder

import (
	"mime/multipart"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// MultipartFile represents an uploaded file from a multipart/form-data request.
// Contains both file content and metadata.
//
// Example:
//
//	type UploadRequest struct {
//	    Avatar *MultipartFile `multipart:"avatar"`
//	}
type MultipartFile struct {
	// Key is the form field name associated with this file.
	Key string

	// File is the actual file content, which can be read as an io.Reader.
	// Remember to close this file when you're done with it to avoid resource leaks.
	File multipart.File

	// Header contains metadata about the uploaded file,
	// such as the original filename, content type, and size.
	Header *multipart.FileHeader
}

// ContentType returns the content type of the uploaded file.
//
// Example:
//
//	// Check file type
//	contentType := req.Avatar.ContentType()
//	if contentType != "image/jpeg" && contentType != "image/png" {
//	    http.Error(w, "Only JPEG and PNG allowed", http.StatusBadRequest)
//	    return
//	}
func (f *MultipartFile) ContentType() string {
	return f.Header.Header.Get("Content-Type")
}

// IsValid checks if the MultipartFile contains valid data.
// Returns false if any essential field is missing.
func (f *MultipartFile) IsValid() bool {
	return len(f.Key) > 0 && f.File != nil && f.Header != nil
}

// Copy creates a duplicate of this MultipartFile with a new file handler.
// Useful when multiple processors need to read from the beginning of the file.
// Note: Both original and copied files must be closed to avoid resource leaks.
func (f *MultipartFile) Copy() (MultipartFile, error) {
	file, err := f.Header.Open()
	if err != nil {
		return MultipartFile{}, err
	}

	cp := *f
	cp.File = file

	return cp, nil
}

// MultipartFiles is a collection of MultipartFile objects.
// Use this to receive all files from a multipart request:
//
//	type UploadRequest struct {
//	    Files MultipartFiles `multipart:",allfiles"`
//	}
type MultipartFiles []MultipartFile

// Close closes all file handles in the collection to prevent resource leaks.
//
// Example:
//
//	func handleUpload(w http.ResponseWriter, r *http.Request) {
//	    var req UploadRequest
//	    if err := roamer.Parse(r, &req); err != nil {
//	        return
//	    }
//	    defer req.Files.Close() // Important: close all files
//	}
func (mf MultipartFiles) Close() error {
	for i := range mf {
		f := &mf[i]

		if err := f.File.Close(); err != nil {
			return errors.WithStack(rerr.SliceIterationError{
				Err:   err,
				Index: i,
			})
		}
	}

	return nil
}
