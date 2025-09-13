---
layout: page
title: API Reference
permalink: /api-reference/
nav_order: 4
---

# API Reference

Complete documentation for all Roamer components.

## Table of Contents

- [Core API](#core-api)
- [Parsers](#parsers)
- [Decoders](#decoders)
- [Formatters](#formatters)
- [Middleware](#middleware)
- [Struct Tags](#struct-tags)

## Core API

### Roamer

The main Roamer struct that orchestrates parsing, decoding, and formatting.

#### Constructor

```go
func NewRoamer(opts ...OptionsFunc) *Roamer
```

Create a new Roamer instance with the specified options.

**Options:**
- `WithParsers(parsers ...Parser)` - Add parsers for different request parts
- `WithDecoders(decoders ...Decoder)` - Add decoders for request bodies
- `WithFormatters(formatters ...Formatter)` - Add formatters for post-processing

#### Methods

```go
func (r *Roamer) Parse(req *http.Request, ptr interface{}) error
```

Parse an HTTP request into the provided struct pointer.

**Parameters:**
- `req` - HTTP request to parse
- `ptr` - Pointer to struct that will receive parsed data

**Returns:** Error if parsing fails

### Context Functions

```go
func ParsedDataFromContext(ctx context.Context, ptr interface{}) error
```

Extract parsed data from request context (used with middleware).

## Parsers

Parsers extract data from different parts of HTTP requests.

### Query Parser

Parse query parameters from the URL.

```go
func NewQuery(opts ...QueryOption) *Query
```

**Options:**
- `WithSplitSymbol(symbol string)` - Set symbol for splitting array values (default: ",")
- `WithDisabledSplit()` - Disable automatic splitting of values

**Struct Tag:** `query:"parameter_name"`

**Example:**
```go
type Request struct {
    Page    int      `query:"page"`
    Tags    []string `query:"tags"`  // Splits "tag1,tag2,tag3"
    Search  string   `query:"q"`
}
```

### Header Parser

Parse HTTP headers.

```go
func NewHeader() *Header
```

**Struct Tag:** `header:"Header-Name"` or `header:"Header1,Header2"` (fallback)

**Example:**
```go
type Request struct {
    UserAgent string `header:"User-Agent"`
    Language  string `header:"Accept-Language,Accept-Lang"`
    Token     string `header:"Authorization"`
}
```

### Cookie Parser

Parse HTTP cookies.

```go
func NewCookie() *Cookie
```

**Struct Tag:** `cookie:"cookie_name"`

**Example:**
```go
type Request struct {
    SessionID string `cookie:"session_id"`
    Theme     string `cookie:"theme"`
    UserID    string `cookie:"user_id"`
}
```

### Path Parser

Parse path parameters (requires router integration).

```go
func NewPath(pathFunc PathValueFunc) *Path
```

**PathValueFunc signature:**
```go
type PathValueFunc func(req *http.Request, paramName string) (string, bool)
```

**Struct Tag:** `path:"parameter_name"`

**Example:**
```go
type Request struct {
    ID       string `path:"id"`
    Category string `path:"category"`
}
```

**Router Integrations:**
- Chi: `parser.NewPath(rchi.NewPath(router))`
- Gorilla: `parser.NewPath(rgorilla.Path)`
- HttpRouter: `parser.NewPath(rhttprouter.Path)`

## Decoders

Decoders handle request body parsing based on Content-Type.

### JSON Decoder

Parse JSON request bodies.

```go
func NewJSON(opts ...JSONOption) *JSON
```

**Options:**
- `WithContentType[T](contentType string)` - Custom content type (default: "application/json")

**Struct Tag:** `json:"field_name"`

### XML Decoder

Parse XML request bodies.

```go
func NewXML(opts ...XMLOption) *XML
```

**Options:**
- `WithContentType[T](contentType string)` - Custom content type (default: "application/xml")

**Struct Tag:** `xml:"field_name"`

### Form URL-Encoded Decoder

Parse form-encoded request bodies.

```go
func NewFormURL(opts ...FormURLOption) *FormURL
```

**Options:**
- `WithContentType[T](contentType string)` - Custom content type
- `WithSplitSymbol(symbol string)` - Set symbol for array splitting

**Struct Tag:** `form:"field_name"`

### Multipart Form Data Decoder

Parse multipart form data (including file uploads).

```go
func NewMultipartFormData(opts ...MultipartFormDataOption) *MultipartFormData
```

**Options:**
- `WithContentType[T](contentType string)` - Custom content type
- `WithMaxMemory(maxMemory int64)` - Max memory for file parsing (default: 32MB)

**Struct Tag:** `multipart:"field_name"`

**Special Types:**
- `*MultipartFile` - Single uploaded file
- `MultipartFiles` - Multiple uploaded files
- `multipart:",allfiles"` - Get all uploaded files

**Example:**
```go
type FileUploadRequest struct {
    Title    string                 `multipart:"title"`
    File     *decoder.MultipartFile `multipart:"file"`
    AllFiles decoder.MultipartFiles `multipart:",allfiles"`
}
```

## Formatters

Formatters post-process parsed values.

### String Formatter

Format string values.

```go
func NewString() *String
```

**Struct Tag:** `string:"operation1,operation2"`

**Operations:**
- `trim_space` - Remove leading and trailing whitespace
- `lower` - Convert to lowercase
- `upper` - Convert to uppercase
- `slug` - Convert to URL-friendly slug

**Example:**
```go
type Request struct {
    Name     string `json:"name" string:"trim_space,title_case"`
    Username string `json:"username" string:"trim_space,lower,slug"`
}
```

### Numeric Formatter

Apply constraints and transformations to numeric values.

```go
func NewNumeric() *Numeric
```

**Struct Tag:** `numeric:"constraint1,constraint2"`

**Constraints:**
- `min=N` - Enforce minimum value
- `max=N` - Enforce maximum value
- `abs` - Convert to absolute value
- `round` - Round to nearest integer (floats only)
- `ceil` - Round up (floats only)
- `floor` - Round down (floats only)

**Example:**
```go
type Request struct {
    Price    float64 `json:"price" numeric:"min=0,max=1000"`
    Quantity int     `json:"quantity" numeric:"min=1,abs"`
    Rating   float64 `json:"rating" numeric:"min=0,max=5,round"`
}
```

### Time Formatter

Format and manipulate time values.

```go
func NewTime() *Time
```

**Struct Tag:** `time:"operation1,operation2"`

**Operations:**
- `timezone=TZ` - Convert to specified timezone (e.g., `UTC`, `America/New_York`)
- `truncate=UNIT` - Truncate to unit (`hour`, `minute`, `second`, or duration)
- `start_of_day` - Set to beginning of day (00:00:00)
- `end_of_day` - Set to end of day (23:59:59.999999999)

**Example:**
```go
type Request struct {
    StartTime time.Time `json:"start_time" time:"timezone=UTC,truncate=hour"`
    Date      time.Time `query:"date" time:"start_of_day"`
    Deadline  time.Time `json:"deadline" time:"end_of_day"`
}
```

### Slice Formatter

Format and manipulate slice values.

```go
func NewSlice() *Slice
```

**Struct Tag:** `slice:"operation1,operation2"`

**Operations:**
- `unique` - Remove duplicate values
- `sort` - Sort in ascending order
- `sort_desc` - Sort in descending order
- `compact` - Remove zero/empty values
- `limit=N` - Limit to first N elements

**Example:**
```go
type Request struct {
    Tags       []string  `query:"tags" slice:"unique,sort"`
    Categories []string  `json:"categories" slice:"compact,limit=10"`
    Scores     []float64 `json:"scores" slice:"sort_desc,limit=5"`
}
```

## Middleware

### Type-Safe Middleware

```go
func Middleware[T any](roamer *Roamer) func(http.Handler) http.Handler
```

Create middleware that parses requests and stores results in context.

**Usage:**
```go
http.Handle("/endpoint", 
    roamer.Middleware[RequestStruct](r)(http.HandlerFunc(handler)))
```

**Accessing parsed data:**
```go
func handler(w http.ResponseWriter, req *http.Request) {
    var data RequestStruct
    if err := roamer.ParsedDataFromContext(req.Context(), &data); err != nil {
        // Handle error
    }
    // Use parsed data
}
```

## Struct Tags

### Common Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `json:"field"` | Parse from JSON body | `json:"name"` |
| `query:"param"` | Parse from query parameter | `query:"page"` |
| `header:"Header-Name"` | Parse from HTTP header | `header:"User-Agent"` |
| `cookie:"name"` | Parse from cookie | `cookie:"session_id"` |
| `path:"param"` | Parse from path variable | `path:"id"` |
| `form:"field"` | Parse from form data | `form:"email"` |
| `multipart:"field"` | Parse from multipart data | `multipart:"file"` |
| `xml:"field"` | Parse from XML body | `xml:"username"` |

### Default Values

```go
type Request struct {
    Page    int    `query:"page" default:"1"`
    PerPage int    `query:"per_page" default:"20"`
    Sort    string `query:"sort" default:"asc"`
}
```

The `default` tag provides fallback values when no data is found.

### Multiple Formatters

You can apply multiple formatters to the same field:

```go
type Request struct {
    Tags []string `query:"tags" slice:"unique,sort" string:"trim_space,lower"`
}
```

Formatters are applied in the order they're registered with Roamer.

## Error Handling

### Common Errors

- `NotSupported` - Type conversion not supported
- `FormatterNotFound` - Unknown formatter requested
- Standard parsing errors for malformed data

### Error Context

Errors are wrapped with context information to help with debugging:

```go
if err := r.Parse(req, &data); err != nil {
    log.Printf("Parsing failed: %+v", err) // Includes stack trace
    http.Error(w, "Invalid request", http.StatusBadRequest)
    return
}
```

## Type Support

### Supported Types

Roamer supports automatic conversion to these Go types:

**Basic Types:**
- `string`
- `bool`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`

**Time:**
- `time.Time` - Supports multiple formats (RFC3339, RFC1123, "2006-01-02", etc.)

**Slices:**
- `[]string`, `[]int`, `[]float64`, etc.
- Automatic splitting on comma (configurable)

**Pointers:**
- All basic types as pointers (`*string`, `*int`, etc.)

**Custom Types:**
- Types with custom `UnmarshalText` methods
- Enums and custom string types

### Type Conversion Examples

```go
type Request struct {
    // String to int
    Age int `query:"age"`
    
    // String to bool (accepts: true/false, 1/0, yes/no)
    Active bool `query:"active"`
    
    // String to time.Time
    CreatedAt time.Time `query:"created_at"`
    
    // Comma-separated to slice
    Tags []string `query:"tags"`
    
    // Optional values
    OptionalEmail *string `json:"email"`
}
```

## Performance Considerations

### Best Practices

1. **Use specific request structs** - Only include fields needed for each endpoint
2. **Cache roamer instances** - Reuse instances across requests for better performance
3. **Consider reflection overhead** - Roamer uses reflection for struct analysis

### Memory Usage

- Roamer uses `sync.Pool` for object reuse where possible
- Path parser is the only component that uses caching
- Memory allocations are minimized in parsing hot paths

### Benchmarking

For performance-critical applications, benchmark your specific use cases:

```go
func BenchmarkRoamerParse(b *testing.B) {
    r := roamer.NewRoamer(/* configure */)
    req := /* create test request */
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var data RequestStruct
        if err := r.Parse(req, &data); err != nil {
            b.Fatal(err)
        }
    }
}
```