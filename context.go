// Package roamer provides a flexible HTTP request parser.
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

// ParsedDataFromContext extracts parsed data of type T from a context and assigns it
// to the provided pointer. This function is typically used in HTTP handlers to retrieve
// data that was previously parsed and stored by roamer middleware.
//
// The function returns an error if:
// - The provided pointer is nil
// - A parsing error was previously stored in the context
// - No data of the expected type is found in the context
//
// Example:
//
//	func MyHandler(w http.ResponseWriter, r *http.Request) {
//	    // Extract parsed data from the request context
//	    var userData UserData
//	    if err := roamer.ParsedDataFromContext(r.Context(), &userData); err != nil {
//	        // Handle error (e.g., return an appropriate HTTP error)
//	        http.Error(w, err.Error(), http.StatusBadRequest)
//	        return
//	    }
//
//	    // Use the parsed data
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
	if !ok {
		return errors.WithStack(rerr.NoData)
	}

	*ptr = *v
	return nil
}

// ContextWithParsedData creates a new context containing the parsed data.
// This function is primarily used by roamer middleware to store successfully
// parsed data in the request context for downstream handlers.
//
// Example:
//
//	// Used internally by middleware
//	parsedData := &MyStruct{...}
//	newCtx := roamer.ContextWithParsedData(ctx, parsedData)
//	// Pass the new context to the next handler
func ContextWithParsedData(ctx context.Context, data any) context.Context {
	return context.WithValue(ctx, ContextKeyParsedData, data)
}

// ContextWithParsingError creates a new context containing a parsing error.
// This function is primarily used by roamer middleware to store any parsing
// errors that occurred during request processing, allowing downstream handlers
// to detect and handle these errors.
//
// Example:
//
//	// Used internally by middleware
//	if err := parser.Parse(req); err != nil {
//	    newCtx := roamer.ContextWithParsingError(ctx, err)
//	    // Pass the new context with error to the next handler
//	}
func ContextWithParsingError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, ContextKeyParsingError, err)
}
