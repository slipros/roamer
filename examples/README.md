# Roamer Examples

This directory contains practical examples of using the Roamer HTTP request parsing library.

## Basic Examples

- [**basic**](basic/) - Simple usage examples to get started
  - `cmd/simple/main.go` - Basic request parsing
  - `cmd/middleware/main.go` - Using roamer as middleware

## Router Integration Examples

- [**chi_router**](chi_router/) - Integration with Chi router
- [**gorilla_router**](gorilla_router/) - Integration with Gorilla Mux
- [**httprouter**](httprouter/) - Integration with HttpRouter

## Content Type Examples

- [**json**](json/) - Working with JSON requests
- [**xml**](xml/) - Working with XML requests
- [**form**](form/) - Working with URL-encoded forms
- [**multipart**](multipart/) - Working with multipart form data (file uploads)

## Formatter Examples

- [**formatters**](formatters/) - Using built-in formatters
  - String formatting (trim, case conversion, etc.)
  - Numeric constraints (min, max, rounding)
  - Time manipulation (timezone, truncation)
  - Slice operations (unique, sort, limit)

## Advanced Examples

- [**custom_parser**](custom_parser/) - Creating custom parsers for extracting data from request context
  - Implementing the `roamer.Parser` interface
  - Extracting member/user data injected by middleware
  - Type conversion support (string, UUID, struct, pointer)
  - Real-world pattern for authentication and authorization data

- [**custom_decoder**](custom_decoder/) - Creating custom decoders for YAML content type
  - Implementing the `roamer.Decoder` interface
  - Adding support for custom content types beyond JSON/XML/Form
  - Processing complex nested YAML structures

- [**custom_formatter**](custom_formatter/) - Creating custom formatters for phone numbers
  - Implementing the `roamer.Formatter` interface
  - Domain-specific data transformations (E.164 phone format)
  - Processing formatter tag operations

- [**body_preservation**](body_preservation/) - Reading request body multiple times
  - Using `roamer.WithPreserveBody()` option
  - Logging and validating bodies before parsing
  - Understanding when body preservation is necessary

## Running Examples

Each example directory contains a complete, runnable Go application. To run an example:

```bash
# Simple parsing example
cd examples/basic/cmd/simple
go run main.go

# Router integration example
cd examples/chi_router
go run main.go

# Formatters example
cd examples/formatters
go run main.go
```

Then in another terminal, test with curl (see each example's output for specific commands).
