package decoder

import "mime/multipart"

// MultipartFile parsed multipart form-data file.
type MultipartFile struct {
	Key    string
	File   multipart.File
	Header *multipart.FileHeader
}

// ContentType returns content type of parsed file.
func (f *MultipartFile) ContentType() string {
	return f.Header.Header.Get("Content-Type")
}

// MultipartFiles parsed multipart form-data files.
type MultipartFiles []*MultipartFile

// Close close all files.
func (mf MultipartFiles) Close() error {
	for _, f := range mf {
		if err := f.File.Close(); err != nil {
			return err
		}
	}

	return nil
}
