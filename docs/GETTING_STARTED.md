# Getting Started

## Prerequisites

- Go
- A compatible `llama.cpp` `llama-server` executable on `PATH` (or set
  `MYLLAMA_LLAMA_SERVER_PATH`)
- A local, non-empty `.gguf` model file

## Start a one-off model

In one terminal, run the service:

```sh
go run ./cmd/myLlama
```

In another, use the bundled CLI:

```sh
go run ./cmd/myLlamactl register /absolute/path/to/model.gguf
go run ./cmd/myLlamactl start
go run ./cmd/myLlamactl status
go run ./cmd/myLlamactl chat "Say hello in one sentence."
go run ./cmd/myLlamactl metrics
go run ./cmd/myLlamactl stop
```

Use `--stream` to print raw server-sent events:

```sh
go run ./cmd/myLlamactl chat --stream "Count from one to three."
```

Set `MYLLAMA_API_URL` or pass `--address http://127.0.0.1:8183` to point
the CLI at another local API address.

## Use named profiles

For repeatable settings, copy and edit
[the example profile file](../configs/profiles.example.json), then start with:

```sh
MYLLAMA_PROFILES_PATH=/absolute/path/to/profiles.json go run ./cmd/myLlama
go run ./cmd/myLlamactl profiles
go run ./cmd/myLlamactl activate qwen3-4b
go run ./cmd/myLlamactl start
```

Profiles include a model path, context size, GPU-layer setting, and optional
readiness/request timeouts, output-token limits, and queue capacity. Omitted
limits inherit the service defaults. A profile may be activated only while the
runner is stopped or unhealthy.

For endpoint-level examples, see the [API reference](API.md). For limits and
environment variables, see [Operations and configuration](OPERATIONS.md).

## First-run setup and scripts

Installed binaries support `myLlamactl init [--directory PATH]`. It creates private
`profiles.json` and `myLlama.env` files in the OS user configuration directory
under `myLlama`. It never overwrites either file. Source the reported environment
file before starting the server; add profiles manually or register a model through
the API. Setup does not install the runner, download weights, or start a service.

`myLlamactl doctor [--model PATH]` checks startup configuration, the runner
executable, configured profiles and port availability. It does not verify GGUF
contents, estimate memory or launch inference. An occupied port is reported as a
preflight failure, including when it belongs to an already-running myLlama.

Use global `--json` before the command for compact machine-readable output:

```sh
myLlamactl --json status
myLlamactl --json chat "Say hello."
myLlamactl --json doctor --model /absolute/path/to/model.gguf
myLlamactl --version
myLlama --version
```

JSON chat returns the entire completion, including usage. Errors are nonzero exits
with a JSON error on stderr when `--json` is set; successful command data goes to
stdout. Doctor prints its structured check report on stdout even when checks fail.
Streaming remains SSE and cannot be combined with `--json`.
