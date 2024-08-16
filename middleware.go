package roamer

import "net/http"

// Middleware parse http request and saves the received value/error to context.
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
