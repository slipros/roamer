package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/parser"
)

// ProductRequest defines the structure for product operations
type ProductRequest struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `query:"category" default:"general"`
}

// ProductResponse represents the API response
type ProductResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

func main() {
	// Initialize roamer
	r := roamer.NewRoamer(
		roamer.WithDecoders(decoder.NewJSON()),
		roamer.WithParsers(parser.NewQuery()),
	)

	// Use roamer as middleware
	http.Handle("/products", roamer.Middleware[ProductRequest](r)(http.HandlerFunc(handleCreateProduct)))

	log.Println("Server starting on :8080")
	log.Println("Try: curl -X POST http://localhost:8080/products?category=electronics \\")
	log.Println("  -H 'Content-Type: application/json' \\")
	log.Println("  -d '{\"name\":\"Laptop\",\"price\":999.99}'")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleCreateProduct(w http.ResponseWriter, req *http.Request) {
	var productReq ProductRequest

	// Get parsed data from context
	if err := roamer.ParsedDataFromContext(req.Context(), &productReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create response
	response := ProductResponse{
		ID:       "product-456",
		Name:     productReq.Name,
		Price:    productReq.Price,
		Category: productReq.Category,
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
