---
layout: page
title: Extending Roamer
permalink: /extending/
nav_order: 5
---

# Extending Roamer

Roamer is designed to be easily extended with custom parsers, decoders, and formatters. This guide shows you how to create each type of extension.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Creating Custom Parsers](#creating-custom-parsers)
- [Creating Custom Decoders](#creating-custom-decoders)
- [Creating Custom Formatters](#creating-custom-formatters)
- [Integration Examples](#integration-examples)
- [Best Practices](#best-practices)

## Architecture Overview

Roamer's extensible architecture is built around three main interfaces:

- **Parser** - Extract data from HTTP request parts (headers, query, cookies, etc.)
- **Decoder** - Parse request body content based on Content-Type
- **Formatter** - Post-process parsed values before setting them on struct fields

Each component works independently and can be combined with built-in or other custom components.

## Creating Custom Parsers

A parser extracts data from an HTTP request based on struct tags.

### Parser Interface

```go
type Parser interface {
    Parse(req *http.Request, tag reflect.StructTag, cache Cache) (any, bool)
    Tag() string
}
```

### Custom Header Parser Example

Let's create a parser that extracts headers with a specific prefix:

```go
package main

import (
    "net/http"
    "reflect"
    "strings"
    
    "github.com/slipros/roamer"
    "github.com/slipros/roamer/parser"
)

const TagCustomHeader = "x-header"

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

// Usage example
type APIRequest struct {
    UserID    string `x-header:"user-id"`    // Looks for X-App-user-id
    TenantID  string `x-header:"tenant-id"`  // Looks for X-App-tenant-id
    RequestID string `x-header:"request-id"` // Looks for X-App-request-id
}

func main() {
    r := roamer.NewRoamer(
        roamer.WithParsers(NewCustomHeaderParser("X-App")),
    )
    
    // Now you can use the x-header tag in your structs
    http.HandleFunc("/api", func(w http.ResponseWriter, req *http.Request) {
        var apiReq APIRequest
        
        if err := r.Parse(req, &apiReq); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        // Process request with extracted headers
        w.WriteHeader(http.StatusOK)
    })
}
```

### Environment Variable Parser

Here's a more complex example that parses environment variables:

```go
package main

import (
    "net/http"
    "os"
    "reflect"
    
    "github.com/slipros/roamer"
    "github.com/slipros/roamer/parser"
)

const TagEnv = "env"

type EnvParser struct{}

func NewEnvParser() *EnvParser {
    return &EnvParser{}
}

func (p *EnvParser) Parse(req *http.Request, tag reflect.StructTag, _ parser.Cache) (any, bool) {
    envVar, ok := tag.Lookup(TagEnv)
    if !ok {
        return "", false
    }
    
    value := os.Getenv(envVar)
    return value, len(value) > 0
}

func (p *EnvParser) Tag() string {
    return TagEnv
}

type ConfigRequest struct {
    DatabaseURL string `env:"DATABASE_URL"`
    APIKey      string `env:"API_KEY"`
    Debug       string `env:"DEBUG"`
}
```

## Creating Custom Decoders

A decoder parses request body content based on the Content-Type header.

### Decoder Interface

```go
type Decoder interface {
    Decode(req *http.Request, ptr any) error
    ContentType() string
}
```

### MessagePack Decoder Example

Let's create a decoder for MessagePack format:

```go
package main

import (
    "net/http"
    
    "github.com/slipros/roamer"
    "github.com/vmihailenco/msgpack/v5" // Third-party MessagePack library
)

const ContentTypeMsgPack = "application/msgpack"

type MsgPackDecoder struct {
    contentType string
}

func NewMsgPackDecoder(opts ...MsgPackOption) *MsgPackDecoder {
    d := &MsgPackDecoder{
        contentType: ContentTypeMsgPack,
    }
    
    for _, opt := range opts {
        opt(d)
    }
    
    return d
}

// Decode implements the Decoder interface
func (d *MsgPackDecoder) Decode(r *http.Request, ptr any) error {
    return msgpack.NewDecoder(r.Body).Decode(ptr)
}

// ContentType implements the Decoder interface
func (d *MsgPackDecoder) ContentType() string {
    return d.contentType
}

// Option pattern for configuration
type MsgPackOption func(*MsgPackDecoder)

func WithContentType(contentType string) MsgPackOption {
    return func(d *MsgPackDecoder) {
        d.contentType = contentType
    }
}

// Usage
func main() {
    r := roamer.NewRoamer(
        roamer.WithDecoders(
            NewMsgPackDecoder(),
            // Or with custom content type
            NewMsgPackDecoder(WithContentType("application/x-msgpack")),
        ),
    )
    
    // Now you can decode MessagePack content in your requests
}
```

### YAML Decoder Example

```go
package main

import (
    "net/http"
    
    "github.com/slipros/roamer"
    "gopkg.in/yaml.v3"
)

const ContentTypeYAML = "application/yaml"

type YAMLDecoder struct {
    contentType string
}

func NewYAMLDecoder() *YAMLDecoder {
    return &YAMLDecoder{
        contentType: ContentTypeYAML,
    }
}

func (d *YAMLDecoder) Decode(r *http.Request, ptr any) error {
    return yaml.NewDecoder(r.Body).Decode(ptr)
}

func (d *YAMLDecoder) ContentType() string {
    return d.contentType
}

type YAMLRequest struct {
    Name        string            `yaml:"name"`
    Version     string            `yaml:"version"`
    Dependencies map[string]string `yaml:"dependencies"`
}
```

## Creating Custom Formatters

A formatter post-processes parsed values before they're set on struct fields.

### Formatter Interface

```go
type Formatter interface {
    Format(tag reflect.StructTag, ptr any) error
    Tag() string
}
```

### Phone Number Formatter Example

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

const TagPhone = "phone"

type PhoneFormatter struct {
    formatters map[string]func(string) string
}

func NewPhoneFormatter() *PhoneFormatter {
    return &PhoneFormatter{
        formatters: map[string]func(string) string{
            "e164":        formatToE164,
            "strip":       stripNonDigits,
            "us_format":   formatUSPhone,
            "international": formatInternational,
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
        return errors.Wrapf(rerr.NotSupported, "phone formatter only supports *string, got %T", ptr)
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

// Formatting functions
func formatToE164(phone string) string {
    digits := stripNonDigits(phone)
    if !strings.HasPrefix(digits, "+") {
        // Assume US number if no country code
        if len(digits) == 10 {
            return "+1" + digits
        }
        return "+" + digits
    }
    return digits
}

func stripNonDigits(phone string) string {
    re := regexp.MustCompile(`[^\d+]`)
    return re.ReplaceAllString(phone, "")
}

func formatUSPhone(phone string) string {
    digits := stripNonDigits(phone)
    if len(digits) == 10 {
        return fmt.Sprintf("(%s) %s-%s", digits[0:3], digits[3:6], digits[6:10])
    }
    return phone
}

func formatInternational(phone string) string {
    digits := stripNonDigits(phone)
    if strings.HasPrefix(digits, "+") {
        return digits
    }
    return "+" + digits
}

// Usage example
type ContactRequest struct {
    HomePhone   string `json:"home_phone" phone:"us_format"`
    MobilePhone string `json:"mobile_phone" phone:"e164"`
    WorkPhone   string `json:"work_phone" phone:"strip"`
}

func main() {
    r := roamer.NewRoamer(
        roamer.WithFormatters(NewPhoneFormatter()),
    )
    
    // Phone numbers will be automatically formatted
}
```

### Address Formatter Example

```go
package main

import (
    "reflect"
    "strings"
    "unicode"
    
    "github.com/slipros/roamer"
)

const TagAddress = "address"

type AddressFormatter struct{}

func NewAddressFormatter() *AddressFormatter {
    return &AddressFormatter{}
}

func (f *AddressFormatter) Format(tag reflect.StructTag, ptr any) error {
    operation, ok := tag.Lookup(TagAddress)
    if !ok {
        return nil
    }
    
    strPtr, ok := ptr.(*string)
    if !ok {
        return nil // Skip non-string fields
    }
    
    switch operation {
    case "normalize":
        *strPtr = normalizeAddress(*strPtr)
    case "upper":
        *strPtr = strings.ToUpper(*strPtr)
    case "title":
        *strPtr = strings.Title(strings.ToLower(*strPtr))
    }
    
    return nil
}

func (f *AddressFormatter) Tag() string {
    return TagAddress
}

func normalizeAddress(addr string) string {
    // Clean up extra whitespace
    addr = strings.TrimSpace(addr)
    addr = regexp.MustCompile(`\s+`).ReplaceAllString(addr, " ")
    
    // Common abbreviations
    replacements := map[string]string{
        " St ":     " Street ",
        " Ave ":    " Avenue ",
        " Blvd ":   " Boulevard ",
        " Dr ":     " Drive ",
        " Rd ":     " Road ",
        " Ct ":     " Court ",
        " Ln ":     " Lane ",
    }
    
    for old, new := range replacements {
        addr = strings.ReplaceAll(addr, old, new)
    }
    
    return addr
}

type AddressRequest struct {
    StreetAddress string `json:"street" address:"normalize"`
    City          string `json:"city" address:"title"`
    State         string `json:"state" address:"upper"`
}
```

## Integration Examples

### Complete Custom Extension

Here's a complete example showing how to create a comprehensive extension:

```go
package main

import (
    "encoding/csv"
    "fmt"
    "net/http"
    "reflect"
    "strconv"
    "strings"
    
    "github.com/slipros/roamer"
    "github.com/slipros/roamer/parser"
)

// Custom CSV Parser
const TagCSV = "csv"

type CSVParser struct{}

func NewCSVParser() *CSVParser {
    return &CSVParser{}
}

func (p *CSVParser) Parse(r *http.Request, tag reflect.StructTag, _ parser.Cache) (any, bool) {
    paramName, ok := tag.Lookup(TagCSV)
    if !ok {
        return nil, false
    }
    
    csvData := r.URL.Query().Get(paramName)
    if csvData == "" {
        return nil, false
    }
    
    reader := csv.NewReader(strings.NewReader(csvData))
    records, err := reader.ReadAll()
    if err != nil {
        return nil, false
    }
    
    // Flatten all records into a single slice
    var result []string
    for _, record := range records {
        result = append(result, record...)
    }
    
    return result, true
}

func (p *CSVParser) Tag() string {
    return TagCSV
}

// Custom CSV Decoder
const ContentTypeCSV = "text/csv"

type CSVDecoder struct{}

func NewCSVDecoder() *CSVDecoder {
    return &CSVDecoder{}
}

func (d *CSVDecoder) Decode(r *http.Request, ptr any) error {
    // Assume ptr is a slice of structs for CSV rows
    reader := csv.NewReader(r.Body)
    records, err := reader.ReadAll()
    if err != nil {
        return err
    }
    
    // This is a simplified example - in reality you'd use reflection
    // to populate the struct slice based on CSV headers
    fmt.Printf("CSV records: %+v\n", records)
    return nil
}

func (d *CSVDecoder) ContentType() string {
    return ContentTypeCSV
}

// Custom Validation Formatter
const TagValidate = "validate"

type ValidationFormatter struct{}

func NewValidationFormatter() *ValidationFormatter {
    return &ValidationFormatter{}
}

func (f *ValidationFormatter) Format(tag reflect.StructTag, ptr any) error {
    validation, ok := tag.Lookup(TagValidate)
    if !ok {
        return nil
    }
    
    switch validation {
    case "email":
        return validateEmail(ptr)
    case "positive":
        return validatePositive(ptr)
    case "non_empty":
        return validateNonEmpty(ptr)
    }
    
    return nil
}

func (f *ValidationFormatter) Tag() string {
    return TagValidate
}

func validateEmail(ptr any) error {
    strPtr, ok := ptr.(*string)
    if !ok {
        return nil
    }
    
    if !strings.Contains(*strPtr, "@") {
        return fmt.Errorf("invalid email format")
    }
    return nil
}

func validatePositive(ptr any) error {
    switch v := ptr.(type) {
    case *int:
        if *v < 0 {
            *v = 0 // or return error
        }
    case *float64:
        if *v < 0 {
            *v = 0 // or return error
        }
    }
    return nil
}

func validateNonEmpty(ptr any) error {
    strPtr, ok := ptr.(*string)
    if !ok {
        return nil
    }
    
    if strings.TrimSpace(*strPtr) == "" {
        return fmt.Errorf("field cannot be empty")
    }
    return nil
}

// Usage example
type ComplexRequest struct {
    // CSV data from query parameter
    Tags []string `csv:"tags"`
    
    // Validated fields
    Email  string  `json:"email" validate:"email"`
    Amount float64 `json:"amount" validate:"positive"`
    Name   string  `json:"name" validate:"non_empty"`
}

func main() {
    r := roamer.NewRoamer(
        roamer.WithParsers(NewCSVParser()),
        roamer.WithDecoders(NewCSVDecoder()),
        roamer.WithFormatters(NewValidationFormatter()),
    )
    
    http.HandleFunc("/complex", func(w http.ResponseWriter, req *http.Request) {
        var complexReq ComplexRequest
        
        if err := r.Parse(req, &complexReq); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        fmt.Printf("Parsed request: %+v\n", complexReq)
        w.WriteHeader(http.StatusOK)
    })
    
    log.Println("Server starting on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}
```

### Router-Specific Extension

Create a custom path parser for a hypothetical router:

```go
package main

import (
    "net/http"
    
    "github.com/slipros/roamer"
    "github.com/slipros/roamer/parser"
    "your/custom/router" // Your custom router
)

// CustomRouterPathParser adapts your router to work with roamer
func CustomRouterPathParser(r *router.YourRouter) parser.PathValueFunc {
    return func(req *http.Request, paramName string) (string, bool) {
        // Implement extraction logic for your router
        value, exists := r.GetPathParam(req, paramName)
        return value, exists
    }
}

func main() {
    customRouter := router.New()
    
    r := roamer.NewRoamer(
        roamer.WithParsers(
            parser.NewQuery(),
            parser.NewPath(CustomRouterPathParser(customRouter)),
        ),
    )
    
    // Use with your router...
}
```

## Best Practices

### Parser Best Practices

1. **Handle missing tags gracefully** - Return `(nil, false)` if tag is not found
2. **Type conversion** - Let Roamer handle type conversion, return appropriate types
3. **Error handling** - Return `(nil, false)` for parsing errors rather than panicking
4. **Performance** - Cache expensive operations when possible

```go
func (p *MyParser) Parse(r *http.Request, tag reflect.StructTag, cache parser.Cache) (any, bool) {
    tagValue, ok := tag.Lookup(p.Tag())
    if !ok {
        return nil, false // Tag not found
    }
    
    // Use cache for expensive operations
    if cached, exists := cache.Get("key"); exists {
        return cached, true
    }
    
    value := extractValue(r, tagValue)
    if value == "" {
        return nil, false // Value not found
    }
    
    cache.Set("key", value) // Cache result
    return value, true
}
```

### Decoder Best Practices

1. **Content-Type matching** - Be specific about content types you handle
2. **Error handling** - Return descriptive errors for parsing failures
3. **Stream handling** - Don't load entire body into memory for large requests
4. **Security** - Validate input to prevent attacks

```go
func (d *MyDecoder) Decode(r *http.Request, ptr any) error {
    // Limit request size to prevent DoS
    r.Body = http.MaxBytesReader(nil, r.Body, 1<<20) // 1MB limit
    
    // Use streaming decoder when possible
    decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields() // Security: reject unknown fields
    
    return decoder.Decode(ptr)
}
```

### Formatter Best Practices

1. **Type safety** - Check types before formatting
2. **Idempotency** - Formatting should be safe to apply multiple times
3. **Error handling** - Return clear errors for unsupported operations
4. **Performance** - Avoid expensive operations in formatters

```go
func (f *MyFormatter) Format(tag reflect.StructTag, ptr any) error {
    operation, ok := tag.Lookup(f.Tag())
    if !ok {
        return nil // No formatting needed
    }
    
    // Type check first
    strPtr, ok := ptr.(*string)
    if !ok {
        return fmt.Errorf("formatter %s only supports *string, got %T", f.Tag(), ptr)
    }
    
    // Apply formatting
    *strPtr = f.transform(*strPtr, operation)
    return nil
}
```

### Testing Custom Components

Always test your custom components thoroughly:

```go
func TestCustomParser(t *testing.T) {
    parser := NewCustomParser()
    
    tests := []struct {
        name     string
        request  *http.Request
        tag      string
        expected any
        found    bool
    }{
        {
            name:     "valid tag",
            request:  createTestRequest(),
            tag:      `custom:"test"`,
            expected: "expected_value",
            found:    true,
        },
        {
            name:     "missing tag",
            request:  createTestRequest(),
            tag:      `other:"test"`,
            expected: nil,
            found:    false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tag := reflect.StructTag(tt.tag)
            result, found := parser.Parse(tt.request, tag, nil)
            
            assert.Equal(t, tt.expected, result)
            assert.Equal(t, tt.found, found)
        })
    }
}
```

### Integration Testing

Test your extensions work with Roamer:

```go
func TestCustomExtensionIntegration(t *testing.T) {
    r := roamer.NewRoamer(
        roamer.WithParsers(NewCustomParser()),
        roamer.WithDecoders(NewCustomDecoder()),
        roamer.WithFormatters(NewCustomFormatter()),
    )
    
    type TestRequest struct {
        CustomField string `custom:"field" custom_format:"operation"`
    }
    
    req := createTestRequest()
    var testReq TestRequest
    
    err := r.Parse(req, &testReq)
    require.NoError(t, err)
    
    assert.Equal(t, "expected_formatted_value", testReq.CustomField)
}
```

By following these patterns and best practices, you can create powerful extensions that integrate seamlessly with Roamer's architecture.