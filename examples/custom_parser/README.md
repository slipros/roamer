# Custom Parser Example

This example demonstrates how to create a custom parser for Roamer that extracts member data from the request context.

## Overview

The custom `MemberParser` extracts member information that was previously injected into the request context by middleware. This pattern is useful when you need to parse data from sources other than standard HTTP request components (headers, query parameters, body, etc.).

## Key Components

### MemberParser

The custom parser implements the `roamer.Parser` interface:

```go
type MemberParser struct{}

func (p *MemberParser) Parse(r *http.Request, tag reflect.StructTag, _ parser.Cache) (any, bool)
func (p *MemberParser) Tag() string
```

It extracts member data from the request context based on struct tags:

- `member:"id"` - Extracts the member ID (as string or UUID)
- `member:"organization_id"` - Extracts the organization ID
- `member:"member"` - Extracts the entire Member struct

### Member Middleware

The middleware injects a `Member` object into the request context:

```go
func Member(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        next.ServeHTTP(w, r.WithContext(model.ContextWithMember(r.Context(), &model.Member{
            ID:             uuid.Must(uuid.NewV7()),
            OrganizationID: uuid.Must(uuid.NewV7()),
            Name:           "Jack",
            Role:           "Admin",
        })))
    }
}
```

## Usage

Define a request struct with `member` tags:

```go
type AppRequest struct {
    MemberID       string        `member:"id"`
    MemberIDAsUUID uuid.UUID     `member:"id"`
    Member         *model.Member `member:"member"`
    MemberAsStruct model.Member  `member:"member"`
    Action         string        `query:"action"`
}
```

Initialize Roamer with the custom parser:

```go
r := roamer.NewRoamer(
    roamer.WithParsers(
        NewMemberParser(),
        parser.NewQuery(),
    ),
)
```

## Running the Example

```bash
# From the custom_parser directory
go run main.go
```

## Testing

```bash
# Test the endpoint
curl -X GET 'http://localhost:8080/api/action?action=view'

# Expected output:
# Successfully parsed member data from context:
# MemberID: <generated UUID>
# MemberID (UUID): <generated UUID>
# Member Pointer: &{ID:<uuid> OrganizationID:<uuid> Name:Jack Role:Admin}
# Member Struct: {ID:<uuid> OrganizationID:<uuid> Name:Jack Role:Admin}
# Action: view
```

## Implementation Details

1. **Context Storage**: Member data is stored in the request context using a type-safe key
2. **Type Conversion**: The parser supports automatic conversion to different types (string, UUID, struct)
3. **Middleware Pattern**: The middleware wraps the handler and injects data before the custom parser extracts it

This pattern is particularly useful for:
- Authentication/authorization data
- Request-scoped configuration
- Tenant/organization context
- User session information
