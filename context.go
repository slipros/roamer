package roamer

import (
	"context"

	"github.com/pkg/errors"

	rerr "github.com/slipros/roamer/err"
)

// ContextKey represents a type-safe key for context values used by the roamer package.
// Using a custom type for context keys helps prevent collisions with other packages.
type ContextKey uint8

const (
	// ContextKeyParsedData is the context key for storing parsed data from HTTP requests.
	// This key is used to retrieve the parsed data in handlers after middleware processing.
	ContextKeyParsedData ContextKey = iota + 1

	// ContextKeyParsingError is the context key for storing any parsing errors
	// that occurred during request processing. This allows propagating parsing
	// errors to downstream handlers.
	ContextKeyParsingError
)

// ParsedDataFromContext extracts parsed data from a context into the provided pointer.
// Typically used in HTTP handlers to retrieve data processed by roamer middleware.
//
// Returns an error if:
// - The pointer is nil
// - A parsing error exists in the context
// - No data of the expected type is found
//
// Example:
//
//	func MyHandler(w http.ResponseWriter, r *http.Request) {
//	    var userData UserData
//	    if err := roamer.ParsedDataFromContext(r.Context(), &userData); err != nil {
//	        http.Error(w, err.Error(), http.StatusBadRequest)
//	        return
//	    }
//	    // Use userData...
//	}
func ParsedDataFromContext[T any](ctx context.Context, ptr *T) error {
	if ptr == nil {
		return errors.Wrap(rerr.NilValue, "ptr")
	}

	if err, ok := ctx.Value(ContextKeyParsingError).(error); ok {
		return errors.WithStack(err)
	}

	v, ok := ctx.Value(ContextKeyParsedData).(*T)
	if !ok {
		return errors.WithStack(rerr.NoData)
	}

	*ptr = *v
	return nil
}

// ContextWithParsedData creates a new context containing the parsed data.
// Used internally by roamer middleware to store data for downstream handlers.
func ContextWithParsedData(ctx context.Context, data any) context.Context {
	return context.WithValue(ctx, ContextKeyParsedData, data)
}

// ContextWithParsingError creates a new context containing a parsing error.
// Used internally by roamer middleware to propagate errors to downstream handlers.
func ContextWithParsingError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, ContextKeyParsingError, err)
}
