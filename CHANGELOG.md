# Changelog

## Unreleased

- Aligned CLI help, validate, and doctor release wording around `v0.1.4` as the current completed release and `v0.1.5` as the active milestone
- Synced release-facing docs across `README`, `README_zh`, `docs/README.md`, `docs/guides/usage.md`, and `docs/roadmap/roadmap.md`
- Refreshed audit notes to match the current post-`v0.1.4` repository state

## v0.1.4

- Unified the generated `light`, `medium`, and `heavy` scaffold error model around `APIError` / `AppError`
- Unified the generated response helper flow around `common.Success`, `common.Error`, and `NewResponse`
- Simplified generated controller error handling and kept service error mapping on the common response path
- Routed generated application registration through `AppModules` to avoid route-registration sprawl
- Kept `extra-light` intentionally minimal while hardening regression coverage for the shared scaffold path

## v0.1.3

- Added generation plan preview with `new/init --print-plan [--json]`
- Added build hook safety switches such as `build --no-hooks` and `build --yes`
- Added layered `doctor` modes for generator, project, and standalone environments
- Added `explain matrix` for preset and capability visibility
- Improved verbose output structure for `validate`, `doctor`, and `explain matrix`
- Hardened generated default error responses and multi-handler timeout coverage
- Fixed explicit `false` config override handling and removed empty default hook lists from generated `fiberx.yaml`
- Simplified generated business-module initialization and auto-created SQLite parent directories for default paths

## v0.1.2

- Added shared scaffold constants for `light`, `medium`, and `heavy`
- Added a base application error model and response compatibility layer
- Added configurable request timeout support for business routes
- Added default `middleware.timeout` config to generated `light`, `medium`, and `heavy` projects
- Kept system routes outside timeout wrapping
- Kept `extra-light` as the minimal scaffold

## v0.1.1

- Added Fiber v3 app hooks skeleton with `app.Hooks()` integration
- Added default graceful shutdown wiring for generated Fiber v2 and v3 projects
- Added stronger default middleware setup for `medium`, `heavy`, and `light`
- Added optional JSON backend selection with `--json-lib stdlib|sonic|go-json`
- Added `json_lib` to generated project metadata and inspection flows
- Documented the trust boundary for build hooks

## v0.1.0

- Added four official presets: `heavy`, `medium`, `light`, `extra-light`
- Added stable capability support for `redis`, `swagger`, and `embedded-ui`
- Added runtime selection for logger, database, and data access options
- Added generated project metadata, diff inspection, and readonly upgrade inspection
- Added project-level build automation with profiles, packaging, checksums, hooks, UPX, build metadata, and release manifest output
- Reduced the repository to the generator mainline only
