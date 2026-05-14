# Contributing

## Scope

`fiberx` is maintained as a generator-first repository. Changes should improve one of these areas:

- generator assets
- planning and validation rules
- rendering and metadata flows
- upgrade inspection
- build automation
- regression coverage
- release-facing documentation

## Workflow

1. Make focused changes.
2. Keep user-facing contracts stable unless the change intentionally updates them.
3. Update docs when the CLI, generated scaffold, or release surface changes.
4. Run the relevant verification lane before submitting work.

## Local Checks

Default local verification is the fast lane:

```bash
go test ./...
go run ./cmd/fiberx validate
go run ./cmd/fiberx doctor
```

Use the integration lane only when the change affects generated service startup, runtime database behavior, real build artifacts, or other heavy end-to-end paths:

```bash
go test -tags=integration ./cmd/fiberx ./internal/core
```

If the integration lane touches PostgreSQL or MySQL scenarios, provide the matching `FIBERX_TEST_*` DSNs or run through the GitHub Actions integration workflow.

## Repository Notes

- `generator/` is the maintained source of truth for scaffold behavior.
- `sample/` is reference-only material for comparison and discussion; it is not the source of truth and is not required to stay perfectly synchronized with every template change.
- `output/` is local scratch space and should stay out of version control except for `.gitkeep`.
- Generated release binaries should not be committed.

## Style

- Prefer small, explicit templates over broad string replacement.
- Keep generated code readable for everyday Go developers.
- Avoid introducing framework-like abstractions unless they clearly improve the generator mainline.
- Preserve compatibility rules for presets, capabilities, and runtime options.

## Release-Facing Changes

When a change affects release behavior, update the relevant documents:

- `README.md`
- `README_zh.md`
- `docs/README.md`
- `docs/guides/usage.md`
- `docs/roadmap/roadmap.md`
- `CHANGELOG.md`
