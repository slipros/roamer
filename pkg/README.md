# Roamer Extensions

This directory contains integration packages for popular HTTP routers, making it easy to use Roamer with your preferred routing framework.

## Available Extensions

Roamer provides built-in support for the following HTTP routers:

| Router | Package | Description |
|--------|---------|-------------|
| [Chi](https://github.com/go-chi/chi) | [pkg/chi](https://github.com/slipros/roamer/tree/main/pkg/chi) | Path parameter adapter for the Chi router |
| [Gorilla Mux](https://github.com/gorilla/mux) | [pkg/gorilla](https://github.com/slipros/roamer/tree/main/pkg/gorilla) | Path parameter adapter for the Gorilla Mux router |
| [HttpRouter](https://github.com/julienschmidt/httprouter) | [pkg/httprouter](https://github.com/slipros/roamer/tree/main/pkg/httprouter) | Path parameter adapter for the HttpRouter router |

## What These Extensions Do

Each extension provides a path parameter parser adapter that allows Roamer to extract path parameters from HTTP requests using the router's native path parameter extraction mechanism. This enables you to use the `path` tag in your structs to parse URL path parameters.

## Usage

1. Import the appropriate extension package for your router
2. Create a path parser using the extension's adapter function
3. Register the path parser with Roamer
4. Use the `path` tag in your structs to extract path parameters

For detailed examples, please refer to the README in each extension directory.

## Creating Custom Extensions

If you're using a router that's not listed above, you can easily create your own extension by implementing a function that conforms to the `parser.PathValueFunc` type:

```go
// PathValueFunc returns path variable value with name from http request.
type PathValueFunc = func(r *http.Request, name string) (string, bool)
```

Then, use this function with the `parser.NewPath` constructor:

```go
parser.NewPath(myCustomPathParserFunc)
```

See the existing extensions for implementation examples.
