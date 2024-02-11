package decoder

import (
	"encoding/xml"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const (
	// ContentTypeXML content-type header for xml decoder.
	ContentTypeXML = "application/xml"
)

// XMLOptionsFunc function for setting xml options.
type XMLOptionsFunc = func(*XML)

// XML xml decoder.
type XML struct {
	contentType string
}

// NewXML returns new xml decoder.
func NewXML(opts ...XMLOptionsFunc) *XML {
	x := XML{
		contentType: ContentTypeXML,
	}

	for _, opt := range opts {
		opt(&x)
	}

	return &x
}

// Decode decodes request body into ptr.
func (x *XML) Decode(r *http.Request, ptr any) error {
	if err := xml.NewDecoder(r.Body).Decode(ptr); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}

	return nil
}

// ContentType returns content-type header value.
func (x *XML) ContentType() string {
	return x.contentType
}

// setContentType set content-type value.
func (x *XML) setContentType(contentType string) {
	x.contentType = contentType
}
