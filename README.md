# roamer

[![Go Report Card](https://goreportcard.com/badge/github.com/slipros/roamer)](https://goreportcard.com/report/github.com/slipros/roamer)
[![Build Status](https://github.com/slipros/roamer/actions/workflows/test.yml/badge.svg)](https://github.com/slipros/roamer/actions)
[![Coverage Status](https://coveralls.io/repos/github/slipros/roamer/badge.svg)](https://coveralls.io/github/slipros/roamer)
[![Go Reference](https://pkg.go.dev/badge/github.com/slipros/roamer.svg)](https://pkg.go.dev/github.com/slipros/roamer)
[![Go Version](https://img.shields.io/github/go-mod/go-version/slipros/roamer)](https://github.com/slipros/roamer)
[![GitHub release](https://img.shields.io/github/v/release/SLIpros/roamer.svg)](https://github.com/slipros/roamer/releases)

Roamer is a flexible, extensible HTTP request parser for Go that makes handling and extracting data from HTTP requests effortless. It provides a declarative way to map HTTP request data to Go structs using struct tags, with support for multiple data sources and content types.

```mermaid
graph TD
    subgraph "Input"
        A[HTTP Request]
    end

    subgraph "Data Sources"
        B1[Headers]
        B2[Cookies]
        B3[Query Params]
        B4[Path Variables]
        B5[Request Body]
        B6[Custom]
    end

    subgraph "Roamer Core Engine"
        direction LR
        P[Parsers]
        D[Decoders]
        F[Formatters]
    end

    subgraph "Output"
        E[Populated Go Struct]
    end

    A --> B1 & B2 & B3 & B4 & B5 & B6

    B1 & B2 & B3 & B4 & B6 -- values for --> P
    B5 -- content for --> D

    P -- parsed data --> F
    D -- decoded data --> F

    F -- formatted values --> E

    classDef source stroke:#d4ac0d,stroke-width:4px
    classDef core stroke:#0097c0,stroke-width:4px
    classDef io stroke:#333,stroke-width:4px
    class A,E io
    class B1,B2,B3,B4,B5,B6 source
    class P,D,F core
```

## Features

- **Multiple data sources**: Parse data from HTTP headers, cookies, query parameters, path variables, and custom sources (including request context)
- **Content-type based decoding**: Automatically decode request bodies based on Content-Type header
- **Default Values**: Set default values for fields using the `default` tag if no value is found in the request.
- **Formatters**: Format parsed data (e.g., trim spaces from strings, apply numeric constraints, handle time zones, manipulate slices)
- **Router integration**: Built-in support for popular routers (Chi, Gorilla Mux, HttpRouter)
- **Type conversion**: Automatic conversion of string values to appropriate Go types
- **Extensibility**: Easily create custom parsers, decoders, and formatters
- **Middleware support**: Convenient middleware for integrating with HTTP handlers
- **Performance optimizations**: Efficient reflection techniques and caching for improved performance

## Installation

```bash
go get -u github.com/slipros/roamer@latest
```

For router integrations:

```bash
# Chi router
go get -u github.com/slipros/roamer/pkg/chi@latest

# Gorilla Mux router
go get -u github.com/slipros/roamer/pkg/gorilla@latest

# HttpRouter
go get -u github.com/slipros/roamer/pkg/httprouter@latest
```

## Basic Usage

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/formatter"
	"github.com/slipros/roamer/parser"
)

// Define a request struct with appropriate tags
type CreateUserRequest struct {
	// From JSON body
	Name  string `json:"name" string:"trim_space"`
	Email string `json:"email" string:"trim_space"`
	
	// From query parameters
	Age       int       `query:"age"`
	CreatedAt time.Time `query:"created_at"`
	
	// From headers
	UserAgent string `header:"User-Agent"`
	Referer   string `header:"Referer,X-Referer"`
}

// Response struct is separate from request parsing
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	// Initialize roamer with needed components
	r := roamer.NewRoamer(
		roamer.WithDecoders(decoder.NewJSON()),
		roamer.WithParsers(
			parser.NewHeader(),
			parser.NewQuery(),
		),
		roamer.WithFormatters(
			formatter.NewString(),
			formatter.NewNumeric(),
			formatter.NewTime(),
			formatter.NewSlice(),
		),
	)
	
	// Create an HTTP handler
	http.HandleFunc("/users", func(w http.ResponseWriter, req *http.Request) {
		var userReq CreateUserRequest
		
		// Parse the request into the user struct
		if err := r.Parse(req, &userReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		// Process the request data (in a real app, save to database etc.)
		
		// Create a response
		response := UserResponse{
			ID:        "user-123",
			Name:      userReq.Name,
			Email:     userReq.Email,
			Age:       userReq.Age,
			CreatedAt: time.Now(),
		}
		
		// Return the response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})
	
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
```

## Using Middleware

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/parser"
)

// Request-specific struct
type CreateUserRequest struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `query:"age"`
	CreatedAt time.Time `query:"created_at"`
}

// Response struct (not used with roamer)
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	r := roamer.NewRoamer(
		roamer.WithDecoders(decoder.NewJSON()),
		roamer.WithParsers(parser.NewQuery()),
	)
	
	// Create an HTTP handler with middleware
	http.Handle("/users", roamer.Middleware[CreateUserRequest](r)(http.HandlerFunc(handleCreateUser)))
	
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleCreateUser(w http.ResponseWriter, req *http.Request) {
	var userReq CreateUserRequest
	
	// Get parsed data from context
	if err := roamer.ParsedDataFromContext(req.Context(), &userReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Process the request (in a real app, save to database etc.)
	
	// Create and return a response
	response := UserResponse{
		ID:        "user-123",
		Name:      userReq.Name,
		Email:     userReq.Email,
		Age:       userReq.Age,
		CreatedAt: time.Now(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
```

### Default Values

You can provide default values for fields using the `default` tag. The default value is applied only if no value is found by any parser (e.g., from a query parameter or header) and the field has its zero value.

```go
// Define a request struct with default values
type ListRequest struct {
    // Page will be 1 if the "page" query param is not provided.
	Page    int `query:"page" default:"1"`
	
	// PerPage will be 20 if the "per_page" query param is not provided.
	PerPage int `query:"per_page" default:"20"`
	
	// Sort will be "asc" if the "sort" query param is not provided.
	Sort    string `query:"sort" default:"asc"`
}

// Example usage:
// r := roamer.NewRoamer(roamer.WithParsers(parser.NewQuery()))
// req, _ := http.NewRequest("GET", "/items", nil)
// var listReq ListRequest
// r.Parse(req, &listReq) 
// listReq.Page is now 1, PerPage is 20, Sort is "asc"
```

## Router Integration Examples

### Chi Router

```go
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

// Request-specific struct for product creation
type CreateProductRequest struct {
	ID          string `path:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `query:"category"`
}

// Response struct (not used with roamer)
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
	
	// Apply middleware and define routes
	router.Route("/products", func(r chi.Router) {
		r.With(roamer.Middleware[CreateProductRequest](roamerInstance)).Post("/{id}", handleCreateProduct)
	})
	
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleCreateProduct(w http.ResponseWriter, req *http.Request) {
	var productReq CreateProductRequest
	
	if err := roamer.ParsedDataFromContext(req.Context(), &productReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Process the request (in a real app, save to database etc.)
	
	// Create and return a response
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
```

### Gorilla Mux Router

```go
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

// Request-specific struct for order retrieval
type GetOrderRequest struct {
	ID     string `path:"id"`
	Status string `query:"status"`
}

// Response struct (not used with roamer)
type OrderResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
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
	
	// Apply middleware and define routes
	router.Handle("/orders/{id}", roamer.Middleware[GetOrderRequest](r)(http.HandlerFunc(handleGetOrder))).Methods("GET")
	
	log.Println("Server starting on :8080")
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
	
	// Process the request (in a real app, fetch from database etc.)
	
	// Create and return a response
	response := OrderResponse{
		ID:        orderReq.ID,
		Status:    orderReq.Status,
		CustomerID: "customer-456",
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
```

### HttpRouter

```go
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

// Request-specific struct for item creation
type CreateItemRequest struct {
	ID    string  `path:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// Response struct (not used with roamer)
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
	
	// Process the request (in a real app, save to database etc.)
	
	// Create and return a response
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
```

## Creating an Extension for Any Router

Roamer is designed to work with any router by implementing a simple path parser adapter. Here's how to create an integration for any custom router or framework:

```go
package main

import (
	"log"
	"net/http"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/parser"
	"your/custom/router"  // Your custom router package
)

// CustomRouterPathParser adapts your custom router to work with roamer
func CustomRouterPathParser(r *router.YourRouter) parser.PathValueFunc {
	return func(req *http.Request, paramName string) (string, bool) {
		// Implement the logic to extract path parameters from your router
		// For example:
		value, ok := r.GetPathParam(req, paramName)
		return value, ok
	}
}

func main() {
	// Initialize your custom router
	customRouter := router.New()
	
	// Initialize roamer with your custom path parser
	r := roamer.NewRoamer(
		roamer.WithParsers(
			parser.NewHeader(),
			parser.NewQuery(),
			parser.NewPath(CustomRouterPathParser(customRouter)),
		),
	)
	
	// Use with your router...
}
```

This approach allows roamer to work with any router that can extract path parameters from requests, regardless of its internal implementation.

## Working with Different Content Types

### JSON

```go
// Request-specific struct for JSON data
type CreateUserRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Age     int    `json:"age"`
	IsAdmin bool   `json:"is_admin"`
}

// Initialize roamer with JSON decoder
r := roamer.NewRoamer(
	roamer.WithDecoders(decoder.NewJSON()),
)

// With custom content type
r := roamer.NewRoamer(
	roamer.WithDecoders(
		decoder.NewJSON(decoder.WithContentType[*decoder.JSON]("application/vnd.company.user+json")),
	),
)
```

### XML

```go
// Request-specific struct for XML data
type CreateUserXMLRequest struct {
	Name    string `xml:"name"`
	Email   string `xml:"email"`
	Age     int    `xml:"age"`
	IsAdmin bool   `xml:"is_admin"`
}

// Initialize roamer with XML decoder
r := roamer.NewRoamer(
	roamer.WithDecoders(decoder.NewXML()),
)
```

### Form URL-Encoded

```go
// Request-specific struct for form data
type ContactFormRequest struct {
	Name     string `form:"name"`
	Email    string `form:"email"`
	Age      int    `form:"age"`
	Comments string `form:"comments"`
}

// Initialize roamer with FormURL decoder
r := roamer.NewRoamer(
	roamer.WithDecoders(decoder.NewFormURL()),
)

// With custom split symbol for array values
r := roamer.NewRoamer(
	roamer.WithDecoders(
		decoder.NewFormURL(decoder.WithSplitSymbol(";")),
	),
)
```

### Multipart Form Data

```go
// Request-specific struct for file upload
type FileUploadRequest struct {
	Title       string                 `multipart:"title"`
	Description string                 `multipart:"description"`
	File        *decoder.MultipartFile `multipart:"file"`
	AllFiles    decoder.MultipartFiles `multipart:",allfiles"`
}

// Initialize roamer with MultipartFormData decoder
r := roamer.NewRoamer(
	roamer.WithDecoders(
		decoder.NewMultipartFormData(),
	),
)

// With custom max memory limit (default is 32MB)
r := roamer.NewRoamer(
	roamer.WithDecoders(
		decoder.NewMultipartFormData(decoder.WithMaxMemory(64 << 20)), // 64MB
	),
)
```

## Formatters

Roamer provides several built-in formatters to post-process parsed values. Formatters are applied after values are extracted from the HTTP request and converted to the appropriate type.

### String Formatter

The String formatter provides text processing operations:

```go
type UserRequest struct {
    Name     string `json:"name" string:"trim_space,title"`
    Username string `query:"username" string:"trim_space,snake"`
    Bio      string `json:"bio" string:"trim_space"`
}

// Initialize with string formatter
r := roamer.NewRoamer(
    roamer.WithDecoders(decoder.NewJSON()),
    roamer.WithParsers(parser.NewQuery()),
    roamer.WithFormatters(formatter.NewString()),
)
```

Available string operations:
- `trim_space` - Remove leading and trailing whitespace
- `lower` - Convert to lowercase 
- `upper` - Convert to uppercase
- `title` - Convert to title case (capitalize first letter of each word)
- `snake` - Convert to snake_case format
- `camel` - Convert to camelCase format
- `kebab` - Convert to kebab-case format
- `base64_encode` - Encode string to base64
- `base64_decode` - Decode base64 string
- `url_encode` - URL encode string
- `url_decode` - URL decode string
- `sanitize_html` - Basic HTML sanitization (escape < and >)
- `reverse` - Reverse string characters
- `trim_prefix=PREFIX` - Remove specified prefix from string
- `trim_suffix=SUFFIX` - Remove specified suffix from string
- `truncate=N` - Truncate string to N characters
- `replace=OLD:NEW` or `replace=OLD:NEW:COUNT` - Replace occurrences of OLD with NEW (optionally limit to COUNT replacements)
- `substring=START` or `substring=START:END` - Extract substring from START to END (or to end of string)
- `pad_left=LENGTH` or `pad_left=LENGTH:CHAR` - Pad string to LENGTH on the left (with CHAR or space)
- `pad_right=LENGTH` or `pad_right=LENGTH:CHAR` - Pad string to LENGTH on the right (with CHAR or space)

### Numeric Formatter

The Numeric formatter provides mathematical operations and constraints:

```go
type ProductRequest struct {
    Price    float64 `json:"price" numeric:"min=0,max=1000"`
    Quantity int     `json:"quantity" numeric:"min=1,abs"`
    Rating   float64 `query:"rating" numeric:"min=0,max=5,round"`
    Discount float32 `json:"discount" numeric:"ceil"`
}

// Initialize with numeric formatter
r := roamer.NewRoamer(
    roamer.WithDecoders(decoder.NewJSON()),
    roamer.WithParsers(parser.NewQuery()),
    roamer.WithFormatters(formatter.NewNumeric()),
)
```

Available numeric operations:
- `min=N` - Enforce minimum value (clamps to N if value is less)
- `max=N` - Enforce maximum value (clamps to N if value is greater)
- `abs` - Convert to absolute value
- `round` - Round to nearest integer (float types only)
- `ceil` - Round up to next integer (float types only)
- `floor` - Round down to previous integer (float types only)

### Time Formatter

The Time formatter provides time manipulation operations:

```go
type EventRequest struct {
    StartTime time.Time `json:"start_time" time:"timezone=UTC,truncate=hour"`
    EndTime   time.Time `json:"end_time" time:"timezone=America/New_York"`
    Date      time.Time `query:"date" time:"start_of_day"`
    Deadline  time.Time `json:"deadline" time:"end_of_day"`
}

// Initialize with time formatter
r := roamer.NewRoamer(
    roamer.WithDecoders(decoder.NewJSON()),
    roamer.WithParsers(parser.NewQuery()),
    roamer.WithFormatters(formatter.NewTime()),
)
```

Available time operations:
- `timezone=TZ` - Convert to specified timezone (e.g., `UTC`, `America/New_York`)
- `truncate=UNIT` - Truncate to time unit (`hour`, `minute`, `second`, or duration like `1h30m`)
- `start_of_day` - Set time to beginning of day (00:00:00)
- `end_of_day` - Set time to end of day (23:59:59.999999999)

### Slice Formatter

The Slice formatter provides operations for slice manipulation:

```go
type SearchRequest struct {
    Tags       []string  `query:"tags" slice:"unique,sort"`
    Categories []string  `json:"categories" slice:"compact,limit=10"`
    Scores     []float64 `json:"scores" slice:"sort_desc,limit=5"`
    IDs        []int     `query:"ids" slice:"unique,compact"`
}

// Initialize with slice formatter
r := roamer.NewRoamer(
    roamer.WithDecoders(decoder.NewJSON()),
    roamer.WithParsers(parser.NewQuery()),
    roamer.WithFormatters(formatter.NewSlice()),
)
```

Available slice operations:
- `unique` - Remove duplicate values
- `sort` - Sort in ascending order
- `sort_desc` - Sort in descending order
- `compact` - Remove zero values (empty strings, 0, nil, etc.)
- `limit=N` - Limit slice to first N elements

### Combining Multiple Formatters

You can use multiple formatters together in a single roamer instance:

```go
type ComprehensiveRequest struct {
    // String formatting
    Name     string `json:"name" string:"trim_space,title"`
    
    // Numeric constraints
    Age      int     `json:"age" numeric:"min=18,max=120"`
    Salary   float64 `query:"salary" numeric:"min=0,round"`
    
    // Time manipulation  
    BirthDate time.Time `json:"birth_date" time:"timezone=UTC,start_of_day"`
    
    // Slice operations
    Skills    []string `json:"skills" slice:"unique,sort" string:"trim_space,lower"`
    Scores    []int    `query:"scores" slice:"compact,sort_desc,limit=10" numeric:"min=0,max=100"`
}

r := roamer.NewRoamer(
    roamer.WithDecoders(decoder.NewJSON()),
    roamer.WithParsers(
        parser.NewHeader(),
        parser.NewQuery(),
    ),
    roamer.WithFormatters(
        formatter.NewString(),
        formatter.NewNumeric(), 
        formatter.NewTime(),
        formatter.NewSlice(),
    ),
)
```

**Note**: When multiple formatters are applied to the same field, they are processed in the order the formatters are registered with roamer.

## Architecture

The architecture is robust, modular, and designed for performance and extensibility.

1.  **Clear Separation of Concerns:** The core concepts of `Decoder` (for request bodies), `Parser` (for other request parts like headers, query, etc.), and `Formatter` (for post-processing values) create a clean and understandable system.
2.  **High Extensibility:** The interface-based design allows users to easily add support for new data formats (e.g., MessagePack), data sources (e.g., gRPC metadata), or custom formatters without modifying the core library.
3.  **Concurrency Safety:** The `Roamer` instance is thread-safe due to the use of `sync.Map` and `sync.Pool`. A single instance can be safely shared across multiple goroutines, which is essential for web server environments.
4.  **Router Independence:** The library is decoupled from any specific HTTP router. The `parser.Path` component relies on a user-provided function to extract path parameters, making it universally compatible.
5.  **Flexible Configuration:** The functional options pattern (`NewRoamer(opts ...OptionsFunc)`) provides a clean, readable, and extensible API for configuration.

## Extending Roamer

Roamer is designed to be easily extended with custom parsers, decoders, and formatters. Here are examples of how to create each type of extension.

### Creating a Custom Parser

A parser extracts data from an HTTP request based on a struct tag. Here's an example of a custom parser that extracts data from a custom HTTP header:

```go
package main

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/parser"
)

const (
	TagCustomHeader = "x-header"
)

// CustomHeaderParser parses headers with a specific prefix
type CustomHeaderParser struct {
	prefix string
}

func NewCustomHeaderParser(prefix string) *CustomHeaderParser {
	return &CustomHeaderParser{prefix: prefix}
}

// Parse implements the Parser interface
func (p *CustomHeaderParser) Parse(r *http.Request, tag reflect.StructTag, _ parser.Cache) (any, bool) {
	tagValue, ok := tag.Lookup(TagCustomHeader)
	if !ok {
		return "", false
	}
	
	// Look for header with the specified prefix
	headerName := p.prefix + "-" + tagValue
	headerValue := r.Header.Get(headerName)
	if len(headerValue) == 0 {
		return "", false
	}
	
	return headerValue, true
}

// Tag implements the Parser interface
func (p *CustomHeaderParser) Tag() string {
	return TagCustomHeader
}

// Usage
func main() {
	r := roamer.NewRoamer(
		roamer.WithParsers(NewCustomHeaderParser("X-App")),
	)
	
	// Now you can use the x-header tag in your structs:
	// type MyRequestStruct struct {
	//     UserID string `x-header:"user-id"`  // Will look for X-App-user-id header
	// }
}
```

### Creating a Custom Decoder

A decoder transforms the body of an HTTP request based on its Content-Type header. Here's an example of a custom decoder for MessagePack:

```go
package main

import (
	"net/http"

	"github.com/slipros/roamer"
	"github.com/vmihailenco/msgpack/v5" // Third-party MessagePack library
)

const (
	ContentTypeMsgPack = "application/msgpack"
)

// MsgPackDecoder decodes MessagePack format
type MsgPackDecoder struct {
	contentType string
}

func NewMsgPackDecoder() *MsgPackDecoder {
	return &MsgPackDecoder{
		contentType: ContentTypeMsgPack,
	}
}

// Decode implements the Decoder interface
func (d *MsgPackDecoder) Decode(r *http.Request, ptr any) error {
	return msgpack.NewDecoder(r.Body).Decode(ptr)
}

// ContentType implements the Decoder interface
func (d *MsgPackDecoder) ContentType() string {
	return d.contentType
}

// SetContentType allows changing the content type
func (d *MsgPackDecoder) setContentType(contentType string) {
	d.contentType = contentType
}

// WithContentType is an options function
func WithContentType(contentType string) func(*MsgPackDecoder) {
	return func(d *MsgPackDecoder) {
		d.setContentType(contentType)
	}
}

// Usage
func main() {
	r := roamer.NewRoamer(
		roamer.WithDecoders(NewMsgPackDecoder()),
	)
	
	// Now you can decode MessagePack content in your request structs
}
```

### Creating a Custom Formatter

A formatter processes parsed data before setting it on the struct field. Here's an example of a custom formatter for phone numbers:

```go
package main

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/slipros/roamer"
	rerr "github.com/slipros/roamer/err"
)

const (
	TagPhone = "phone"
)

// PhoneFormatter formats phone numbers
type PhoneFormatter struct {
	formatters map[string]func(string) string
}

func NewPhoneFormatter() *PhoneFormatter {
	return &PhoneFormatter{
		formatters: map[string]func(string) string{
			"e164": formatToE164,
			"strip": stripNonDigits,
		},
	}
}

// Format implements the Formatter interface
func (f *PhoneFormatter) Format(tag reflect.StructTag, ptr any) error {
	tagValue, ok := tag.Lookup(TagPhone)
	if !ok {
		return nil
	}
	
	strPtr, ok := ptr.(*string)
	if !ok {
		return errors.Wrapf(rerr.NotSupported, "%T", ptr)
	}
	
	formatter, ok := f.formatters[tagValue]
	if !ok {
		return errors.WithStack(rerr.FormatterNotFound{Tag: TagPhone, Formatter: tagValue})
	}
	
	*strPtr = formatter(*strPtr)
	return nil
}

// Tag implements the Formatter interface
func (f *PhoneFormatter) Tag() string {
	return TagPhone
}

// Format functions
func formatToE164(phone string) string {
	// Strip all non-digit characters
	digits := stripNonDigits(phone)
	
	// Add + prefix if not present
	if !strings.HasPrefix(digits, "+") {
		return "+" + digits
	}
	return digits
}

func stripNonDigits(phone string) string {
	re := regexp.MustCompile(`[^\d+]`)
	return re.ReplaceAllString(phone, "")
}

// Usage
func main() {
	r := roamer.NewRoamer(
		roamer.WithFormatters(NewPhoneFormatter()),
	)
	
	// Now you can use the phone tag in your structs:
	// type ContactRequest struct {
	//     PhoneNumber string `phone:"e164"`  // Format as E.164 international format
	//     RawPhone    string `phone:"strip"` // Strip all non-digit characters
	// }
}
```

## Performance Optimization

Roamer is designed with performance in mind, using efficient reflection techniques and caching where possible. For optimal performance:

- Use request structs that only include fields needed for specific endpoints
- Consider the performance implications of heavy reflection usage
- Benchmark your specific use case to identify bottlenecks

## Best Practices for Using Roamer

### Separate Request and Response Structures

Always create dedicated structs for parsing requests, separate from your response structures:

```go
// Request struct - used with roamer
type ProductRequest struct {
    Name     string  `json:"name" string:"trim_space"`
    Price    float64 `json:"price"`
    Category string  `query:"category"`
}

// Response struct - not used with roamer
type ProductResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Price     float64   `json:"price"`
    Category  string    `json:"category"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Use Request Structs Tailored to Endpoints

Create specific request structs for each endpoint to minimize parsing overhead:

```go
// Get request only needs ID and optional filters
type GetProductRequest struct {
    ID       string `path:"id"`
    Category string `query:"category"`
}

// Create request needs more fields
type CreateProductRequest struct {
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
    Category    string  `query:"category"`
}

// Update request may need ID from path and body fields
type UpdateProductRequest struct {
    ID          string  `path:"id"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
}
```

### Testing with Roamer

Here's an example of how to test an HTTP handler that uses roamer:

```go
func TestHandleCreateProduct(t *testing.T) {
    // Create a test router and roamer instance
    router := chi.NewRouter()
    r := roamer.NewRoamer(
        roamer.WithDecoders(decoder.NewJSON()),
        roamer.WithParsers(
            parser.NewHeader(),
            parser.NewQuery(),
            parser.NewPath(rchi.NewPath(router)),
        ),
    )
    
    // Create a test handler with roamer middleware
    router.With(roamer.Middleware[CreateProductRequest](r)).Post("/{id}", handleCreateProduct)
    
    // Create a test server
    ts := httptest.NewServer(router)
    defer ts.Close()
    
    // Create test request payload
    payload := `{"name":"Test Product","description":"A test product","price":29.99}`
    
    // Send test request
    resp, err := http.Post(
        ts.URL+"/products/123?category=test",
        "application/json",
        strings.NewReader(payload),
    )
    require.NoError(t, err)
    defer resp.Body.Close()
    
    // Check response
    require.Equal(t, http.StatusCreated, resp.StatusCode)
    
    // Decode response
    var product ProductResponse
    err = json.NewDecoder(resp.Body).Decode(&product)
    require.NoError(t, err)
    
    // Assert expected values
    require.Equal(t, "123", product.ID)
    require.Equal(t, "Test Product", product.Name)
    require.Equal(t, "A test product", product.Description)
    require.Equal(t, 29.99, product.Price)
    require.Equal(t, "test", product.Category)
}
```

## Complete Example

Here's a complete example that combines multiple roamer features:

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/formatter"
	"github.com/slipros/roamer/parser"
	rchi "github.com/slipros/roamer/pkg/chi"
)

type Custom string

const (
	CustomValue Custom = "value"
)

// Request-specific structs for different endpoints
type CreateProductRequest struct {
	// Body data
	Name        string  `json:"name" string:"trim_space"`
	Description string  `json:"description" string:"trim_space"`
	Price       float64 `json:"price" numeric:"min=0"`
	
	// Path parameter
	ID string `path:"id"`
	
	// Query parameters
	Category    string    `query:"category" string:"trim_space,lower"`
	Tags        []string  `query:"tags" slice:"unique,compact,sort"`
	CustomType  *Custom   `query:"custom_type"`
}

type GetProductRequest struct {
	// We only need the ID for GET requests
	ID       string `path:"id"`
	Category string `query:"category"` // Optional filter
}

// Response structure (not used with roamer)
type ProductResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
}

func main() {
	// Initialize Chi router
	router := chi.NewRouter()
	router.Use(middleware.Logger, middleware.Recoverer)
	
	// Initialize roamer
	roamerInstance := roamer.NewRoamer(
		roamer.WithDecoders(decoder.NewJSON()),
		roamer.WithParsers(
			parser.NewHeader(),
			parser.NewQuery(),
			parser.NewPath(rchi.NewPath(router)),
		),
		roamer.WithFormatters(
			formatter.NewString(),
			formatter.NewNumeric(),
			formatter.NewTime(),
			formatter.NewSlice(),
		),
	)
	
	// Define routes with appropriate request structs for each endpoint
	router.Route("/api/products", func(r chi.Router) {
		// Use CreateProductRequest for POST
		r.With(roamer.Middleware[CreateProductRequest](roamerInstance)).Post("/{id}", handleProductCreate)
		
		// Use GetProductRequest for GET - only parses what's needed
		r.With(roamer.Middleware[GetProductRequest](roamerInstance)).Get("/{id}", handleProductGet)
	})
	
	// Start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleProductCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	
	// Parse the incoming request data only
	if err := roamer.ParsedDataFromContext(r.Context(), &req); err != nil {
		http.Error(w, "Invalid product data: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// Process the request (in a real app, save to database etc.)
	log.Printf("Creating product: %s in category %s", req.Name, req.Category)
	
	// Create a response (separate from request parsing)
	response := ProductResponse{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		CreatedAt:   time.Now(),
	}
	
	// Return the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func handleProductGet(w http.ResponseWriter, r *http.Request) {
	var req GetProductRequest
	
	// Parse only the parameters needed for retrieval
	if err := roamer.ParsedDataFromContext(r.Context(), &req); err != nil {
		http.Error(w, "Invalid parameters: "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// In a real app, fetch from database using req.ID
	response := ProductResponse{
		ID:          req.ID,
		Name:        "Sample Product",
		Description: "This is a sample product description",
		Price:       99.99,
		Category:    req.Category,
		CreatedAt:   time.Now().Add(-24 * time.Hour), // Sample creation time
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
```

## FAQ

### Why use roamer instead of manually parsing HTTP requests?

Roamer offers several advantages:
1. **Declarative syntax** - Define your request structure once with tags, rather than writing repetitive parsing code
2. **Type safety** - Automatic conversion from strings to appropriate Go types
3. **Reduced boilerplate** - No need to manually extract values from different request sources
4. **Separation of concerns** - Clean separation between request parsing and business logic
5. **Extensibility** - Easy to add custom parsers, decoders, and formatters

### Can I use roamer with WebSockets or other non-HTTP protocols?

Roamer is primarily designed for HTTP requests, but its architecture is extensible. You could create custom parsers and decoders for other protocols, though you would need to adapt the interface to work with non-HTTP requests.

### How does roamer handle validation?

Roamer focuses on parsing, not validation. For validation, consider combining roamer with a validation library such as:
- [raoptimus/validator.go](https://github.com/raoptimus/validator.go)
- [go-playground/validator](https://github.com/go-playground/validator)
- [go-ozzo/ozzo-validation](https://github.com/go-ozzo/ozzo-validation)

Example:
```go
func handleCreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    
    // Parse the request
    if err := roamer.ParsedDataFromContext(r.Context(), &req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Validate the parsed data
    if err := validate.Struct(req); err != nil {
        http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
        return
    }
    
    // Process the validated request...
}
```

### How does roamer perform with high-load applications?

Roamer is designed with performance in mind, using efficient reflection techniques and caching where possible. For high-load applications, consider:

1. Using request structs that only include fields needed for specific endpoints
2. Benchmarking your specific use case to identify bottlenecks
3. Profiling memory usage and allocations in your specific context

### Is roamer production-ready in terms of performance?

Yes. A performance analysis was conducted, and the codebase is written to a very high standard with performance in mind. Key optimizations are already in place, demonstrating a strong understanding of potential bottlenecks. The overall design is efficient and minimizes unnecessary allocations. The analysis concluded that the project is in an excellent state regarding performance and memory allocations, and no changes were recommended.

### Can I use roamer with OpenAPI/Swagger specifications?

Yes, roamer works well with code generated from OpenAPI specifications. You can add roamer tags to your generated models or create dedicated request structs that map to your API specification.

## Contributing

Contributions are welcome! Feel free to submit issues or pull requests.

## License

Roamer is licensed under the MIT License. See the LICENSE file for details.

## Additional Resources

- [Go Documentation](https://pkg.go.dev/github.com/slipros/roamer)
- [GitHub Repository](https://github.com/slipros/roamer)
- [Issue Tracker](https://github.com/slipros/roamer/issues)

---

**Note:** This documentation is based on the latest version of roamer. Make sure to check the project's official documentation for the most up-to-date information.