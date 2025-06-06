// Package roamer implements HTTP request parsing utilities.
package roamer

import "net/http"

// Middleware creates an HTTP middleware that parses requests into a specified type
// and stores the result in the request context for later retrieval with ParsedDataFromContext.
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
//	http.Handle("/users", roamer.Middleware[UserRequest](r)(
//	    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	        var data UserRequest
//	        if err := roamer.ParsedDataFromContext(r.Context(), &data); err != nil {
//	            http.Error(w, err.Error(), http.StatusBadRequest)
//	            return
//	        }
//	        fmt.Fprintf(w, "Hello, %s!", data.Name)
//	    }),
//	))
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

// SliceMiddleware creates an HTTP middleware that parses the request into a slice
// of specified type and stores the result in the request context. This is particularly
// useful for endpoints that handle arrays of objects (e.g., batch operations).
//
// Example:
//
//	// Define a data structure for your API endpoint
//	type Product struct {
//	    ID    int     `json:"id"`
//	    Name  string  `json:"name"`
//	    Price float64 `json:"price"`
//	}
//
//	// Create a handler for batch product creation
//	r := roamer.NewRoamer(roamer.WithDecoders(decoder.NewJSON()))
//
//	http.Handle("/products/batch", roamer.SliceMiddleware[Product](r)(
//	    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	        // Get parsed data from context
//	        var products []Product
//	        if err := roamer.ParsedDataFromContext(r.Context(), &products); err != nil {
//	            http.Error(w, err.Error(), http.StatusBadRequest)
//	            return
//	        }
//
//	        // Process the batch of products
//	        fmt.Fprintf(w, "Received %d products", len(products))
//	    }),
//	))
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
