# Usage Guide

This guide explains how to use the current `fiberx` generator from the repository root.

## Release Snapshot

- current release: `v0.1.4`
- current milestone: `v0.1.5`
- generatable presets: `heavy`, `medium`, `light`, `extra-light`
- implemented capabilities: `redis`, `swagger`, `embedded-ui`
- default stack: `fiber-v3 + cobra + viper`
- default runtime on `medium / heavy / light`: `zap + sqlite + stdlib`

## Run From Source

From the repository root:

```bash
go run ./cmd/fiberx --help
```

Available commands:

- `fiberx new <name>`
- `fiberx init`
- `fiberx list presets`
- `fiberx list capabilities`
- `fiberx explain preset <name>`
- `fiberx explain capability <name>`
- `fiberx explain matrix [--json]`
- `fiberx inspect [path]`
- `fiberx diff [path]`
- `fiberx upgrade inspect [path]`
- `fiberx upgrade plan [path]`
- `fiberx build [target...]`
- `fiberx validate [--verbose]`
- `fiberx doctor [--verbose]`

Equivalent source form:

```bash
go run ./cmd/fiberx <command>
```

## Create A New Project

Generate a new project into `<cwd>/<projectName>`:

```bash
go run ./cmd/fiberx new demo --preset medium
```

Preview before writing files:

```bash
go run ./cmd/fiberx new demo --preset medium --print-plan
go run ./cmd/fiberx new demo --preset medium --print-plan --json
```

Examples:

```bash
go run ./cmd/fiberx new demo --preset light
go run ./cmd/fiberx new demo --preset extra-light
go run ./cmd/fiberx new demo --preset medium --with redis
go run ./cmd/fiberx new demo --preset medium --fiber-version v2 --cli-style native
go run ./cmd/fiberx new demo --preset medium --logger slog --db pgsql --data-access sqlx
go run ./cmd/fiberx new demo --preset medium --json-lib sonic
go run ./cmd/fiberx new demo --preset light --db mysql --data-access sqlc
```

If `--module` is omitted, `fiberx` falls back to:

```text
github.com/example/<project-name>
```

## Initialize In The Current Directory

Generate into the current working directory:

```bash
go run ./cmd/fiberx init --preset light
```

With an explicit project name:

```bash
go run ./cmd/fiberx init --name demo --preset medium
go run ./cmd/fiberx init --name demo --preset light --json-lib go-json
go run ./cmd/fiberx init --preset medium --print-plan
```

## Inspect Presets And Capabilities

```bash
go run ./cmd/fiberx list presets
go run ./cmd/fiberx list capabilities
go run ./cmd/fiberx explain preset medium
go run ./cmd/fiberx explain capability redis
go run ./cmd/fiberx explain matrix
go run ./cmd/fiberx explain matrix --json
```

## Metadata, Diff, And Upgrade

```bash
go run ./cmd/fiberx inspect ./demo
go run ./cmd/fiberx diff ./demo
go run ./cmd/fiberx upgrade inspect ./demo
go run ./cmd/fiberx upgrade plan ./demo
```

## Build Generated Projects

Run from a generated project directory:

```bash
fiberx build
fiberx build server
fiberx build --target linux/amd64
fiberx build --profile prod
fiberx build --dry-run
fiberx build --no-hooks
fiberx build --yes
```

Generated projects can use:

- `fiberx.yaml`
- profiles
- archive and checksum output
- build metadata and release manifest
- target hooks and optional UPX compression

`fiberx build` may execute project-defined hooks. Only run hooks in trusted repositories. Use `fiberx build --dry-run` to inspect planned commands before execution. If hooks are present, interactive runs ask for confirmation by default. In non-interactive environments, use `--yes` to approve hooks or `--no-hooks` to skip them.

## Validate And Diagnose The Generator

Default outputs are intentionally short and release-oriented:

```bash
go run ./cmd/fiberx validate
go run ./cmd/fiberx doctor
```

Use verbose mode when you need the detailed matrix and internal diagnostics:

```bash
go run ./cmd/fiberx validate --verbose
go run ./cmd/fiberx doctor --verbose
```

Typical default `validate` output:

```text
fiberx validate: ok
release: v0.1.4
generator: <version> (<commit>)
presets: heavy,medium,light,extra-light
capabilities: redis,swagger,embedded-ui
default stack: fiber-v3 + cobra + viper
note: use --verbose for full diagnostics
```

Typical default `doctor` output:

```text
fiberx doctor
mode: generator
generator: <version> (<commit>)
release: v0.1.4
go: <runtime>
workspace: <cwd>
manifest root: <root>
status: ok
note: use --verbose for full diagnostics
```

When run inside a generated project, `doctor` automatically switches to project mode and summarizes metadata, diff status, and compatibility.

## Related Docs

- [Release process](./release-process.md)
- [Build hook safety](./build-hook-safety.md)
- [Template selection](./template-selection.md)
- [Capability policy](./capability-policy.md)
- [Config profiles](./config-profiles.md)
- [Generated project metadata](./generated-project-metadata.md)
- [Verification matrix](./verification-matrix.md)
- [Roadmap](../roadmap/roadmap.md)
