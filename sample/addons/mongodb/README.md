# mongodb addon

Reusable MongoDB helper for Fiber-template projects.

This addon stays independent from `v3/*` templates so you can copy `mongodb.go` into any project and wire it at the application boundary.

## Features

- Official `go.mongodb.org/mongo-driver/v2`
- `URI`-first configuration with structured host/auth fallback
- Thin service wrapper with `Client`, `Database`, `Collection`, `Ping`, and `Close`
- Collection-oriented CRUD helpers for common operations
- Full access to the raw official driver for sessions, transactions, aggregation, indexes, and advanced queries

## Files

- `mongodb.go`: runtime implementation, designed to be copied as a single file
- `mongodb_test.go`: unit tests with mocks, no real MongoDB dependency
- `go.mod`: standalone module boundary for local testing

## Quick Start

```go
package main

import (
	"context"
	"log"

	addonmongo "github.com/gofurry/fiberx/addons/mongodb"
)

type User struct {
	Name string `bson:"name"`
}

func main() {
	service, err := addonmongo.New(context.Background(), addonmongo.Config{
		Hosts:    []string{"127.0.0.1:27017"},
		Database: "awesome_app",
		AppName:  "fiberx",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = service.Close(context.Background())
	}()

	users := service.Collection("users")
	if _, err := users.InsertOne(context.Background(), User{Name: "Alice"}); err != nil {
		log.Fatal(err)
	}

	var result User
	if err := users.FindOne(context.Background(), map[string]any{"name": "Alice"}, &result); err != nil {
		log.Fatal(err)
	}
}
```

## Configuration

`Config` supports two modes:

- `URI` non-empty: start from the MongoDB URI and then apply addon-level options such as pool sizing, app name, retry flags, and TLS.
- `URI` empty: build a MongoDB URI from `Hosts`, `Username`, `Password`, `AuthSource`, `Database`, `ReplicaSet`, `Direct`, timeouts, retry flags, and TLS.

Useful fields:

- `Database`: default database used by `Database()` and `Collection(name)`
- `ConnectTimeout`: connect timeout during client setup
- `ServerSelectionTimeout`: server selection timeout during topology discovery
- `SocketTimeout`: included in the generated URI when using structured config
- `RetryReads` / `RetryWrites`: optional explicit retry flags
- `TLS.Enabled`: enables TLS
- `TLS.InsecureSkipVerify`: disables server certificate verification for development/testing only

## API Notes

- `Service.Client()` exposes the raw `*mongo.Client`.
- `Service.Database(name...)` returns the raw `*mongo.Database`.
- `Service.Collection(name)` returns an addon wrapper around the default database.
- `Collection.Raw()` exposes the raw `*mongo.Collection`.
- `FindOne` expects a non-nil pointer target and calls official `Decode(...)`.
- `FindMany` expects a slice pointer target and calls cursor `All(...)`.
- If you need transactions or sessions, use `Service.Client()` or `Collection.Raw()` directly with the official driver.

## Local Test

```bash
cd addons/mongodb
go test ./...
go vet ./...
```
