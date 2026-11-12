# Operations and Configuration

myLlama is process-level configuration: settings are supplied at startup,
not through HTTP. Restart the service after changing them.

| Variable | Default | Purpose |
| --- | --- | --- |
| `MYLLAMA_LISTEN_ADDRESS` | `127.0.0.1:8080` | myLlama API bind address |
| `MYLLAMA_LLAMA_SERVER_PATH` | `llama-server` | Runner executable path or command |
| `MYLLAMA_RUNNER_ADDRESS` | `127.0.0.1:8081` | Managed runner bind address |
| `MYLLAMA_CONTEXT_SIZE` | `2048` | Runner context-token limit |
| `MYLLAMA_GPU_LAYERS` | `99` | Requested GPU-offloaded layers |
| `MYLLAMA_READINESS_TIMEOUT` | `60s` | Start/readiness deadline |
| `MYLLAMA_SHUTDOWN_TIMEOUT` | `10s` | Graceful-stop deadline |
| `MYLLAMA_AUTO_RESTART` | `false` | Enable bounded restart after unexpected exit |
| `MYLLAMA_AUTO_RESTART_DELAY` | `2s` | Delay before a restart attempt |
| `MYLLAMA_MAX_AUTO_RESTARTS` | `3` | Restart-attempt budget |
| `MYLLAMA_MAX_REQUEST_BYTES` | `1048576` | Largest JSON request body |
| `MYLLAMA_MAX_INPUT_BYTES` | `32768` | Combined message-content size |
| `MYLLAMA_MAX_GENERATED_TOKENS` | `512` | Largest accepted `max_tokens` |
| `MYLLAMA_DEFAULT_MAX_TOKENS` | `256` | Output limit when omitted |
| `MYLLAMA_QUEUE_CAPACITY` | `16` | Waiting-request capacity before HTTP 429 |
| `MYLLAMA_REQUEST_TIMEOUT` | `2m` | End-to-end chat deadline |
| `MYLLAMA_PROFILES_PATH` | unset | Path to named-profile JSON |
| `MYLLAMA_IDLE_TIMEOUT` | `0s` (disabled) | Idle time before unloading the model |
| `MYLLAMA_LOG_PATH` | unset (stdout) | Explicit JSON log file; parent must exist |
| `MYLLAMA_LOG_MAX_BYTES` | `10485760` | Maximum bytes per log file |
| `MYLLAMA_LOG_BACKUPS` | `3` | Retained numbered backups, from 1 to 100 |

The queue has one worker. Requests execute FIFO; canceled queued work is
skipped. A long-lived stream occupies that worker until it ends. Named profiles
may override `readiness_timeout`, `request_timeout`, `max_generated_tokens`,
`default_max_tokens`, and `queue_capacity`; omitted values inherit this table.

## Health, recovery, and diagnostics

`GET /v1/runners/status` reports `stopped`, `starting`, `ready`, `unhealthy`,
or `stopping`. Failed readiness checks and unexpected child exits produce a
bounded, sanitized `last_error`. It is cleared after a successful start or an
intentional stop.

Automatic restart is off by default. When enabled, an unexpected exit becomes
visible as `unhealthy`, then retries after the configured delay until its budget
is exhausted. Explicit stops never restart.

## Metrics and logs

`GET /metrics` reports process-lifetime request counts, error categories, queue
wait, latency, generated tokens, and generation rate. Metrics reset on service
restart. Structured logs contain operational fields and request IDs, never
prompts or generated text.

### File logs and privacy

Set `MYLLAMA_LOG_PATH` to enable file output instead of stdout. Files use owner-only
permissions (`0600`) and rotate before a record would exceed the size limit.
Backups `.1` (newest) through `.N` are retained; total retained data is at most
`(N + 1) * max bytes` for files created under these limits. Existing oversized
files/settings changes may temporarily exceed that bound until rotation. Rotation
is size/count based, not age based. One service process must own each path; do not
share it between instances or rotate it with another tool. Parent directories
must exist, and disk/write errors can prevent log persistence. The file writer
rejects records larger than the file limit; choose a practical limit (default 10 MiB).

Logs include request IDs, timing, error categories and process events, never
request bodies or generated content. Raw upstream stdout/stderr is debug-level
and excluded by the fixed info-level logger. Paths and platform error messages
can still appear; inspect logs before sharing them. This guarantee does not cover
logging performed independently by the inference runtime.

### Idle shutdown and service exit

Set `MYLLAMA_IDLE_TIMEOUT=15m` to unload the runner after inactivity. Zero disables
it; negative durations are rejected. Queued work, active streams and lifecycle
POSTs keep it busy. Read-only status/metrics polling does not reset the timer.
Checks run at most once a second. The API stays available; restarting the runner
requires an explicit `start`, with the selected model retained in memory.

The idle decision and new admissions are coordinated so a model cannot be stopped
under an admitted chat request. New requests arriving after idle stop require a
restart. On SIGINT/SIGTERM the service drains HTTP requests for up to 10 seconds,
closes remaining connections, then stops its managed child with the configured
grace period and kill escalation. This does not make SIGKILL recoverable.

## Verify and build

```sh
make verify
make build
```

The initial measured Qwen run on this machine completed 10 sequential requests
with no failures, averaging 87.8 ms end-to-end and 41.0 generated tokens/sec.
This is a machine-specific observation, not a general performance claim.
