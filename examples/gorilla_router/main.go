package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/parser"
	rgorilla "github.com/slipros/roamer/pkg/gorilla"
)

// GetOrderRequest for order retrieval
type GetOrderRequest struct {
	ID     string `path:"id"`
	Status string `query:"status"`
}

// OrderResponse represents the API response
type OrderResponse struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	CustomerID string `json:"customer_id"`
}

func main() {
	// Initialize Gorilla Mux router
	router := mux.NewRouter()

	// Initialize roamer with Gorilla path parser
	r := roamer.NewRoamer(
		roamer.WithDecoders(decoder.NewJSON()),
		roamer.WithParsers(
			parser.NewQuery(),
			parser.NewPath(rgorilla.Path),
		),
	)

	// Define routes
	router.Handle("/orders/{id}",
		roamer.Middleware[GetOrderRequest](r)(http.HandlerFunc(handleGetOrder)),
	).Methods("GET")

	log.Println("Server starting on :8080")
	log.Println("Try: curl http://localhost:8080/orders/order-789?status=pending")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleGetOrder(w http.ResponseWriter, req *http.Request) {
	var orderReq GetOrderRequest

	if err := roamer.ParsedDataFromContext(req.Context(), &orderReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := OrderResponse{
		ID:         orderReq.ID,
		Status:     orderReq.Status,
		CustomerID: "customer-456",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
