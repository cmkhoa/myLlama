# Larger feature proposals

These are plans, not implemented capabilities. Order follows the roadmap.
Effort is relative to this small Go codebase; it is not a delivery estimate.

## 1. Reproducible benchmarks — medium

**Problem:** average metrics cannot explain whether a model/settings change helped.

**First slice:** `myLlamactl bench --profile NAME --runs 10 --warmup 2
--output report.json`. Use a documented synthetic workload, drain between tests,
and separate cold start from warm inference. Require explicit activation; do not
silently switch someone else's active model.

Measure time to first content token using a monotonic clock, total latency,
queue wait where observable, completion tokens, generation rate, success/error
counts, P50/P95 and sample count. Explain percentile methods and distinguish
upstream timing from client-observed timing. Record myLlama/runtime revisions,
model identity/hash when available, OS/architecture and effective profile settings.
Do not store generated content or user prompts by default.

**Acceptance:** a mock stream verifies timing/event interpretation; failed trials
remain in the report; identical report schemas support comparisons; one real-model
report can be reproduced with a single documented command.

**Depends on:** version metadata and stable metrics. **Later:** profile comparisons,
regression thresholds, concurrency sweeps, energy measurements when available.

## 2. Local dashboard — medium

**Problem:** lifecycle and performance information is spread across CLI calls.

**First slice:** an embedded local web UI for status, profiles, start/stop, live
metrics, and a streaming test-chat panel. Serve bundled assets from the Go binary.
Use the existing API rather than a second implementation of lifecycle logic.

Show starting/stopping/unhealthy/idle states, request IDs on errors and connection
loss clearly. Keep logs bounded; do not expose arbitrary file reads. Never render
model output as trusted HTML. Use same-origin requests and validate Origin/Host
for browser-triggered mutations; localhost binding alone is not a browser access
control policy.

**Acceptance:** keyboard-accessible controls, cancellation, reconnect, profile
errors, long streams and narrow-screen layouts work in browser tests. The UI
cannot claim a model is ready until the service reports it.

**Depends on:** stable status/metrics and explicit frontend architecture decision.
**Later:** bounded events, benchmark comparisons, profile editor with validation.

## 3. Hardware-aware model profiles — medium to large

**Problem:** a valid path can still lead to an incompatible or oversized model.

**First slice:** inspect GGUF magic/version and bounded metadata: architecture,
quantization, tensor count, context hints and file size. Never allocate from an
untrusted length without limits. Report unsupported metadata instead of guessing.

Estimate model + KV cache + runtime headroom separately. Treat GPU/unified-memory
availability as a capability-specific observation, not a single portable number.
Display assumptions and ranges; do not promise a memory-fit guarantee. Recommend
settings first; automatic tuning needs separate approval/design.

**Acceptance:** truncated/corrupt/huge metadata fixtures fail safely; estimates
are compared against actual runs for several context sizes; CPU-only and Apple
Silicon cases have documented limitations.

**Depends on:** doctor, benchmarks. **Later:** calibrated safe presets, pre-load
memory checks and memory-aware switching/admission.

## 4. Safe model switching — medium

**Problem:** manually stopping and activating models can interrupt clients.

**First slice:** one explicit switch operation with states: requested, draining,
stopping, loading, ready, failed. Reject new inference with a retryable error
while draining; preserve already-admitted work until a bounded deadline. Keep
switch admission separate from inference admission.

Define cancellation and rollback: before stopping, cancellation leaves the old
model intact; after stopping, failure is visible and an optional rollback is a
separate bounded attempt. Track operation IDs to prevent duplicate switches.
Revalidate the target model before unloading the current one.

**Acceptance:** concurrent switches serialize or fail predictably; canceled
streams drain; failed load/rollback never reports ready; no orphan processes;
queue promises and timeout behavior are documented.

**Depends on:** lifecycle cleanup; benchmark-guided timeout defaults. Memory
estimates improve warnings but are not required for the first explicit switch.
**Later:** warm/cold caches and automatic routing.

## 5. Structured output — medium

**Problem:** generated prose is difficult to consume reliably in application code.

**First slice:** support a bounded `response_format` subset with JSON schema;
validate supported schema features before proxying and validate completed JSON
against the schema. Surface truncation and validation failure separately from
successful output. Streamed validation semantics must be designed explicitly;
start with non-streaming if needed.

Advertise model/runtime capabilities and fail early when unsupported. Constrained
generation still requires application-level validation of facts and semantics.

**Acceptance:** valid nested output, unsupported schemas, oversized/deep schemas,
malformed JSON, truncation and upstream errors have fixtures. Include a local
classification/extraction example with exact runtime revision.

**Depends on:** explicit compatibility contract. **Later:** typed CLI exports and
a separate tool-call transport proposal; do not silently execute generated tools.

## 6. Fair concurrency and optional sessions — large

**Problem:** one long stream blocks every other client with the current worker.

**First slice:** configurable parallel slots with one as the default. Define
per-client admission, bounded total queueing and fair selection. Align slot/context
settings with the tested runtime; measure memory costs before recommending more
parallelism. Every canceled request must free admission and eventually its slot.

Keep conversation sessions separate from inference slots. If sessions are added,
use opaque IDs, idle expiry, bounded count/history and explicit reset/delete.
No disk persistence by default; never reuse one client's history for another.

**Acceptance:** no starvation under mixed workloads; limits hold during disconnect
and timeout storms; session isolation tests; throughput, TTFT and memory comparison
for one/two slots. Include lifecycle operations during concurrent inference.

**Depends on:** benchmarks, resource observations and safe lifecycle coordination.
**Later:** priority classes, resumable work only with a separate retention design.

## 7. Embeddings and local retrieval — medium, then larger

**Problem:** chat alone cannot support semantic search over local material.

**First slice:** an embeddings endpoint with model capability validation, bounded
batch/input size, stable dimensions and token metrics. Require an embedding-capable
profile; do not assume the active chat model supports embeddings. Define explicit
model switching behavior while only one process is supported.

**Acceptance:** the same fixture input yields stable shape/order, bad dimensions
and unsupported profiles fail clearly, and a small local semantic-search example
works with a pinned model/runtime combination.

**Depends on:** capability discovery; switching or dedicated embedding mode.
**Later:** local ingestion with file size/type limits, chunk provenance, incremental
indexing and deletion; retrieval quality evaluation before a full document-chat UI.

## 8. Multiple backends and model routing — large / research

**Problem:** backend-specific APIs make it costly to support another local runtime.

**First slice:** describe capabilities and lifecycle behavior for a second backend
before abstracting code. Extract interfaces only around shared needs: start/stop,
health, chat, cancellation, usage and resource reporting. Keep backend-specific
settings explicit and reject unsupported operations.

**Acceptance:** the same contract suite passes for the reference backend and a
second implementation, including cancellation and process cleanup. Public API
behavior remains consistent; diagnostics retain backend details needed for support.

**Depends on:** concrete second-backend use case and stable contracts.
**Later:** memory-budgeted multi-model pools, warm/cold eviction, model routing,
transparent failover only after retry/idempotency rules are defined.

## Choosing the next feature

After the 1.0 release gates, start with benchmarks. They produce the evidence
needed for tuning and concurrency. Choose the dashboard next for daily usability,
or GGUF inspection for runtime depth. Keep these as separate changes with their
own acceptance criteria rather than combining them into a platform rewrite.
