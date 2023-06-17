
[![Go Report Card](https://goreportcard.com/badge/github.com/SLIpros/roamer)](https://goreportcard.com/report/github.com/SLIpros/roamer)
[![Build Status](https://github.com/SLIpros/roamer/actions/workflows/test.yml/badge.svg)](https://github.com/SLIpros/roamer/actions)
[![Coverage Status](https://coveralls.io/repos/github/SLIpros/roamer/badge.svg?branch=main)](https://coveralls.io/github/SLIpros/roamer?branch=main)
[![Go Reference](https://pkg.go.dev/badge/github.com/SLIpros/roamer.svg)](https://pkg.go.dev/github.com/SLIpros/roamer)
[![GitHub release](https://img.shields.io/github/v/release/SLIpros/roamer.svg)](https://github.com/SLIpros/roamer/releases)

# roamer
Flexible http request parser

## Install
```go
go get -u github.com/SLIpros/roamer@latest
```

## Examples
### Default

```go
package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/SLIpros/roamer"
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
	r := roamer.NewRoamer()
	
	router := chi.NewRouter()
	router.Use(middleware.Logger, roamer.Middleware[Body](r))
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var body Body
		if err := roamer.Data(r.Context(), &body); err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		if err := json.NewEncoder(w).Encode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	http.ListenAndServe(":3000", router)
}
```
### With specified parsers

```go
package main

import (
	"github.com/SLIpros/roamer"
	"github.com/SLIpros/roamer/parser"
)

type Body struct {
	UserAgent string  `header:"User-Agent"`
	Int       int     `query:"int"`
}

func main() {
	r := roamer.NewRoamer(
		roamer.SetParsers(
			parser.Header, // parse http headers
			parser.Query, // parse http query params
		),
	)
}
```
### With specified decoders

```go
package main

import (
	"github.com/SLIpros/roamer"
	"github.com/SLIpros/roamer/decoder"
)

type Body struct {
	UserAgent string  `json:"agent"`
	Action    string  `xml:"action"`
}

func main() {
	r := roamer.NewRoamer(
		roamer.SetDecoders(
			decoder.NewJSON(), // parse json body relying on http request content-type header
			decoder.NewXML(), // parse xml body relying on http request content-type header
		),
	)
}
```

### With path parser

```go
package main

import (
	"github.com/SLIpros/roamer"
	"github.com/SLIpros/roamer/parser"
	roamerChi "github.com/SLIpros/roamer/pkg/chi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Body struct {
	Path      string  `path:"path"`
	UserAgent string  `header:"User-Agent"`
	Int       int     `query:"int"`
}

func main() {
	router := chi.NewRouter()
	
	r := roamer.NewRoamer(
		roamer.SetParsers(
			parser.Header, // parse http headers
			parser.Query, // parse http query params
			parser.NewPath(roamerChi.NewPath(router)), // parse http path params
		),
	)

	router.Use(middleware.Logger, roamer.Middleware[Body](r))
	router.Post("/test/{path}", func(w http.ResponseWriter, r *http.Request) {
		var body Body
		if err := roamer.Data(r.Context(), &body); err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		if err := json.NewEncoder(w).Encode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	http.ListenAndServe(":3000", router)
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

	"github.com/SLIpros/roamer"
	"github.com/SLIpros/roamer/parser"
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

func ParserProfile(r *http.Request, tag reflect.StructTag, _ parser.Cache) (string, any, bool) {
	tagValue, ok := tag.Lookup(TagProfile)
	if !ok {
		return "", nil, false
	}

	profile, ok := r.Context().Value(ContextKeyProfile).(*Profile)
	if !ok {
		return "", nil, false
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
		return "", nil, false
	}

	return TagProfile, v, true
}

type Body struct {
	ClientID   *uuid.UUID `profile:"client_id"`
	Age        int        `profile:"age"`
	ProfilePtr *Profile   `profile:"profile"`
	Profile    Profile    `profile:"profile"`
}

func main() {
	r := roamer.NewRoamer(
		roamer.SetParsers(
			ParserProfile, // parse profile
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
		if err := roamer.Data(r.Context(), &body); err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		if err := json.NewEncoder(w).Encode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	http.ListenAndServe(":3000", router)
}
```
