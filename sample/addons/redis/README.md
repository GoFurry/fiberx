# redis addon

Reusable Redis helper for Fiber-template projects.

This addon stays independent from `v3/*` templates so you can copy `redis.go` into any project and wire it at the application boundary.

## Why This Is An Addon

Redis is common, but it is still optional for many projects. Keeping it as an addon lets templates stay focused while giving projects a standard, reusable Redis capability when they need it.

## Features

- thin wrapper around `github.com/redis/go-redis/v9`
- explicit config for address, username, password, database index, and pool size
- `New`, `Ping`, `Raw`, and `Close`
- common helpers for string keys, hash operations, prefix scans, and pipelines
- copy-friendly API with no dependency on `v3/*` template internals

## Files

- `redis.go`: runtime implementation
- `redis_test.go`: integration-style tests using `miniredis`
- `go.mod`: standalone module boundary for local testing

## Config

```go
cfg := redis.Config{
    Addr:     "127.0.0.1:6379",
    Username: "",
    Password: "",
    DB:       0,
    PoolSize: 10,
}
```

## Quick Start

```go
package main

import (
	"context"
	"log"

	addonredis "github.com/gofurry/fiberx/addons/redis"
)

func main() {
	service, err := addonredis.New(context.Background(), addonredis.Config{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = service.Close()
	}()

	if err := service.Set(context.Background(), "app:status", "ok"); err != nil {
		log.Fatal(err)
	}

	value, err := service.GetString(context.Background(), "app:status")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(value)
}
```

## Exposed API

- `New(ctx, cfg)`
- `(*Service).Raw()`
- `(*Service).Ping(ctx)`
- `(*Service).Close()`
- `(*Service).Set(...)`
- `(*Service).SetExpire(...)`
- `(*Service).SetNX(...)`
- `(*Service).GetString(...)`
- `(*Service).Del(...)`
- `(*Service).HSet(...)`
- `(*Service).HSetMap(...)`
- `(*Service).HGet(...)`
- `(*Service).HMGet(...)`
- `(*Service).HGetAll(...)`
- `(*Service).HDel(...)`
- `(*Service).Incr(...)`
- `(*Service).CountByPrefix(...)`
- `(*Service).FindByPrefix(...)`
- `(*Service).DelByPrefix(...)`
- `(*Service).PipelineExec(...)`

## Pairing Notes

- Fits naturally with `medium` and `heavy`
- Can also be copied into `light` projects when Redis becomes necessary later
- Pairs well with future queue, rate-limit, and cache-oriented application code

## Out Of Scope

- distributed locking helpers
- pub/sub abstraction
- stream consumer abstraction
- cluster or sentinel orchestration helpers

## Local Test

```bash
cd addons/redis
go test ./...
go vet ./...
```
