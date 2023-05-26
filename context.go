package roamer

import (
	"context"

	"github.com/pkg/errors"

	roamerError "github.com/SLIpros/roamer/error"
)

// ContextKey context key.
type ContextKey uint8

const (
	// ContextKeyData parsed data.
	ContextKeyData ContextKey = iota + 1
	// ContextKeyError parsing error.
	ContextKeyError
)

// Data return parsed data.
func Data[T any](ctx context.Context, ptr *T) error {
	if ptr == nil {
		return errors.WithMessage(roamerError.ErrNil, "context")
	}

	if err, ok := ctx.Value(ContextKeyError).(error); ok {
		return err
	}

	v, ok := ctx.Value(ContextKeyData).(*T)
	if !ok {
		return roamerError.ErrNoData
	}

	*ptr = *v
	return nil
}

// SetData set parsed data to context.
func SetData(ctx context.Context, value any) context.Context {
	return context.WithValue(ctx, ContextKeyData, value)
}

// setError set parsing error to context.
func setError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, ContextKeyError, err)
}
