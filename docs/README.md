# Docs

This directory contains the maintained design notes, guides, and release roadmap for the `fiberx` generator mainline.

## Release Status

- `v0.1.0`: completed
- `v0.1.1`: completed
- `v0.1.2`: completed
- `v0.1.3`: completed
- `v0.1.4`: completed
- `v0.1.5`: in progress

## Current Scope

- `v0.1.2`: shared scaffold uplift, timeout routing, response compatibility, and release-facing documentation
- `v0.1.3`: CLI preview UX, build safety switches, layered doctor output, explain matrix, and scaffold hardening
- `v0.1.4`: common error and response layer consolidation plus `AppModules` route entry for generated `light / medium / heavy` scaffolds
- `v0.1.5`: release wording alignment across CLI outputs, changelog, usage guide, and top-level docs

## Core Documents

- [`architecture/fiberx-generator-architecture.md`](./architecture/fiberx-generator-architecture.md)
- [`architecture/template-boundaries.md`](./architecture/template-boundaries.md)
- [`architecture/repository-rules.md`](./architecture/repository-rules.md)
- [`guides/usage.md`](./guides/usage.md)
- [`guides/release-process.md`](./guides/release-process.md)
- [`guides/build-hook-safety.md`](./guides/build-hook-safety.md)
- [`guides/template-selection.md`](./guides/template-selection.md)
- [`guides/capability-policy.md`](./guides/capability-policy.md)
- [`guides/config-profiles.md`](./guides/config-profiles.md)
- [`guides/generated-project-metadata.md`](./guides/generated-project-metadata.md)
- [`guides/response-contract.md`](./guides/response-contract.md)
- [`guides/verification-matrix.md`](./guides/verification-matrix.md)
- [`guides/deployment-runbook.md`](./guides/deployment-runbook.md)
- [`roadmap/roadmap.md`](./roadmap/roadmap.md)

## Repository Notes

- `sample/` is kept as reference snapshots and test-facing examples.
- `output/` is local scratch space and stays out of version control except for `.gitkeep`.

Start from the root README for repository positioning, then use the generator architecture document when making structural decisions.
