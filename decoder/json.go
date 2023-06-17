package decoder

import (
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

const (
	// ContentTypeJSON content-type header for json decoder.
	ContentTypeJSON = "application/json"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// JSON json decoder.
type JSON struct{}

// NewJSON returns new json decoder.
func NewJSON() *JSON {
	return &JSON{}
}

// Decode decodes request body into ptr based on content-type header.
func (j *JSON) Decode(r *http.Request, ptr any) error {
	if err := json.NewDecoder(r.Body).Decode(ptr); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}

	return nil
}

// ContentType returns content-type header value.
func (j *JSON) ContentType() string {
	return ContentTypeJSON
}
