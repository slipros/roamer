package main

import (
	"encoding/xml"
	"log"
	"net/http"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
)

// BookRequest demonstrates XML parsing
type BookRequest struct {
	XMLName xml.Name `xml:"book"`
	Title   string   `xml:"title"`
	Author  string   `xml:"author"`
	Year    int      `xml:"year"`
	ISBN    string   `xml:"isbn"`
}

// BookResponse represents the API response
type BookResponse struct {
	XMLName xml.Name `xml:"book"`
	ID      string   `xml:"id"`
	Title   string   `xml:"title"`
	Author  string   `xml:"author"`
	Year    int      `xml:"year"`
	ISBN    string   `xml:"isbn"`
}

func main() {
	// Initialize roamer with XML decoder
	r := roamer.NewRoamer(
		roamer.WithDecoders(decoder.NewXML()),
	)

	http.HandleFunc("/books", func(w http.ResponseWriter, req *http.Request) {
		var bookReq BookRequest

		if err := r.Parse(req, &bookReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response := BookResponse{
			ID:     "book-456",
			Title:  bookReq.Title,
			Author: bookReq.Author,
			Year:   bookReq.Year,
			ISBN:   bookReq.ISBN,
		}

		w.Header().Set("Content-Type", "application/xml")
		xml.NewEncoder(w).Encode(response)
	})

	log.Println("Server starting on :8080")
	log.Println("Try: curl -X POST http://localhost:8080/books \\")
	log.Println("  -H 'Content-Type: application/xml' \\")
	log.Println("  -d '<book><title>Go Programming</title><author>John Doe</author><year>2024</year><isbn>978-1234567890</isbn></book>'")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
