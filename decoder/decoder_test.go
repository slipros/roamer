package decoder

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"testing"

	"github.com/slipros/roamer/internal/cache"
	"github.com/stretchr/testify/require"
)

const requestURL = "test.com"

func toJSON(t *testing.T, v any) io.Reader {
	var buffer bytes.Buffer
	err := json.NewEncoder(&buffer).Encode(&v)
	require.NoError(t, err, "unable convert `%T` to json", v)

	return &buffer
}

func toXML(t *testing.T, v any) io.Reader {
	var buffer bytes.Buffer
	err := xml.NewEncoder(&buffer).Encode(&v)
	require.NoError(t, err, "unable convert `%T` to xml", v)

	return &buffer
}

type decoderWithStructureCache interface {
	Tag() string
	Decode(r *http.Request, ptr any) error
	SetStructureCache(cache *cache.Structure)
}

func injectStructureCache[T decoderWithStructureCache](dec T) T {
	c := cache.NewStructure(cache.WithDecoders([]string{dec.Tag()}))
	dec.SetStructureCache(c)

	return dec
}
