package decoder_test

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
)

// readCloser wraps a bytes.Buffer to implement io.ReadCloser
type readCloser struct {
	*bytes.Buffer
}

func (rc *readCloser) Close() error {
	return nil
}

// ExampleNewJSON demonstrates how to create and use a JSON decoder.
func ExampleNewJSON() {
	// Define a structure for JSON data
	type User struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// Create a JSON decoder
	jsonDecoder := decoder.NewJSON()

	// Create a roamer instance with the JSON decoder
	r := roamer.NewRoamer(
		roamer.WithDecoders(jsonDecoder),
	)

	// Create a request with JSON body
	jsonBody := `{"id": 123, "name": "John Doe", "email": "john@example.com"}`
	req := &http.Request{
		Method: "POST",
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		Body:          &readCloser{bytes.NewBufferString(jsonBody)},
		ContentLength: int64(len(jsonBody)),
	}

	// Parse the request
	var user User
	err := r.Parse(req, &user)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("User ID: %d\n", user.ID)
	fmt.Printf("User Name: %s\n", user.Name)
	fmt.Printf("User Email: %s\n", user.Email)

	// Output:
	// User ID: 123
	// User Name: John Doe
	// User Email: john@example.com
}

// ExampleJSON_ContentType demonstrates retrieving the content type handled by a JSON decoder.
func ExampleJSON_ContentType() {
	d := decoder.NewJSON()
	fmt.Printf("Content-Type: %s\n", d.ContentType())
	fmt.Printf("Struct Tag: %s\n", d.Tag())

	// Output:
	// Content-Type: application/json
	// Struct Tag: json
}

// ExampleJSON_Decode demonstrates direct use of the JSON decoder's Decode method.
func ExampleJSON_Decode() {
	// Define a structure
	type Product struct {
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}

	// Create decoder
	jsonDecoder := decoder.NewJSON()

	// Create request with JSON body
	jsonBody := `{"name": "Laptop", "price": 999.99}`
	req := &http.Request{
		Method: "POST",
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
		Body:          &readCloser{bytes.NewBufferString(jsonBody)},
		ContentLength: int64(len(jsonBody)),
	}

	// Decode directly
	var product Product
	err := jsonDecoder.Decode(req, &product)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Product: %s, Price: $%.2f\n", product.Name, product.Price)

	// Output:
	// Product: Laptop, Price: $999.99
}
