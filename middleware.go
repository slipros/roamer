package roamer

import "net/http"

// Middleware creates an HTTP middleware that parses requests into a specified type
// and stores the result in the request context for downstream handlers to retrieve.
//
// The middleware parses the incoming HTTP request using the provided Roamer instance,
// stores the parsed data (or error) in the request context, and then calls the next
// handler in the chain. Downstream handlers can retrieve the parsed data using
// ParsedDataFromContext.
//
// # Type Parameter
//
//   - T: The type to parse the request into. Must be a struct type.
//
// # Behavior
//
//   - If roamer is nil, the middleware passes through without parsing
//   - On successful parsing, stores parsed data in context with ContextKeyParsedData
//   - On parsing error, stores error in context with ContextKeyParsingError
//   - Always calls the next handler, even if parsing fails (error handling is delegated)
//
// # Error Handling
//
// The middleware does NOT stop the request chain on parsing errors. Instead, it stores
// the error in the context and delegates error handling to downstream handlers. This
// allows handlers to decide how to respond to parsing errors.
//
// Parameters:
//   - roamer: The Roamer instance to use for parsing. If nil, middleware is a no-op.
//
// Returns:
//   - func(next http.Handler) http.Handler: A middleware function that wraps the next handler.
//
// Example:
//
//	type UserRequest struct {
//	    ID   int    `query:"id"`
//	    Name string `json:"name"`
//	}
//
//	r := roamer.NewRoamer(
//	    roamer.WithDecoders(decoder.NewJSON()),
//	    roamer.WithParsers(parser.NewQuery()),
//	)
//
//	// Use with http.Handle
//	http.Handle("/users", roamer.Middleware[UserRequest](r)(
//	    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	        var data UserRequest
//	        if err := roamer.ParsedDataFromContext(r.Context(), &data); err != nil {
//	            http.Error(w, err.Error(), http.StatusBadRequest)
//	            return
//	        }
//	        fmt.Fprintf(w, "Hello, %s (ID: %d)!", data.Name, data.ID)
//	    }),
//	))
//
//	// Or with chi router
//	router := chi.NewRouter()
//	router.Use(roamer.Middleware[UserRequest](r))
//	router.Post("/users", func(w http.ResponseWriter, r *http.Request) {
//	    var data UserRequest
//	    if err := roamer.ParsedDataFromContext(r.Context(), &data); err != nil {
//	        http.Error(w, err.Error(), http.StatusBadRequest)
//	        return
//	    }
//	    // Process user...
//	})
func Middleware[T any](roamer *Roamer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if roamer == nil {
				next.ServeHTTP(w, r)
				return
			}

			var v T
			if err := roamer.Parse(r, &v); err != nil {
				ctxWithError := ContextWithParsingError(r.Context(), err)
				next.ServeHTTP(w, r.WithContext(ctxWithError))
				return
			}

			ctxWithData := ContextWithParsedData(r.Context(), &v)
			next.ServeHTTP(w, r.WithContext(ctxWithData))
		})
	}
}

// SliceMiddleware creates an HTTP middleware that parses the request body into a slice
// of the specified type and stores the result in the request context.
//
// This middleware is particularly useful for API endpoints that handle arrays of objects,
// such as batch operations, bulk updates, or list submissions. It parses the request body
// (typically JSON) into a slice and makes it available to downstream handlers via the context.
//
// # Type Parameter
//
//   - T: The element type of the slice. The request body will be parsed into []T.
//
// # Behavior
//
//   - If roamer is nil, the middleware passes through without parsing
//   - Parses the request body into a slice of type []T
//   - On success, stores the slice in context with ContextKeyParsedData
//   - On error, stores the error in context with ContextKeyParsingError
//   - Always calls the next handler (error handling is delegated)
//
// # Use Cases
//
//   - Batch creation endpoints: POST /api/users/batch with [{...}, {...}, ...]
//   - Bulk update operations: PUT /api/products/bulk with array of products
//   - List submission forms: POST /api/tasks with array of task items
//
// Parameters:
//   - roamer: The Roamer instance to use for parsing. If nil, middleware is a no-op.
//
// Returns:
//   - func(next http.Handler) http.Handler: A middleware function that wraps the next handler.
//
// Example:
//
//	type Product struct {
//	    ID    int     `json:"id"`
//	    Name  string  `json:"name"`
//	    Price float64 `json:"price"`
//	}
//
//	r := roamer.NewRoamer(roamer.WithDecoders(decoder.NewJSON()))
//
//	// Batch product creation endpoint
//	http.Handle("/products/batch", roamer.SliceMiddleware[Product](r)(
//	    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	        var products []Product
//	        if err := roamer.ParsedDataFromContext(r.Context(), &products); err != nil {
//	            http.Error(w, err.Error(), http.StatusBadRequest)
//	            return
//	        }
//
//	        // Validate and process the batch
//	        if len(products) == 0 {
//	            http.Error(w, "Empty product list", http.StatusBadRequest)
//	            return
//	        }
//
//	        for i, product := range products {
//	            // Process each product
//	            fmt.Printf("Processing product %d: %s\n", i, product.Name)
//	        }
//
//	        fmt.Fprintf(w, "Successfully processed %d products", len(products))
//	    }),
//	))
//
//	// Request body example:
//	// [
//	//   {"id": 1, "name": "Widget", "price": 9.99},
//	//   {"id": 2, "name": "Gadget", "price": 19.99}
//	// ]
func SliceMiddleware[T any](roamer *Roamer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if roamer == nil {
				next.ServeHTTP(w, r)
				return
			}

			var v []T
			if err := roamer.Parse(r, &v); err != nil {
				ctxWithError := ContextWithParsingError(r.Context(), err)
				next.ServeHTTP(w, r.WithContext(ctxWithError))
				return
			}

			ctxWithData := ContextWithParsedData(r.Context(), &v)
			next.ServeHTTP(w, r.WithContext(ctxWithData))
		})
	}
}
