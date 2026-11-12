# myLlama

myLlama is a localhost control plane for one `llama.cpp` model runner. It
starts and supervises `llama-server`, serializes chat requests, and exposes
health, metrics, logs, profiles, and an OpenAI-shaped chat API. `llama.cpp`
continues to perform inference.

It is for a developer running a local GGUF model—not a public service, model
registry, or inference engine.

## Install a release

The first release train starts at **1.0.0-rc.1**. Stable 1.0 follows the
[release gates](ROADMAP.md); not every roadmap feature must ship first.

Download the matching macOS/Linux archive and checksums from
[Releases](https://github.com/cmkhoa/myLlama/releases), verify them using
[the release guide](docs/RELEASING.md), then extract the archive and run
`sh install.sh` inside it. The installer copies both commands to `~/.local/bin`
(or `INSTALL_DIR`). Add that directory to your `PATH`.

```sh
myLlamactl init
# Source the myLlama.env file reported by init.
myLlamactl doctor --model /absolute/path/to/model.gguf
myLlama
```

`llama-server` and model weights are installed separately. `doctor` is a preflight
check; occupied ports are expected if you already started the service.

## Quick start

Prerequisites: Go, a compatible `llama-server` executable on `PATH`, and a
local `.gguf` model.

```sh
go run ./cmd/myLlama
go run ./cmd/myLlamactl register /absolute/path/to/model.gguf
go run ./cmd/myLlamactl start
go run ./cmd/myLlamactl chat "Explain myLlama in one sentence."
go run ./cmd/myLlamactl stop
```

The API listens on `127.0.0.1:8080`; the managed runner listens on
`127.0.0.1:8081`. Non-loopback binds are rejected.

## Documentation

- [Getting started](docs/GETTING_STARTED.md) — profiles, the CLI, and a first request
- [API reference](docs/API.md) — endpoints, request shapes, streaming, and errors
- [Operations and configuration](docs/OPERATIONS.md) — environment variables, limits, metrics, recovery, and verification
- [Design log](DESIGN_LOG.md) — material architecture and product decisions
- [Roadmap](ROADMAP.md) — milestones and release gates
- [Feature proposals](docs/FEATURE_IDEAS.md) — plans for larger capabilities
- [Releases](docs/RELEASING.md) — versions, CI, distribution and upgrades
- [Contributing](CONTRIBUTING.md) — development and focused changes

## Development

```sh
make verify
make build
```

myLlama does not download models, accept arbitrary runner arguments, retain
prompts or completions, or support public/multi-user deployment.

## License

myLlama is licensed under the [MIT License](LICENSE). The `llama.cpp` runtime
and model weights are distributed separately under their respective licenses.
