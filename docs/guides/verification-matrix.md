# Verification Matrix Guide

`fiberx` uses two verification lanes so daily work does not pay the cost of the full black-box and database matrix.

## Fast Lane

The default fast lane is the contract for local development and PR CI:

```bash
go test ./...
go run ./cmd/fiberx validate
go run ./cmd/fiberx doctor
```

Fast lane coverage includes:

- CLI output and validation contracts
- generation contract tests across presets, capabilities, and runtime selections
- metadata, inspect, diff, and upgrade main paths
- representative generated project compile smoke for `extra-light default` and `heavy default`
- dry-run build planning and build config validation

Fast lane does not require PostgreSQL or MySQL services.

## Integration Lane

The integration lane is reserved for heavy end-to-end coverage:

```bash
go test -tags=integration ./cmd/fiberx ./internal/core
```

Integration lane coverage includes:

- generated service startup and HTTP black-box checks
- runtime database matrix across SQLite, PostgreSQL, and MySQL
- generated CRUD, docs, UI, metrics, gzip, and scheduler paths
- real CLI build artifact generation, archive output, release metadata, and hook execution behavior

When PostgreSQL or MySQL scenarios are needed, provide `FIBERX_TEST_PGSQL_DSN` and `FIBERX_TEST_MYSQL_DSN` or use the GitHub Actions integration workflow.

## CI Mapping

- `.github/workflows/ci.yml` runs the fast lane on pull requests and pushes to `dev` / `main`.
- `.github/workflows/integration.yml` runs the integration lane on a nightly schedule and on manual dispatch.
- Release preparation should treat the integration lane as an explicit gate instead of making every PR pay the full cost.

## Sample Boundary

- `generator/` remains the maintained source of truth.
- `sample/` is reference-only material for human comparison, screenshots, and discussion.
- A `sample/` drift is not automatically a generator bug unless it contradicts the maintained generator surface or current regression contract.
