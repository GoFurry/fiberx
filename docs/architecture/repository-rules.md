# Repository Rules

These rules keep `fiberx` maintainable as a long-lived generator repository.

## Preset Evolution Rules

- Do not add new official preset tiers for one-off combinations.
- Do not move optional capabilities into preset defaults unless they are part of that preset's high-frequency path.
- Do not turn business demos into feature-rich example applications inside generated output.

## Generator Ownership Rules

- Generator assets under `generator/` are the maintained source of truth for scaffolding behavior.
- Runtime options, capability policy, metadata, upgrade inspection, and build engineering must evolve inside the generator mainline.
- `sample/` is reference-only output, not a second source of truth.
- Do not reintroduce parallel repository-local legacy systems as a second maintenance surface.

## Documentation Rules

- Root README explains repository positioning and points to deeper docs.
- The generator architecture document is the top-level design basis for future implementation work.
- Long-term architecture and evolution rules live under `docs/`, not in generated project READMEs.
- Docs should describe the maintained generator surface, not removed historical directory layouts.

## Quality Rules

- Root `go test ./...` is the fast-lane regression entrypoint.
- Fast CI should validate the maintained generator mainline without external database services.
- Heavy black-box, runtime database, and real build artifact checks belong to the explicit integration lane.
- Release gates should require the integration lane, but routine PR validation should not.

## Scope Rules

- `heavy` is the upper bound for built-in engineering baseline.
- `extra-light` is the lower bound for built-in engineering baseline.
- Future complexity should be added through generator work, docs, runtime options, or build engineering, not by reviving separate in-repo extension systems.
