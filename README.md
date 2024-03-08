[![Go Report Card](https://goreportcard.com/badge/github.com/slipros/roamer)](https://goreportcard.com/report/github.com/slipros/roamer)
[![Build Status](https://github.com/slipros/roamer/actions/workflows/test.yml/badge.svg)](https://github.com/slipros/roamer/actions)
[![Coverage Status](https://coveralls.io/repos/github/SLIpros/roamer/badge.svg?branch=main)](https://coveralls.io/github/SLIpros/roamer?branch=main)
[![Go Reference](https://pkg.go.dev/badge/github.com/slipros/roamer.svg)](https://pkg.go.dev/github.com/slipros/roamer)
[![GitHub release](https://img.shields.io/github/v/release/SLIpros/roamer.svg)](https://github.com/slipros/roamer/releases)

# roamer
Flexible http request parser

## Install
```go
go get -u github.com/slipros/roamer@latest
```

## Decoder

Decode body of http request based on `Content-Type` header.

| Type      | Content-Type                      |
|-----------|-----------------------------------|
| json      | application/json                  |
| xml       | application/xml                   |
| form      | application/x-www-form-urlencoded |
| multipart | multipart/form-data               |
| `custom`  | `any`                             |

### Json decoder with custom content type

```go
package main

import (
	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
)

func main() {
	_ = roamer.NewRoamer(
		roamer.WithDecoders(
			decoder.NewJSON(decoder.WithContentType[*decoder.JSON]("my content type")),
		),
	)
}
```

## Parser
Parsing data from source.

| Type     | Source      |
|----------|-------------|
| header   | http header |
| query    | http query  |
| path     | router path |
| `custom` | `any`       |

## Examples
```
curl --location 'http://127.0.0.1:3000?int=1&int8=2&int16=3&int32=4&int64=5&time=2021-01-01T02%3A07%3A14Z&custom_type=value' \
--header 'Content-Type: application/json' \
--data-raw '{
    "string": "Hello",
    "email": "test@test.com"
}'
```

```go
package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/slipros/roamer/parser"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Custom string

const (
	CustomValue Custom = "value"
)

type Body struct {
	String string  `json:"string"`
	Email  *string `json:"email"`

	Int        int       `query:"int"`
	Int8       int8      `query:"int8"`
	Int16      int16     `query:"int16"`
	Int32      int32     `query:"int32"`
	Int64      int64     `query:"int64"`
	Time       time.Time `query:"time"`
	CustomType *Custom   `query:"custom_type"`
}

func main() {
	r := roamer.NewRoamer(
		roamer.WithDecoders(decoder.NewJSON()),
		roamer.WithParsers(parser.NewQuery()),
	)

	router := chi.NewRouter()
	router.Use(middleware.Logger, roamer.Middleware[Body](r))
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var body Body
		if err := roamer.ParsedDataFromContext(r.Context(), &body); err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		if err := json.NewEncoder(w).Encode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	
	if err := http.ListenAndServe(":3000", router); err != nil {
		panic(err)
	}
}
```
### With path parser
```
curl --location --request POST 'http://127.0.0.1:3000/test/some_value?int=1' \
--header 'User-Agent: PostmanRuntime/7.33.0'
```

```go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/parser"
	rchi "github.com/slipros/roamer/pkg/chi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Body struct {
	Path      string `path:"path"` // after parse value will be = some_value
	UserAgent string `header:"User-Agent"` // after parse value will be = PostmanRuntime/7.33.0
	Int       int    `query:"int"` // after parse value will be = 1
}

func main() {
	router := chi.NewRouter()

	r := roamer.NewRoamer(
		roamer.WithParsers(
			parser.NewHeader(),                        // parse http headers
			parser.NewQuery(),                         // parse http query params
			parser.NewPath(rchi.NewPath(router)), // parse http path params
		),
	)

	router.Use(middleware.Logger, roamer.Middleware[Body](r))
	router.Post("/test/{path}", func(w http.ResponseWriter, r *http.Request) {
		var body Body
		if err := roamer.ParsedDataFromContext(r.Context(), &body); err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		if err := json.NewEncoder(w).Encode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	
	if err := http.ListenAndServe(":3000", router); err != nil {
		panic(err)
	}
}
```

### With custom parser

```go
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/slipros/roamer"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gofrs/uuid"
)

type ContextKey string

const (
	ContextKeyProfile ContextKey = "profile"
)

type Profile struct {
	Age      int
	Email    string
	ClientID uuid.UUID
}

const (
	TagProfile = "profile"
)

type ProfileParser struct{}

func (p *ProfileParser) Parse(r *http.Request, tag reflect.StructTag, _ roamer.Cache) (any, bool) {
	tagValue, ok := tag.Lookup(TagProfile)
	if !ok {
		return nil, false
	}

	profile, ok := r.Context().Value(ContextKeyProfile).(*Profile)
	if !ok {
		return nil, false
	}

	var v any
	switch tagValue {
	case "client_id":
		v = profile.ClientID
	case "email":
		v = profile.Email
	case "age":
		v = &profile.Age
	case "profile":
		v = profile
	default:
		return nil, false
	}

	return v, true
}

func (p *ProfileParser) Tag() string {
	return TagProfile
}

type Body struct {
	ClientID   *uuid.UUID `profile:"client_id"`
	Age        int        `profile:"age"`
	ProfilePtr *Profile   `profile:"profile"`
	Profile    Profile    `profile:"profile"`
}

func main() {
	r := roamer.NewRoamer(
		roamer.WithParsers(
			&ProfileParser{}, // parse profile from context
		),
	)

	router := chi.NewRouter()

	profileMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			profile := Profile{
				Email:    "profile@profile.com",
				ClientID: uuid.FromStringOrNil("e4aa78cd-a98a-4d9e-84ee-fea61c1c047b"),
				Age:      100,
			}

			ctxWithProfile := context.WithValue(r.Context(), ContextKeyProfile, &profile)
			next.ServeHTTP(w, r.WithContext(ctxWithProfile))
		})
	}

	router.Use(middleware.Logger, profileMiddleware, roamer.Middleware[Body](r))
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var body Body
		if err := roamer.ParsedDataFromContext(r.Context(), &body); err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		if err := json.NewEncoder(w).Encode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	
	if err := http.ListenAndServe(":3000", router); err != nil {
		panic(err)
	}
}
```

### With multipart/form-data decoder
```
curl --location 'http://127.0.0.1:3000' \
--header 'X-Referer: http://localhost:3000' \
--header 'Authorization: Bearer 018ad70c-6c98-789c-ac0e-b8e51931e628' \
--form 'campaignId="campaign"' \
--form 'fileId="1337"' \
--form 'file=@"/C:/Users/slipros/Downloads/devices.csv"' \
--form 'file2=@"/C:/Users/slipros/Downloads/devices.csv"'
```

```go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/slipros/roamer"
	"github.com/slipros/roamer/decoder"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type UploadDevicesFile struct {
	CampaignID string                 `multipart:"campaignId"` // after parse multipart/form-data key campaignId = campaign
	FileID     int                    `multipart:"fileId"`     // parse multipart/form-data key fileId = 1337
	File       *decoder.MultipartFile `multipart:"file"`       // parse multipart/form-data key file
	Files      decoder.MultipartFiles `multipart:",allfiles"`  // parse all multipart/form-data files = [file, file2]
}

func main() {
	r := roamer.NewRoamer(
		roamer.WithDecoders(
			decoder.NewMultipartFormData(),
		),
	)

	router := chi.NewRouter()
	router.Use(middleware.Logger, roamer.Middleware[UploadDevicesFile](r))
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var body UploadDevicesFile
		if err := roamer.ParsedDataFromContext(r.Context(), &body); err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		if err := json.NewEncoder(w).Encode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	
	if err := http.ListenAndServe(":3000", router); err != nil {
		panic(err)
    }
}
```

## Experimental

### FastStructFieldParser

Significantly reduces the number of heap memory allocations.

```go
package main

import (
	"github.com/slipros/roamer"
)

r := NewRoamer(
	WithParsers(parser.NewHeader(), parser.NewQuery()), 
	WithExperimentalFastStructFieldParser(), // enables experimental fast struct field parser
)
```

```text
goos: windows
goarch: amd64
pkg: github.com/slipros/roamer
cpu: 12th Gen Intel(R) Core(TM) i9-12900K
BenchmarkParse_With_Body_Header_Query
BenchmarkParse_With_Body_Header_Query-16                                 4182058
               279.5 ns/op            64 B/op          8 allocs/op
BenchmarkParse_With_Body_Header_Query_FastStructFieldParser
BenchmarkParse_With_Body_Header_Query_FastStructFieldParser-16           5059383
               241.2 ns/op             0 B/op          0 allocs/op
PASS
```