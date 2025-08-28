package err

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors_Successfully(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name       string
		err        error
		wantMsg    string
		wantUnwrap error
	}{
		{
			name:       "FormatterNotFound error",
			err:        FormatterNotFound{Tag: "test_tag", Formatter: "test_formatter"},
			wantMsg:    "formatter 'test_formatter' not found for tag 'test_tag'",
			wantUnwrap: nil,
		},
		{
			name:       "DecodeError with wrapped error",
			err:        DecodeError{Err: baseErr},
			wantMsg:    "base error",
			wantUnwrap: baseErr,
		},
		{
			name:       "SliceIterationError with wrapped error",
			err:        SliceIterationError{Index: 5, Err: baseErr},
			wantMsg:    "slice element with index 5: base error",
			wantUnwrap: baseErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantMsg, tt.err.Error())
			assert.Equal(t, tt.wantUnwrap, errors.Unwrap(tt.err))
		})
	}
}
