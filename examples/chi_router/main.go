package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/parser"
	rchi "github.com/slipros/roamer/pkg/chi"
)

// CreateProductRequest for product creation
type CreateProductRequest struct {
	ID          string `path:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `query:"category"`
}

// ProductResponse represents the API response
type ProductResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

func main() {
	// Initialize Chi router
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	// Initialize roamer with Chi path parser
	roamerInstance := roamer.NewRoamer(
		roamer.WithDecoders(decoder.NewJSON()),
		roamer.WithParsers(
			parser.NewHeader(),
			parser.NewQuery(),
			parser.NewPath(rchi.NewPath(router)),
		),
	)

	// Define routes
	router.Route("/products", func(r chi.Router) {
		r.With(roamer.Middleware[CreateProductRequest](roamerInstance)).
			Post("/{id}", handleCreateProduct)
	})

	log.Println("Server starting on :8080")
	log.Println("Try: curl -X POST http://localhost:8080/products/prod-123?category=electronics \\")
	log.Println("  -H 'Content-Type: application/json' \\")
	log.Println("  -d '{\"name\":\"Laptop\",\"description\":\"High-end laptop\"}'")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var productReq CreateProductRequest

	if err := roamer.ParsedDataFromContext(r.Context(), &productReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := ProductResponse{
		ID:          productReq.ID,
		Name:        productReq.Name,
		Description: productReq.Description,
		Category:    productReq.Category,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
