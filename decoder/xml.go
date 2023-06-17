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

// XML xml decoder.
type XML struct{}

// NewXML returns new xml decoder.
func NewXML() *XML {
	return &XML{}
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
	return ContentTypeXML
}
