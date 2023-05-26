package decoder

import (
	"bytes"
	"encoding/xml"
	"io"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

var errBigBad = errors.New("err big bad")

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
