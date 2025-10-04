---
layout: page
title: Getting Started
permalink: /getting-started/
nav_order: 2
---

# Getting Started with Roamer

This guide will help you get up and running with Roamer quickly.

## Installation

Install Roamer using Go modules:

```bash
go get -u github.com/slipros/roamer@latest
```

For router integrations, install the specific packages you need:

```bash
# Chi router
go get -u github.com/slipros/roamer/pkg/chi@latest

# Gorilla Mux router  
go get -u github.com/slipros/roamer/pkg/gorilla@latest

# HttpRouter
go get -u github.com/slipros/roamer/pkg/httprouter@latest
```

## Basic Concepts

Roamer works with three main components:

1. **Parsers** - Extract data from different parts of HTTP requests (headers, query params, cookies, path)
2. **Decoders** - Handle request body decoding based on Content-Type  
3. **Formatters** - Post-process parsed values (trim strings, format numbers, etc.)

## Your First Roamer Application

Let's create a simple API endpoint that handles user creation:

### Step 1: Define Your Request Structure

```go
type CreateUserRequest struct {
    // From JSON body
    Name  string `json:"name" string:"trim_space"`
    Email string `json:"email" string:"trim_space,lower"`
    Age   int    `json:"age" numeric:"min=0,max=120"`
    
    // From query parameters
    Source string `query:"source" default:"web"`
    
    // From headers
    UserAgent string `header:"User-Agent"`
    ContentType string `header:"Content-Type"`
}
```

### Step 2: Initialize Roamer

```go
r := roamer.NewRoamer(
    roamer.WithDecoders(decoder.NewJSON()),
    roamer.WithParsers(
        parser.NewHeader(),
        parser.NewQuery(),
    ),
    roamer.WithFormatters(
        formatter.NewString(),
        formatter.NewNumeric(),
    ),
)
```

### Step 3: Create Your Handler

```go
func handleCreateUser(w http.ResponseWriter, req *http.Request) {
    var userReq CreateUserRequest
    
    // Parse the request
    if err := r.Parse(req, &userReq); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Process the request
    fmt.Printf("Creating user: %+v\n", userReq)
    
    // Send response
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(map[string]any{
        "status": "created",
        "name": userReq.Name,
        "email": userReq.Email,
        "source": userReq.Source,
    }); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }
}
```

### Step 4: Complete Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "github.com/slipros/roamer"
    "github.com/slipros/roamer/decoder"
    "github.com/slipros/roamer/formatter"
    "github.com/slipros/roamer/parser"
)

type CreateUserRequest struct {
    Name      string `json:"name" string:"trim_space"`
    Email     string `json:"email" string:"trim_space,lower"`
    Age       int    `json:"age" numeric:"min=0,max=120"`
    Source    string `query:"source" default:"web"`
    UserAgent string `header:"User-Agent"`
}

func main() {
    // Initialize roamer
    r := roamer.NewRoamer(
        roamer.WithDecoders(decoder.NewJSON()),
        roamer.WithParsers(
            parser.NewHeader(),
            parser.NewQuery(),
        ),
        roamer.WithFormatters(
            formatter.NewString(),
            formatter.NewNumeric(),
        ),
    )
    
    // Create handler
    http.HandleFunc("/users", func(w http.ResponseWriter, req *http.Request) {
        var userReq CreateUserRequest
        
        if err := r.Parse(req, &userReq); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        fmt.Printf("Creating user: %+v\n", userReq)
        
        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(map[string]any{
            "status": "created",
            "name": userReq.Name,
            "email": userReq.Email,
            "source": userReq.Source,
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

## Testing Your Application

Test your endpoint with curl:

```bash
curl -X POST http://localhost:8080/users?source=mobile \
  -H "Content-Type: application/json" \
  -H "User-Agent: MyApp/1.0" \
  -d '{
    "name": "  John Doe  ",
    "email": "JOHN@EXAMPLE.COM",
    "age": 30
  }'
```

The parsed data will have:
- `Name`: "John Doe" (spaces trimmed)
- `Email`: "john@example.com" (trimmed and lowercased)  
- `Age`: 30 (validated between 0-120)
- `Source`: "mobile" (from query param)
- `UserAgent`: "MyApp/1.0" (from header)

## Using Middleware

For cleaner code, you can use Roamer's middleware:

```go
// Initialize roamer
r := roamer.NewRoamer(/* ... */)

// Use middleware
http.Handle("/users", roamer.Middleware[CreateUserRequest](r)(http.HandlerFunc(handleCreateUser)))

func handleCreateUser(w http.ResponseWriter, req *http.Request) {
    var userReq CreateUserRequest
    
    // Get parsed data from context
    if err := roamer.ParsedDataFromContext(req.Context(), &userReq); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Process the request...
}
```

## Common Struct Tags

Here are the most commonly used struct tags in Roamer:

| Tag | Purpose | Example |
|-----|---------|---------|
| `json:"field"` | Parse from JSON body | `json:"name"` |
| `query:"param"` | Parse from query parameter | `query:"page"` |
| `header:"Header-Name"` | Parse from HTTP header | `header:"User-Agent"` |
| `cookie:"name"` | Parse from cookie | `cookie:"session_id"` |
| `path:"param"` | Parse from path variable | `path:"id"` |
| `default:"value"` | Default value if not found | `default:"1"` |
| `string:"operation"` | String formatting | `string:"trim_space,lower"` |
| `numeric:"constraint"` | Numeric constraints | `numeric:"min=0,max=100"` |

## Next Steps

Now that you have Roamer working, explore more advanced features:

- [**Examples**](examples.html) - Router integration, different content types, and complex use cases
- [**API Reference**](api-reference.html) - Complete documentation of all parsers, decoders, and formatters
- [**Extending Roamer**](extending.html) - Create custom components for your specific needs

## Troubleshooting

### Common Issues

**Parsing fails silently**
- Make sure your struct tags match the expected format
- Check that you've registered the appropriate parsers and decoders

**Type conversion errors**
- Verify the data format matches the Go type
- Use string formatters to clean data before type conversion

**Missing values**
- Check tag names match request parameter names exactly
- Use default values for optional fields

**Performance issues**
- Use request-specific structs instead of large generic ones
- Consider caching roamer instances for reuse

**Body cannot be read multiple times**
- Enable `WithPreserveBody()` option if you need multiple reads
- Be aware of memory implications for large request bodies