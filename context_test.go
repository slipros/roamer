package roamer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsedDataFromContext_Failure(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		ptr         *[]string
		expectError bool
		errorCheck  func(t *testing.T, err error)
	}{
		{
			name:        "nil pointer",
			ctx:         context.Background(),
			ptr:         nil,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "ptr")
			},
		},
		{
			name:        "parsing error in context",
			ctx:         ContextWithParsingError(context.Background(), errBigBad),
			ptr:         &[]string{},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				require.ErrorIs(t, err, errBigBad)
			},
		},
		{
			name:        "no data in context",
			ctx:         context.Background(),
			ptr:         &[]string{},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name:        "wrong type in context",
			ctx:         ContextWithParsedData(context.Background(), "not a slice"),
			ptr:         &[]string{},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data []string
			if tt.ptr != nil {
				data = *tt.ptr
			}

			err := ParsedDataFromContext(tt.ctx, tt.ptr)

			if tt.expectError {
				tt.errorCheck(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.ptr != nil {
				require.Equal(t, data, *tt.ptr, "data should not be modified on error")
			}
		})
	}
}

func TestParsedDataFromContext_Successfully(t *testing.T) {
	ctxWithData := ContextWithParsedData(context.Background(), &[]string{"1", "2"})

	var second []string
	err := ParsedDataFromContext(ctxWithData, &second)
	require.NotEmpty(t, second, "not empty data")
	require.NoError(t, err, "has error %v", err)
}
