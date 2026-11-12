# Changelog

Release entries describe implemented behavior. A version is published only when
its Git tag triggers the verified release workflow.

## [1.0.0-rc.1] — first release candidate

### Added

- Local GGUF runner lifecycle, readiness checks, bounded automatic recovery,
  named profiles and loopback-only API.
- Regular and streamed chat, cancellation, bounded queueing and request limits.
- Model listing, operational metrics and sanitized failure status.
- `myLlamactl init`, setup doctor, version metadata and compact JSON output.
- Optional private rotating JSON logs and opt-in idle model shutdown.
- Managed runner cleanup when the service exits.
- Linux/macOS checks, race and vulnerability scanning, and Python SDK smoke tests.
- Versioned macOS/Linux amd64/arm64 binary archives, SHA-256 checksums,
  local installer, smoke tests and GitHub artifact provenance.
- Long-term roadmap, larger feature proposals and release/contribution policies.

### Fixed

- Streaming handlers wait for queue-worker response writes to end on cancellation.
- Explicit stop also clears unhealthy states with no remaining child process.

### Changed

- Project name, commands, Go module, API ownership and environment prefix are
  `myLlama`, `myLlamactl` and `MYLLAMA_`.

### Known limits

- Release candidate: clean-machine and real-model validation are required before
  stable 1.0.0. Deterministic CI uses a mock runner, not actual inference.
- One active model and one inference worker; no embeddings, tool execution,
  public hosting, model downloads or Windows distribution.
- Chat compatibility covers the documented subset, not the complete OpenAI API.
- No prompts/completions retained; upstream process output is not persisted.
