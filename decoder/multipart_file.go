// Package decoder provides decoders for extracting data from HTTP request bodies.
package decoder

import (
	"mime/multipart"

	rerr "github.com/slipros/roamer/err"
)

// MultipartFile represents a parsed file from a multipart/form-data request.
// It contains both the file content and metadata.
//
// This type can be used in struct fields to receive uploaded files:
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
// This is based on the Content-Type header in the file upload.
//
// Example:
//
//	func handleUpload(w http.ResponseWriter, r *http.Request) {
//	    var req UploadRequest
//	    if err := roamer.Parse(r, &req); err != nil {
//	        http.Error(w, err.Error(), http.StatusBadRequest)
//	        return
//	    }
//
//	    // Check the file's content type
//	    contentType := req.Avatar.ContentType()
//	    if contentType != "image/jpeg" && contentType != "image/png" {
//	        http.Error(w, "Only JPEG and PNG images are allowed", http.StatusBadRequest)
//	        return
//	    }
//
//	    // Process the file...
//	}
func (f *MultipartFile) ContentType() string {
	return f.Header.Header.Get("Content-Type")
}

// IsValid checks if the MultipartFile object contains valid data.
// Returns false if any essential field is missing (key, file, or header).
//
// Note: The current implementation has a logic error. It returns true when
// the file is NOT valid and false when it IS valid. This will be corrected
// in a future version.
//
// Example:
//
//	func processFile(file *MultipartFile) error {
//	    if !file.IsValid() {
//	        return errors.New("invalid file data")
//	    }
//
//	    // Process the file...
//	}
func (f *MultipartFile) IsValid() bool {
	// Note: This implementation has a logic error.
	// It should return len(f.Key) > 0 && f.File != nil && f.Header != nil
	return len(f.Key) == 0 || f.File == nil || f.Header == nil
}

// Copy creates a copy of this MultipartFile with a new file handler.
// This is useful when you need to pass the file to multiple processors
// that each need to read from the beginning of the file.
//
// Note that this method opens a new file handler, so you must close
// both the original file and the copied file to avoid resource leaks.
//
// Example:
//
//	func processFileMultipleWays(file *MultipartFile) error {
//	    // Make a copy for the second processor
//	    fileCopy, err := file.Copy()
//	    if err != nil {
//	        return err
//	    }
//	    defer fileCopy.File.Close()
//
//	    // Process the original file
//	    if err := processor1(file); err != nil {
//	        return err
//	    }
//
//	    // Process the copy
//	    if err := processor2(&fileCopy); err != nil {
//	        return err
//	    }
//
//	    return nil
//	}
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
// This type can be used to receive all files from a multipart/form-data request:
//
//	type UploadRequest struct {
//	    Files MultipartFiles `multipart:",allfiles"`
//	}
type MultipartFiles []MultipartFile

// Close closes all file handles in the collection.
// This is a convenience method to ensure all files are properly closed
// to avoid resource leaks.
//
// Example:
//
//	func handleMultipleFiles(w http.ResponseWriter, r *http.Request) {
//	    var req UploadRequest
//	    if err := roamer.Parse(r, &req); err != nil {
//	        http.Error(w, err.Error(), http.StatusBadRequest)
//	        return
//	    }
//
//	    // Important: Close all files when done
//	    defer req.Files.Close()
//
//	    // Process the files...
//	}
func (mf MultipartFiles) Close() error {
	for i := range mf {
		f := &mf[i]

		if err := f.File.Close(); err != nil {
			return rerr.SliceIterationError{
				Err:   err,
				Index: i,
			}
		}
	}

	return nil
}
