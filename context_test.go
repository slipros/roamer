package roamer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBody(t *testing.T) {
	ctxWithError := setError(context.Background(), errBigBad)

	var first []string
	err := Data[[]string](ctxWithError, &first)
	require.Empty(t, first, "empty data")
	require.ErrorIs(t, err, errBigBad, "want %v, got %v", errBigBad, err)

	ctxWithData := SetData(context.Background(), &[]string{"1", "2"})

	var second []string
	err = Data(ctxWithData, &second)
	require.NotEmpty(t, second, "not empty data")
	require.NoError(t, err, "has error %v", err)
}
