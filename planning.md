# myLlama development plan

The original weekend plan is superseded by [ROADMAP.md](ROADMAP.md). This project
now uses small, independently reviewable changes and release milestones.

## Current delivery sequence

1. Rename the source, commands, module, settings and docs to myLlama.
2. Fix the streaming cancellation race found by the new race check.
3. Add first-run setup, doctor, version metadata and CLI JSON chat output.
4. Add bounded private log rotation and document its operational ownership.
5. Add opt-in idle shutdown and service-exit process cleanup.
6. Add SDK contract coverage and document its exact supported subset.
7. Add build archives, installation smoke tests, CI and release automation.
8. Update long-term roadmap, proposals and contribution/release guidance.

Each change is a separate commit; logging/lifecycle and compatibility commits
include relevant behavior tests. Keep release automation last among executable
changes so it can verify the complete feature set. No model files, local profiles,
logs, build outputs, or temporary environments enter version control.

## Release candidate validation

Use `1.0.0-rc.1` as the first candidate. The release pipeline publishes a version
only after a maintainer pushes its explicit tag. A passing pull request does not
create a public release. Before stable 1.0, complete the real-model and clean
installation checks in the roadmap and record the tested runtime revision.

## Next implementation proposal

Start with benchmarks from [FEATURE_IDEAS.md](docs/FEATURE_IDEAS.md). Agree on the
report schema and measurement definitions before implementing the command. Keep
benchmark infrastructure separate from the dashboard, GGUF parser and scheduler.
