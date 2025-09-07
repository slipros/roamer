---
layout: page
title: Examples
permalink: /examples/
nav_order: 3
---

# Examples

This page contains comprehensive examples showing how to use Roamer in different scenarios.

## Table of Contents

- [Basic Usage](#basic-usage)
- [Router Integration](#router-integration)
  - [Chi Router](#chi-router)
  - [Gorilla Mux](#gorilla-mux)
  - [HttpRouter](#httprouter)
- [Content Types](#content-types)
- [Formatters](#formatters)
- [Middleware](#middleware)
- [Custom Extensions](#custom-extensions)

## Basic Usage

### Simple JSON API

```go
package main

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/slipros/roamer"
    "github.com/slipros/roamer/decoder"
    "github.com/slipros/roamer/formatter"
    "github.com/slipros/roamer/parser"
)

type CreateUserRequest struct {
    Name  string `json:"name" string:"trim_space"`
    Email string `json:"email" string:"trim_space,lower_case"`
    Age   int    `query:"age" numeric:"min=18,max=120"`
}

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
        roamer.WithFormatters(
            formatter.NewString(),
            formatter.NewNumeric(),
        ),
    )
    
    http.HandleFunc("/users", func(w http.ResponseWriter, req *http.Request) {
        var userReq CreateUserRequest
        
        if err := r.Parse(req, &userReq); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
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
    })
    
    log.Println("Server starting on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}
```

### With Default Values

```go
type SearchRequest struct {
    Query   string `query:"q"`
    Page    int    `query:"page" default:"1"`
    PerPage int    `query:"per_page" default:"20"`
    Sort    string `query:"sort" default:"relevance"`
}

r := roamer.NewRoamer(roamer.WithParsers(parser.NewQuery()))

// Example: GET /search?q=golang&page=2
// Results in: Query="golang", Page=2, PerPage=20, Sort="relevance"
```

## Router Integration

### Chi Router

```go
package main

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/slipros/roamer"
    "github.com/slipros/roamer/decoder"
    "github.com/slipros/roamer/parser"
    rchi "github.com/slipros/roamer/pkg/chi"
)

type ProductRequest struct {
    ID          string  `path:"id"`
    Name        string  `json:"name"`
    Price       float64 `json:"price"`
    Category    string  `query:"category"`
}

type ProductResponse struct {
    ID       string  `json:"id"`
    Name     string  `json:"name"`
    Price    float64 `json:"price"`
    Category string  `json:"category"`
}

func main() {
    router := chi.NewRouter()
    router.Use(middleware.Logger)
    
    roamerInstance := roamer.NewRoamer(
        roamer.WithDecoders(decoder.NewJSON()),
        roamer.WithParsers(
            parser.NewQuery(),
            parser.NewPath(rchi.NewPath(router)),
        ),
    )
    
    router.Route("/products", func(r chi.Router) {
        r.With(roamer.Middleware[ProductRequest](roamerInstance)).Post("/{id}", handleCreateProduct)
        r.With(roamer.Middleware[ProductRequest](roamerInstance)).Get("/{id}", handleGetProduct)
    })
    
    log.Println("Server starting on :8080")
    if err := http.ListenAndServe(":8080", router); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}

func handleCreateProduct(w http.ResponseWriter, r *http.Request) {
    var req ProductRequest
    
    if err := roamer.ParsedDataFromContext(r.Context(), &req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    response := ProductResponse{
        ID:       req.ID,
        Name:     req.Name,
        Price:    req.Price,
        Category: req.Category,
    }
    
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }
}

func handleGetProduct(w http.ResponseWriter, r *http.Request) {
    var req ProductRequest
    
    if err := roamer.ParsedDataFromContext(r.Context(), &req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Simulate fetching product
    response := ProductResponse{
        ID:       req.ID,
        Name:     "Sample Product",
        Price:    99.99,
        Category: req.Category,
    }
    
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }
}
```

### Gorilla Mux

```go
package main

import (
    "encoding/json"
    "net/http"

    "github.com/gorilla/mux"
    "github.com/slipros/roamer"
    "github.com/slipros/roamer/decoder"
    "github.com/slipros/roamer/parser"
    rgorilla "github.com/slipros/roamer/pkg/gorilla"
)

type OrderRequest struct {
    ID     string `path:"id"`
    Status string `query:"status"`
}

type OrderResponse struct {
    ID         string `json:"id"`
    Status     string `json:"status"`
    CustomerID string `json:"customer_id"`
}

func main() {
    router := mux.NewRouter()
    
    r := roamer.NewRoamer(
        roamer.WithDecoders(decoder.NewJSON()),
        roamer.WithParsers(
            parser.NewQuery(),
            parser.NewPath(rgorilla.Path),
        ),
    )
    
    router.Handle("/orders/{id}", 
        roamer.Middleware[OrderRequest](r)(http.HandlerFunc(handleGetOrder))).Methods("GET")
    
    log.Println("Server starting on :8080")
    if err := http.ListenAndServe(":8080", router); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}

func handleGetOrder(w http.ResponseWriter, r *http.Request) {
    var req OrderRequest
    
    if err := roamer.ParsedDataFromContext(r.Context(), &req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    response := OrderResponse{
        ID:         req.ID,
        Status:     req.Status,
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
    "net/http"

    "github.com/julienschmidt/httprouter"
    "github.com/slipros/roamer"
    "github.com/slipros/roamer/decoder"
    "github.com/slipros/roamer/parser"
    rhttprouter "github.com/slipros/roamer/pkg/httprouter"
)

type ItemRequest struct {
    ID    string  `path:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

type ItemResponse struct {
    ID    string  `json:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

func main() {
    router := httprouter.New()
    
    r := roamer.NewRoamer(
        roamer.WithDecoders(decoder.NewJSON()),
        roamer.WithParsers(
            parser.NewPath(rhttprouter.Path),
        ),
    )

    chain := func(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
        return func(next http.Handler) http.Handler {
            for i := len(middlewares) - 1; i >= 0; i-- {
                next = middlewares[i](next)
            }
            return next
        }
    }
    
    router.Handler("POST", "/items/:id", chain(
        roamer.Middleware[ItemRequest](r),
    )(http.HandlerFunc(handleCreateItem)))
    
    log.Println("Server starting on :8080")
    if err := http.ListenAndServe(":8080", router); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}

func handleCreateItem(w http.ResponseWriter, r *http.Request) {
    var req ItemRequest
    
    if err := roamer.ParsedDataFromContext(r.Context(), &req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    response := ItemResponse{
        ID:    req.ID,
        Name:  req.Name,
        Price: req.Price,
    }
    
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }
}
```

## Content Types

### XML Support

```go
type UserXMLRequest struct {
    Name    string `xml:"name"`
    Email   string `xml:"email"`
    Age     int    `xml:"age"`
    IsAdmin bool   `xml:"is_admin"`
}

r := roamer.NewRoamer(
    roamer.WithDecoders(decoder.NewXML()),
)

// POST /users with XML body:
// <?xml version="1.0"?>
// <user>
//     <name>John Doe</name>
//     <email>john@example.com</email>
//     <age>30</age>
//     <is_admin>false</is_admin>
// </user>
```

### Form URL-Encoded

```go
type ContactFormRequest struct {
    Name     string   `form:"name"`
    Email    string   `form:"email"`
    Message  string   `form:"message"`
    Topics   []string `form:"topics"`
}

r := roamer.NewRoamer(
    roamer.WithDecoders(
        decoder.NewFormURL(decoder.WithSplitSymbol(",")),
    ),
)

// POST /contact with form data:
// name=John+Doe&email=john@example.com&message=Hello&topics=support,billing
```

### Multipart Form Data with File Upload

```go
type FileUploadRequest struct {
    Title       string                 `multipart:"title"`
    Description string                 `multipart:"description"`
    File        *decoder.MultipartFile `multipart:"file"`
    AllFiles    decoder.MultipartFiles `multipart:",allfiles"`
}

r := roamer.NewRoamer(
    roamer.WithDecoders(
        decoder.NewMultipartFormData(decoder.WithMaxMemory(64 << 20)), // 64MB
    ),
)

func handleFileUpload(w http.ResponseWriter, req *http.Request) {
    var uploadReq FileUploadRequest
    
    if err := r.Parse(req, &uploadReq); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Process uploaded file
    if uploadReq.File != nil {
        fmt.Printf("Uploaded file: %s (%d bytes)\n", 
            uploadReq.File.Filename, uploadReq.File.Size)
    }
    
    // Process all files
    for _, file := range uploadReq.AllFiles {
        fmt.Printf("File: %s\n", file.Filename)
    }
}
```

## Formatters

### String Formatting

```go
type UserRequest struct {
    Name     string `json:"name" string:"trim_space,title_case"`
    Username string `json:"username" string:"trim_space,lower_case,slug"`
    Bio      string `json:"bio" string:"trim_space"`
}

r := roamer.NewRoamer(
    roamer.WithDecoders(decoder.NewJSON()),
    roamer.WithFormatters(formatter.NewString()),
)

// Input:  {"name": "  john DOE  ", "username": "  John_Doe  "}
// Output: Name="John Doe", Username="john-doe"
```

### Numeric Constraints

```go
type ProductRequest struct {
    Price    float64 `json:"price" numeric:"min=0,max=1000"`
    Quantity int     `json:"quantity" numeric:"min=1,abs"`
    Rating   float64 `json:"rating" numeric:"min=0,max=5,round"`
    Discount float32 `json:"discount" numeric:"ceil"`
}

r := roamer.NewRoamer(
    roamer.WithDecoders(decoder.NewJSON()),
    roamer.WithFormatters(formatter.NewNumeric()),
)

// Automatically applies constraints and transformations
```

### Time Formatting

```go
type EventRequest struct {
    StartTime time.Time `json:"start_time" time:"timezone=UTC,truncate=hour"`
    EndTime   time.Time `json:"end_time" time:"timezone=America/New_York"`
    Date      time.Time `query:"date" time:"start_of_day"`
    Deadline  time.Time `json:"deadline" time:"end_of_day"`
}

r := roamer.NewRoamer(
    roamer.WithDecoders(decoder.NewJSON()),
    roamer.WithParsers(parser.NewQuery()),
    roamer.WithFormatters(formatter.NewTime()),
)
```

### Slice Operations

```go
type SearchRequest struct {
    Tags       []string  `query:"tags" slice:"unique,sort"`
    Categories []string  `json:"categories" slice:"compact,limit=10"`
    Scores     []float64 `json:"scores" slice:"sort_desc,limit=5"`
    IDs        []int     `query:"ids" slice:"unique,compact"`
}

r := roamer.NewRoamer(
    roamer.WithDecoders(decoder.NewJSON()),
    roamer.WithParsers(parser.NewQuery()),
    roamer.WithFormatters(formatter.NewSlice()),
)

// GET /search?tags=golang,web,golang,api&ids=1,2,0,3
// Results: Tags=["api","golang","web"], IDs=[1,2,3]
```

## Middleware

### Type-Safe Middleware

```go
type CreateUserRequest struct {
    Name  string `json:"name" string:"trim_space"`
    Email string `json:"email" string:"trim_space,lower_case"`
    Age   int    `json:"age" numeric:"min=18"`
}

r := roamer.NewRoamer(
    roamer.WithDecoders(decoder.NewJSON()),
    roamer.WithFormatters(formatter.NewString(), formatter.NewNumeric()),
)

http.Handle("/users", 
    roamer.Middleware[CreateUserRequest](r)(http.HandlerFunc(handleCreateUser)))

func handleCreateUser(w http.ResponseWriter, req *http.Request) {
    var userReq CreateUserRequest
    
    // Data is already parsed and available in context
    if err := roamer.ParsedDataFromContext(req.Context(), &userReq); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Process the validated and formatted request...
}
```

### Custom Middleware Chain

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check authentication
        if r.Header.Get("Authorization") == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// Chain middlewares
http.Handle("/protected/users",
    authMiddleware(
        roamer.Middleware[CreateUserRequest](r)(
            http.HandlerFunc(handleCreateUser))))
```

## Custom Extensions

### Custom Parser

```go
const TagCustomHeader = "x-header"

type CustomHeaderParser struct {
    prefix string
}

func NewCustomHeaderParser(prefix string) *CustomHeaderParser {
    return &CustomHeaderParser{prefix: prefix}
}

func (p *CustomHeaderParser) Parse(r *http.Request, tag reflect.StructTag, _ parser.Cache) (any, bool) {
    tagValue, ok := tag.Lookup(TagCustomHeader)
    if !ok {
        return "", false
    }
    
    headerName := p.prefix + "-" + tagValue
    headerValue := r.Header.Get(headerName)
    return headerValue, len(headerValue) > 0
}

func (p *CustomHeaderParser) Tag() string {
    return TagCustomHeader
}

// Usage
type RequestWithCustomHeader struct {
    UserID string `x-header:"user-id"`  // Looks for X-App-user-id header
}

r := roamer.NewRoamer(
    roamer.WithParsers(NewCustomHeaderParser("X-App")),
)
```

### Custom Formatter

```go
const TagPhone = "phone"

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

func (f *PhoneFormatter) Format(tag reflect.StructTag, ptr any) error {
    tagValue, ok := tag.Lookup(TagPhone)
    if !ok {
        return nil
    }
    
    strPtr, ok := ptr.(*string)
    if !ok {
        return errors.Errorf("unsupported type: %T", ptr)
    }
    
    formatter, ok := f.formatters[tagValue]
    if !ok {
        return errors.Errorf("unknown phone formatter: %s", tagValue)
    }
    
    *strPtr = formatter(*strPtr)
    return nil
}

func (f *PhoneFormatter) Tag() string {
    return TagPhone
}

func formatToE164(phone string) string {
    digits := stripNonDigits(phone)
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
type ContactRequest struct {
    PhoneNumber string `phone:"e164"`  // Format as E.164
    RawPhone    string `phone:"strip"` // Strip non-digits
}

r := roamer.NewRoamer(
    roamer.WithFormatters(NewPhoneFormatter()),
)
```

### Complete Custom Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "regexp"
    "reflect"
    "strings"

    "github.com/slipros/roamer"
    "github.com/slipros/roamer/decoder"
    "github.com/slipros/roamer/parser"
)

// Custom parser for API version from Accept header
type APIVersionParser struct{}

func (p *APIVersionParser) Parse(r *http.Request, tag reflect.StructTag, _ parser.Cache) (any, bool) {
    accept := r.Header.Get("Accept")
    re := regexp.MustCompile(`application/vnd\.myapi\.v(\d+)\+json`)
    matches := re.FindStringSubmatch(accept)
    if len(matches) > 1 {
        return matches[1], true
    }
    return "1", true // default version
}

func (p *APIVersionParser) Tag() string {
    return "api_version"
}

// Custom formatter for cleaning and validating usernames
type UsernameFormatter struct{}

func (f *UsernameFormatter) Format(tag reflect.StructTag, ptr any) error {
    strPtr, ok := ptr.(*string)
    if !ok {
        return nil
    }
    
    // Clean username: lowercase, remove special chars, limit length
    username := strings.ToLower(*strPtr)
    username = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(username, "")
    if len(username) > 20 {
        username = username[:20]
    }
    
    *strPtr = username
    return nil
}

func (f *UsernameFormatter) Tag() string {
    return "username"
}

type UserRequest struct {
    Username   string `json:"username" username:"clean"`
    Email      string `json:"email"`
    APIVersion string `api_version:""`
}

func main() {
    r := roamer.NewRoamer(
        roamer.WithDecoders(decoder.NewJSON()),
        roamer.WithParsers(&APIVersionParser{}),
        roamer.WithFormatters(&UsernameFormatter{}),
    )
    
    http.HandleFunc("/users", func(w http.ResponseWriter, req *http.Request) {
        var userReq UserRequest
        
        if err := r.Parse(req, &userReq); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        fmt.Printf("User: %+v\n", userReq)
        
        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(map[string]string{
            "username":    userReq.Username,
            "email":       userReq.Email,
            "api_version": userReq.APIVersion,
        }); err != nil {
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

Test with:

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -H "Accept: application/vnd.myapi.v2+json" \
  -d '{"username": "John_Doe123!@#", "email": "john@example.com"}'
```

This will clean the username to "john_doe123" and set API version to "2".

## Testing Examples

### Unit Testing with Roamer

```go
func TestUserRequestParsing(t *testing.T) {
    r := roamer.NewRoamer(
        roamer.WithDecoders(decoder.NewJSON()),
        roamer.WithParsers(parser.NewQuery()),
        roamer.WithFormatters(formatter.NewString()),
    )
    
    tests := []struct {
        name     string
        body     string
        query    string
        expected CreateUserRequest
    }{
        {
            name:  "valid request",
            body:  `{"name": "  John  ", "email": "JOHN@EXAMPLE.COM"}`,
            query: "age=30",
            expected: CreateUserRequest{
                Name:  "John",
                Email: "john@example.com",
                Age:   30,
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/users?"+tt.query, 
                strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")
            
            var userReq CreateUserRequest
            err := r.Parse(req, &userReq)
            
            require.NoError(t, err)
            assert.Equal(t, tt.expected, userReq)
        })
    }
}
```

These examples cover the most common use cases for Roamer. For more advanced scenarios, check out the [API Reference](api-reference.html) and [Extending Roamer](extending.html) pages.