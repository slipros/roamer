# Basic Examples

This directory contains basic usage examples for roamer.

## Simple Parsing

Location: `cmd/simple/main.go`

Demonstrates basic request parsing with multiple data sources.

```bash
cd cmd/simple
go run main.go
```

Test with:
```bash
curl -X POST http://localhost:8080/users?role=admin \
  -H 'Content-Type: application/json' \
  -d '{"name":" John Doe ","email":"JOHN@EXAMPLE.COM"}'
```

## Middleware Usage

Location: `cmd/middleware/main.go`

Shows how to use roamer as HTTP middleware.

```bash
cd cmd/middleware
go run main.go
```

Test with:
```bash
curl -X POST http://localhost:8080/products?category=electronics \
  -H 'Content-Type: application/json' \
  -d '{"name":"Laptop","price":999.99}'
```
