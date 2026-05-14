# Release Process

This guide describes the lightweight release flow for the `fiberx` generator.

## 1. Confirm Release Scope

- verify the roadmap entry is up to date
- verify README and docs use the correct release wording
- confirm generated scaffold changes are reflected in docs and examples

## 2. Run Fast Lane

From the repository root:

```bash
go test ./...
go run ./cmd/fiberx validate
go run ./cmd/fiberx doctor
```

This is the default release sanity check and should stay cheap enough to run often.

## 3. Run Integration Lane

Before tagging a release, run the heavy lane locally or through GitHub Actions:

```bash
go test -tags=integration ./cmd/fiberx ./internal/core
```

If the release touches PostgreSQL or MySQL scenarios, provide the matching `FIBERX_TEST_PGSQL_DSN` and `FIBERX_TEST_MYSQL_DSN` values or rely on the `Integration` workflow.

## 4. Review Release-Facing Surface

Check these files before tagging:

- `README.md`
- `README_zh.md`
- `docs/README.md`
- `docs/guides/usage.md`
- `docs/roadmap/roadmap.md`
- `CHANGELOG.md`

## 5. Prepare Release Notes

Release notes should stay short and focus on user-visible changes:

- new generator features
- scaffold changes
- build or upgrade behavior changes
- documentation or release-surface changes when relevant
- verification contract changes when they affect contributor workflow

Avoid internal phase history or implementation detail dumps.

## 6. Tag And Publish

Recommended sequence:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

Then create the GitHub release and use the prepared release notes.
