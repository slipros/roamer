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

// ParsedDataFromContext extracts parsed request data from a context into the provided pointer.
//
// This function is typically used in HTTP handlers to retrieve data that was previously
// parsed and stored in the context by roamer middleware (see Middleware or SliceMiddleware).
//
// # Error Handling
//
// The function returns an error if:
//   - The pointer is nil (NilValue error)
//   - A parsing error occurred and is stored in the context
//   - No data of the expected type is found in the context (NoData error)
//
// # Type Safety
//
// The function uses generics to ensure type safety. The type parameter T must match
// the type used when storing data in the context, otherwise a NoData error is returned.
//
// Parameters:
//   - ctx: The context containing the parsed data (typically from http.Request.Context()).
//   - ptr: A pointer to the destination variable where the data will be copied.
//
// Returns:
//   - error: An error if retrieval fails, or nil if successful.
//
// Example:
//
//	type UserRequest struct {
//	    ID   int    `query:"id"`
//	    Name string `json:"name"`
//	}
//
//	func MyHandler(w http.ResponseWriter, r *http.Request) {
//	    var userData UserRequest
//	    if err := roamer.ParsedDataFromContext(r.Context(), &userData); err != nil {
//	        if errors.Is(err, err.NoData) {
//	            http.Error(w, "No user data in context", http.StatusInternalServerError)
//	            return
//	        }
//	        http.Error(w, err.Error(), http.StatusBadRequest)
//	        return
//	    }
//	    // Use userData...
//	    fmt.Fprintf(w, "Hello, %s!", userData.Name)
//	}
func ParsedDataFromContext[T any](ctx context.Context, ptr *T) error {
	if ptr == nil {
		return errors.Wrap(rerr.NilValue, "ptr")
	}

	if err, ok := ctx.Value(ContextKeyParsingError).(error); ok {
		return errors.WithStack(err)
	}

	v, ok := ctx.Value(ContextKeyParsedData).(*T)
	if !ok || v == nil {
		return errors.WithStack(rerr.NoData)
	}

	*ptr = *v
	return nil
}

// ContextWithParsedData creates a new context containing the parsed data.
//
// This function is used internally by roamer middleware (Middleware, SliceMiddleware)
// to store parsed request data in the context for downstream handlers to retrieve
// using ParsedDataFromContext.
//
// Parameters:
//   - ctx: The parent context.
//   - data: The parsed data to store (typically a pointer to a struct or slice).
//
// Returns:
//   - context.Context: A new context containing the parsed data.
func ContextWithParsedData(ctx context.Context, data any) context.Context {
	return context.WithValue(ctx, ContextKeyParsedData, data)
}

// ContextWithParsingError creates a new context containing a parsing error.
//
// This function is used internally by roamer middleware to propagate parsing errors
// to downstream handlers. The error can be retrieved using ParsedDataFromContext,
// which will return it instead of the parsed data.
//
// Parameters:
//   - ctx: The parent context.
//   - err: The parsing error that occurred.
//
// Returns:
//   - context.Context: A new context containing the error.
func ContextWithParsingError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, ContextKeyParsingError, err)
}
