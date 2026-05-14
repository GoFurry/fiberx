# fiberx

![License](https://img.shields.io/badge/License-MIT-6C757D?style=flat&color=3B82F6)
![Release](https://img.shields.io/github/v/release/gofurry/fiberx?style=flat&color=blue)
![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go&logoColor=white)
[![Go Report Card](https://goreportcard.com/badge/github.com/gofurry/fiberx)](https://goreportcard.com/report/github.com/gofurry/fiberx)

![Weekend Project](https://img.shields.io/badge/weekend-project-8B5CF6?style=flat)
![Made with Love](https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange)

[中文说明](./README_zh.md)

`fiberx` is a CLI-first Fiber project generator repository.

The repository is intentionally focused on the generator mainline itself: assets, planning rules, validation, rendering, upgrade inspection, build automation, and regression coverage.

## Release

- `v0.1.0`: completed
- `v0.1.1`: completed
- `v0.1.2`: completed
- `v0.1.3`: completed
- `v0.1.4`: completed
- `v0.1.5`: in progress

## Docs

- [Docs index](./docs/README.md)
- [Usage guide](./docs/guides/usage.md)
- [Release process](./docs/guides/release-process.md)
- [Build hook safety](./docs/guides/build-hook-safety.md)
- [Generator architecture](./docs/architecture/fiberx-generator-architecture.md)
- [Template boundaries](./docs/architecture/template-boundaries.md)
- [Repository rules](./docs/architecture/repository-rules.md)
- [Contributing](./CONTRIBUTING.md)
- [Changelog](./CHANGELOG.md)
- [Roadmap](./docs/roadmap/roadmap.md)

## Current Generator Tracks

- `medium`: stable production baseline with Swagger and embedded UI by default
- `heavy`: production-oriented track with Swagger, embedded UI, metrics, scheduler jobs, and optional Redis
- `light`: lightweight HTTP service with SQLite-first CRUD and optional Swagger or embedded UI
- `extra-light`: minimal startable base with SQLite startup, health endpoints, and recover-only middleware
- default stack: `Fiber v3 + Cobra + Viper`
- compatibility stack: `Fiber v2 + native-cli`
- runtime options on `medium / heavy / light`: `--logger`, `--db`, `--data-access`
- generated projects include config profiles, runtime metadata, upgrade inspection, and project-level build automation

## Quick Start

```bash
go run ./cmd/fiberx new demo --preset medium
cd demo
go run . serve
```

Compatibility example:

```bash
go run ./cmd/fiberx new demo-legacy --preset medium --fiber-version v2 --cli-style native
```

Runtime options example:

```bash
go run ./cmd/fiberx new demo-data --preset medium --logger slog --db pgsql --data-access sqlx
```

Build automation example:

```bash
go run ./cmd/fiberx build
go run ./cmd/fiberx build --dry-run
go run ./cmd/fiberx build --profile prod
```

## Repository Layout

- `sample/`: reference snapshots and test-facing examples, not the maintained generator mainline
- `output/`: local scratch space for generated artifacts and local binaries; ignored by Git except for `.gitkeep`

## v0.1.3 Release Scope

`v0.1.3` closes the current CLI and scaffold-hardening pass:

- generation plan preview with `new/init --print-plan [--json]`
- build safety switches such as `--no-hooks` and `--yes`
- layered `doctor` output for generator, project, and standalone modes
- `explain matrix` for preset and capability visibility
- improved verbose output separators for `validate`, `doctor`, and `explain matrix`
- safer default error responses in generated projects
- full timeout coverage for multi-handler business routes
- explicit-false config loading fixes
- lightweight explicit service initialization in generated business modules
- SQLite parent directory creation for default database paths

## v0.1.4 Release Scope

`v0.1.4` completed the generated common-layer consolidation work:

- unified the shared error and response layer for `light`, `medium`, and `heavy`
- simplified default controller error handling around the common response flow
- routed top-level API registration through `AppModules` to avoid growth-driven router sprawl
- kept `extra-light` intentionally minimal
- hardened generation regressions around the shared scaffold path

## v0.1.5 Current Scope

`v0.1.5` is the current release-surface synchronization milestone:

- align CLI help, validate, and doctor release wording with the actual repository state
- refresh `CHANGELOG.md` and release-facing docs
- keep README, docs index, usage guide, and roadmap in sync
- reduce confusion between the maintained generator mainline and reference-only snapshots

## Build Hook Safety

- `fiberx build` may execute project-defined hooks.
- Only run hooks in trusted repositories.
- Use `fiberx build --dry-run` to inspect planned commands before execution.

## License

This project is licensed under the MIT License. See [LICENSE](./LICENSE) for details.
