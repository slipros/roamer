package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/parser"
	rhttprouter "github.com/slipros/roamer/pkg/httprouter"
)

// CreateItemRequest for item creation
type CreateItemRequest struct {
	ID    string  `path:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// ItemResponse represents the API response
type ItemResponse struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func main() {
	// Initialize HttpRouter
	router := httprouter.New()

	// Initialize roamer with HttpRouter path parser
	r := roamer.NewRoamer(
		roamer.WithDecoders(decoder.NewJSON()),
		roamer.WithParsers(
			parser.NewPath(rhttprouter.Path),
		),
	)

	// Helper middleware chain
	chain := func(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			for i := len(middlewares) - 1; i >= 0; i-- {
				next = middlewares[i](next)
			}
			return next
		}
	}

	// Define routes with middleware
	router.Handler("POST", "/items/:id", chain(
		roamer.Middleware[CreateItemRequest](r),
	)(http.HandlerFunc(handleCreateItem)))

	log.Println("Server starting on :8080")
	log.Println("Try: curl -X POST http://localhost:8080/items/item-123 \\")
	log.Println("  -H 'Content-Type: application/json' \\")
	log.Println("  -d '{\"name\":\"Laptop\",\"price\":999.99}'")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleCreateItem(w http.ResponseWriter, req *http.Request) {
	var itemReq CreateItemRequest

	if err := roamer.ParsedDataFromContext(req.Context(), &itemReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := ItemResponse{
		ID:    itemReq.ID,
		Name:  itemReq.Name,
		Price: itemReq.Price,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
