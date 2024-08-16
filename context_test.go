package roamer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	ctxWithError := ContextWithParsingError(context.Background(), errBigBad)

	var first []string
	err := ParsedDataFromContext[[]string](ctxWithError, &first)
	require.Empty(t, first, "empty data")
	require.ErrorIs(t, err, errBigBad, "want %v, got %v", errBigBad, err)

	ctxWithData := ContextWithParsedData(context.Background(), &[]string{"1", "2"})

	var second []string
	err = ParsedDataFromContext(ctxWithData, &second)
	require.NotEmpty(t, second, "not empty data")
	require.NoError(t, err, "has error %v", err)
}
