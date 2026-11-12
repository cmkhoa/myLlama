# myLlama roadmap

myLlama is a long-term local AI runtime project. Its focus is dependable model
operations, useful measurements, and a clear developer experience. `llama.cpp`
performs inference; myLlama owns lifecycle, admission, configuration, diagnostics,
and a deliberately documented API.

Milestones are ordered by dependencies and evidence, not calendar promises.
[Feature proposals](docs/FEATURE_IDEAS.md) describe larger investments and
[release policy](docs/RELEASING.md) defines how changes reach users.

## Current foundation

- [x] One managed GGUF runner with validated local paths and localhost binding.
- [x] Register/select profiles, start/stop, readiness deadlines, bounded recovery.
- [x] Regular and streamed chat, cancellation, bounded FIFO admission.
- [x] Per-profile request limits; operational metrics and sanitized status errors.
- [x] Model-list endpoint and a terminal client.
- [x] Source rename to myLlama, including `MYLLAMA_` settings.

## 1.0 release train — make the foundation distributable

The first version is **1.0.0-rc.1**, a release candidate. Stable **1.0.0** follows
validation of the supported contracts. It does not require every future feature.

Delivered in this release-candidate implementation:

- [x] First-run configuration, setup doctor, build version information.
- [x] CLI JSON output, including full non-streaming chat completions.
- [x] Private, bounded, rotating operational log files.
- [x] Opt-in idle model shutdown; managed process cleanup on service shutdown.
- [x] Streaming cancellation race fix and race checks.
- [x] Real Python SDK tests against a managed mock runner, with CLI and lifecycle checks.
- [x] Linux/macOS CI, vulnerability checks, four binary archives, checksums,
  local installation smoke tests, and tag-driven releases with provenance.
- [x] Documented release, compatibility, contribution, and upgrade policies.

Promotion gates, still to complete before a stable release:

- [ ] Merge the implementation after CI passes; publish and exercise the RC.
- [ ] Install downloaded RC archives on clean Linux and macOS machines.
- [ ] Record the exact tested `llama.cpp` revision and real GGUF profile/settings.
- [ ] Run real-model chat, streaming cancellation, restart, idle shutdown, and
  exit cleanup; attach a short reproducible report to the release.
- [ ] Confirm supported CLI/API/settings contracts and document known limits.
- [ ] Set `VERSION` and changelog to `1.0.0`, pass all checks, then tag `v1.0.0`.

Cross-compilation is packaging evidence, not proof that inference works on every
CPU/GPU. No Windows release is advertised until its process shutdown and native
installation tests exist. No model weights or inference runtime are bundled.

## 1.1 — measure and explain

- [ ] Benchmark command: warm-up, repeated trials, TTFT, throughput, P50/P95,
  error rate, and exact machine/model/runtime metadata.
- [ ] GGUF inspection with bounded parsing and clear unsupported-format errors.
- [ ] Expand doctor to flag incompatible runtime versions and conflicting settings.
- [ ] Save benchmark reports without storing prompts or responses by default.

Exit: a developer can reproduce a profile comparison and diagnose startup
problems without guessing. See proposals 1 and 3.

## 1.2 — see and control

- [ ] Local dashboard with runner state, profile selection, queue/metrics, and chat.
- [ ] Safe profile switching with request drain, visible progress, rollback policy,
  cancellation and explicit failure states.
- [ ] API events or bounded polling, with clear connection/recovery behavior.

Exit: switching a model never strands a queued request or leaves an orphan
process, and the UI accurately reflects transitions. See proposals 2 and 4.

## 1.3 — predictable application outputs

- [ ] Schema-constrained JSON output with bounded schemas and response validation.
- [ ] Capability discovery per runtime/model; unsupported requests fail clearly.
- [ ] Embeddings API with purpose-built embedding profiles and dimension metadata.
- [ ] Expand the SDK compatibility matrix only for tested fields.

Exit: an extraction or semantic-search example runs entirely locally through a
documented API. See proposals 5 and 7.

## 1.4 — resource-aware concurrency

- [ ] Fair, bounded scheduling across clients with opt-in parallel slots.
- [ ] Memory/context admission checks and cancellation that frees capacity.
- [ ] Optional in-memory sessions with expiry, deletion and isolation tests.
- [ ] Compare one versus multiple slots using the benchmark suite.

Exit: independent requests do not starve each other or mix history, and configured
limits remain enforceable. See proposals 3 and 6.

## Later / research backlog

- Warm/cold multi-model scheduling and idle resource policies.
- Tool-call request/response transport, followed by explicit tool execution design.
- Local document retrieval after embeddings are stable.
- Resumable model import/download with checksums and explicit storage ownership.
- Additional backends, including evaluating MLX when it meets the runtime contract.
- Native Windows support; package-manager distribution; background-service setup.
- Bounded metrics export and optional diagnostic bundles with preview/redaction.

Each needs a proposal, measured user value, and acceptance criteria before it
enters a release milestone. A new major version is justified by a breaking public
contract, not by the size or novelty of a feature.

## Ongoing maintenance

Every milestone includes regression tests for its failure paths, current supported
Go CI, dependency/action updates, accurate docs, changelog entries, and a rollback
path. Real-model validation complements deterministic CI; neither substitutes for
the other. Prompts/completions remain transient unless an explicit opt-in feature
introduces retention with clear deletion controls.
