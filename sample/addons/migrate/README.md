# migrate addon

Reusable schema migration helper for Fiber-template projects.

This addon keeps strict schema management outside the default template path. Projects can stay on `auto_migrate` for SQLite-first demos, then opt into explicit migrations when they are ready.

## Why This Is An Addon

Schema migration strategy is important, but it is not the right default for every template user. Keeping it as an addon makes the strict path available without forcing it into every template.

## Features

- thin wrapper around `github.com/pressly/goose/v3`
- explicit config for dialect, DSN, migration directory, tracking table, and verbosity
- `New`, `Close`, `Status`, `Up`, `Down`, `Create`, and `Version`
- SQL migrations only in the first version
- default tracking table `schema_migrations`

## Files

- `migrate.go`: runtime implementation
- `migrate_test.go`: SQLite integration tests
- `go.mod`: standalone module boundary for local testing

## Config

```go
cfg := migrate.Config{
    Dialect:      "sqlite",
    DSN:          "./data/app.db",
    Dir:          "./internal/db/migrations",
    Table:        "schema_migrations",
    AllowMissing: true,
    Verbose:      false,
}
```

Supported dialect values:

- `sqlite`
- `postgres`
- `mysql`

## Quick Start

```go
package main

import (
	"context"
	"log"

	addonmigrate "github.com/gofurry/fiberx/addons/migrate"
)

func main() {
	service, err := addonmigrate.New(addonmigrate.Config{
		Dialect:      "sqlite",
		DSN:          "./data/app.db",
		Dir:          "./internal/db/migrations",
		AllowMissing: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = service.Close()
	}()

	path, err := service.Create("create users table", addonmigrate.MigrationKindSQL)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("created migration:", path)

	if err := service.Up(context.Background()); err != nil {
		log.Fatal(err)
	}
}
```

## Recommended Template Pairing

- best fit: `medium`
- also a good fit: `heavy`
- still valid later in `light` projects when SQLite demo flow is no longer enough

## Typical Project Integration

- keep SQLite-first demo projects on `auto_migrate`
- move to explicit migrations when the project needs controlled schema rollout
- use a project-local migration directory such as `internal/db/migrations`
- wire migration commands into your project CLI, for example:
  - `migrate create`
  - `migrate up`
  - `migrate down`
  - `migrate status`

## Out Of Scope

- Go migrations
- model-to-migration diff generation
- automatic template integration
- AST or codegen-based schema tooling

## Local Test

```bash
cd addons/migrate
go test ./...
go vet ./...
```
