package decoder

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewMultipartFormData(t *testing.T) {
	m := NewMultipartFormData()
	require.NotNil(t, m)
	require.Equal(t, ContentTypeMultipartFormData, m.ContentType())

	m = NewMultipartFormData(WithMaxMemory(1000))
	require.NotNil(t, m)
	require.Equal(t, ContentTypeMultipartFormData, m.ContentType())
	require.Equal(t, int64(1000), m.maxMemory)

	m = NewMultipartFormData(WithContentType[*MultipartFormData]("test"))
	require.NotNil(t, m)
	require.Equal(t, "test", m.ContentType())
}

func TestMultipartFormData_Decode(t *testing.T) {
	type args struct {
		r   *http.Request
		ptr any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMultipartFormData()
			if err := m.Decode(tt.args.r, tt.args.ptr); !tt.wantErr && err != nil {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
