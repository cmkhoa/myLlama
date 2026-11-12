# Contributing to myLlama

Use [ROADMAP.md](ROADMAP.md) for milestones and
[docs/FEATURE_IDEAS.md](docs/FEATURE_IDEAS.md) for larger proposals. Discuss a
proposal before introducing a new backend, API surface, persistence mechanism
or scheduler. Keep fixes and features in focused, independently buildable commits.

## Development

Go 1.24+ is the source minimum. Prefer the latest supported Go for development;
release CI uses current stable. No inference model is needed for deterministic
checks:

```sh
make fmt
make verify
make race
make build
```

For SDK/packaging validation, see [RELEASING.md](docs/RELEASING.md).

## Change expectations

- Describe the concrete trigger and resulting behavior in the PR.
- Add tests for failure paths, cancellation and resource bounds when relevant.
- Keep API/settings/CLI docs and the current changelog accurate.
- Never commit models, local paths/profiles, prompts, completions, credentials or
  generated outputs. Operational logs exclude request/response content.
- Preserve loopback-only defaults. Any browser UI needs its own origin policy.
- Add dependencies only when the capability warrants them; keep update ownership
  clear. Use small commits with descriptive prefixes such as `fix:`, `feat:`,
  `docs:` and `ci:`. Prefixes do not trigger releases.

Run a real-model smoke test when changing runner arguments, inference behavior,
readiness or process lifecycle. Record runtime revision, model, effective settings
and platform, while avoiding redistribution of weights or private prompts.
