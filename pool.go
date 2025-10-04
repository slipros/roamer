package roamer

import (
	"net/http"
	"sync"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// NewParseWithPool creates a memory-efficient parser function that uses object pooling
// to reduce allocations when processing HTTP requests. This is particularly useful
// for high-throughput scenarios where creating new instances repeatedly would create
// unnecessary garbage collection pressure.
//
// The returned function parses HTTP requests into instances of type T from a sync.Pool,
// passes the parsed instance to the callback function, and then returns it to the pool
// after zeroing its fields. This ensures memory reuse while preventing data leakage
// between requests.
//
// Performance characteristics:
//   - Reduces heap allocations by reusing instances of type T
//   - Automatically zeros out fields before returning to pool
//   - Thread-safe for concurrent request processing
//   - Minimal overhead compared to direct parsing
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
//	// Create a pooled parser function
//	parseUser := roamer.NewParseWithPool[UserRequest](r)
//
//	// Use in HTTP handler
//	http.HandleFunc("/user", func(w http.ResponseWriter, req *http.Request) {
//	    err := parseUser(req, func(user *UserRequest) error {
//	        // Process the parsed user data
//	        // The user instance will be automatically returned to pool
//	        return processUser(user)
//	    })
//	    if err != nil {
//	        http.Error(w, err.Error(), http.StatusBadRequest)
//	        return
//	    }
//	})
//
// Parameters:
//   - r: The configured Roamer instance to use for parsing requests.
//
// Returns:
//   - A function that accepts an HTTP request and a callback function.
//     The callback receives a pointer to the parsed data of type T.
//     Returns an error if parsing fails or if the callback returns an error.
//
// Notes:
//   - The callback function must not retain references to the parsed instance
//     after returning, as it will be zeroed and returned to the pool.
//   - If you need to keep the data, copy it to a new instance within the callback.
//   - The pool grows dynamically based on concurrent usage patterns.
func NewParseWithPool[T any](r *Roamer) func(req *http.Request, callback func(*T) error) error {
	pool := sync.Pool{
		New: func() any {
			return new(T)
		},
	}

	put := func(ptr *T) {
		var v T
		*ptr = v
		pool.Put(ptr)
	}

	return func(req *http.Request, callback func(*T) error) error {
		if r == nil {
			return errors.Wrap(rerr.NilValue, "roamer")
		}

		if callback == nil {
			return errors.Wrap(rerr.NilValue, "callback")
		}

		result := pool.Get().(*T)
		defer put(result)

		if err := r.Parse(req, result); err != nil {
			return err
		}

		return callback(result)
	}
}
