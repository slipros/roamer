package decoder

import (
	"io"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

const (
	ContentTypeJSON = "application/json"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type JSON struct{}

func NewJSON() *JSON {
	return &JSON{}
}

func (j *JSON) ContentType() string {
	return ContentTypeJSON
}

func (j *JSON) Decode(r *http.Request, ptr any) error {
	if err := json.NewDecoder(r.Body).Decode(ptr); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}

	return nil
}
