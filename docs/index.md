---
layout: home
title: Home
---

# Roamer - HTTP Request Parser for Go

[![Go Report Card](https://goreportcard.com/badge/github.com/slipros/roamer)](https://goreportcard.com/report/github.com/slipros/roamer)
[![Build Status](https://github.com/slipros/roamer/actions/workflows/test.yml/badge.svg)](https://github.com/slipros/roamer/actions)
[![Coverage Status](https://coveralls.io/repos/github/slipros/roamer/badge.svg)](https://coveralls.io/github/slipros/roamer)
[![Go Reference](https://pkg.go.dev/badge/github.com/slipros/roamer.svg)](https://pkg.go.dev/github.com/slipros/roamer)
[![GitHub release](https://img.shields.io/github/v/release/SLIpros/roamer.svg)](https://github.com/slipros/roamer/releases)

Roamer is a flexible, extensible HTTP request parser for Go that makes handling and extracting data from HTTP requests effortless. It provides a declarative way to map HTTP request data to Go structs using struct tags, with support for multiple data sources and content types.

## Key Features

- **Multiple data sources**: Parse data from HTTP headers, cookies, query parameters, and path variables
- **Content-type based decoding**: Automatically decode request bodies based on Content-Type header  
- **Default Values**: Set default values for fields using the `default` tag if no value is found in the request
- **Formatters**: Format parsed data (trim spaces, apply numeric constraints, handle time zones, manipulate slices)
- **Router integration**: Built-in support for popular routers (Chi, Gorilla Mux, HttpRouter)
- **Type conversion**: Automatic conversion of string values to appropriate Go types
- **Extensibility**: Easily create custom parsers, decoders, and formatters
- **Middleware support**: Convenient middleware for integrating with HTTP handlers
- **Performance optimizations**: Efficient reflection techniques and caching

## Quick Start

### Installation

```bash
go get -u github.com/slipros/roamer@latest
```

### Basic Example

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
    // From JSON body
    Name  string `json:"name" string:"trim_space"`
    Email string `json:"email" string:"trim_space"`
    
    // From query parameters
    Age       int       `query:"age"`
    CreatedAt time.Time `query:"created_at"`
    
    // From headers
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
            formatter.NewTime(),
        ),
    )
    
    // Create HTTP handler
    http.HandleFunc("/users", func(w http.ResponseWriter, req *http.Request) {
        var userReq CreateUserRequest
        
        if err := r.Parse(req, &userReq); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        // Process the parsed request...
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "status": "created",
            "name": userReq.Name,
        })
    })
    
    http.ListenAndServe(":8080", nil)
}
```

## Architecture Overview

```mermaid
graph TD
    subgraph "Input"
        A[HTTP Request]
    end

    subgraph "1. Data Sources"
        B1[Headers]
        B2[Cookies]
        B3[Query Params]
        B4[Path Variables]
        B5[Request Body]
    end

    subgraph "2. Roamer Core Engine"
        direction LR
        P[Parsers]
        D[Decoders]
        F[Formatters]
    end

    subgraph "Output"
        E[Populated Go Struct]
    end

    A --> B1 & B2 & B3 & B4 & B5

    B1 & B2 & B3 & B4 -- values for --> P
    B5 -- content for --> D

    P -- parsed data --> F
    D -- decoded data --> F

    F -- formatted values --> E

    classDef source fill:#fef9e7,stroke:#d4ac0d,stroke-width:2px
    classDef core fill:#d4f1f9,stroke:#0097c0,stroke-width:2px
    classDef io fill:#f5f5f5,stroke:#333,stroke-width:2px
    class A,E io
    class B1,B2,B3,B4,B5 source
    class P,D,F core
```

## Getting Started

Ready to get started? Check out our detailed guides:

- [**Getting Started**](getting-started.html) - Installation and basic usage
- [**Examples**](examples.html) - Comprehensive examples for different use cases  
- [**API Reference**](api-reference.html) - Complete API documentation
- [**Extending Roamer**](extending.html) - Create custom parsers, decoders, and formatters

## Community & Support

- **GitHub Repository**: [github.com/slipros/roamer](https://github.com/slipros/roamer)
- **Issues & Bug Reports**: [GitHub Issues](https://github.com/slipros/roamer/issues)
- **Go Package Documentation**: [pkg.go.dev](https://pkg.go.dev/github.com/slipros/roamer)

## License

Roamer is licensed under the [MIT License](https://github.com/slipros/roamer/blob/main/LICENSE).