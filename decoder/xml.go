package decoder

import (
	"encoding/xml"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const (
	ContentTypeXML = "application/xml"
)

type XML struct{}

func NewXML() *XML {
	return &XML{}
}

func (x *XML) ContentType() string {
	return ContentTypeXML
}

func (x *XML) Decode(r *http.Request, ptr any) error {
	if err := xml.NewDecoder(r.Body).Decode(ptr); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}

	return nil
}
