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

// JSONOptionsFunc function for setting xml options.
type JSONOptionsFunc = func(*JSON)

// JSON json decoder.
type JSON struct {
	contentType string
}

// NewJSON returns new json decoder.
func NewJSON(opts ...JSONOptionsFunc) *JSON {
	j := JSON{
		contentType: ContentTypeJSON,
	}

	for _, opt := range opts {
		opt(&j)
	}

	return &j
}

// Decode decodes request body into ptr.
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
	return j.contentType
}

// setContentType set content-type value.
func (j *JSON) setContentType(contentType string) {
	j.contentType = contentType
}
