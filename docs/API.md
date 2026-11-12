# API Reference

All endpoints are localhost-only. JSON errors use:

```json
{"error":{"code":"...","message":"...","request_id":"..."}}
```

Every response includes the same ID in `X-Request-ID`.

| Method | Path | Purpose |
| --- | --- | --- |
| `POST` | `/v1/models/register` | Validate and select a one-off local GGUF path |
| `GET` | `/v1/models/profiles` | List configured model profiles |
| `GET` | `/v1/models` | List OpenAI-shaped local model IDs |
| `POST` | `/v1/models/activate` | Select a profile while stopped or unhealthy |
| `POST` | `/v1/runners/start` | Launch the selected runner and wait for readiness |
| `POST` | `/v1/runners/stop` | Gracefully stop the managed runner |
| `GET` | `/v1/runners/status` | Inspect lifecycle state and failure diagnostic |
| `POST` | `/v1/chat/completions` | Submit a regular or streamed completion |
| `GET` | `/metrics` | Inspect process-lifetime operational metrics |

## Model and runner lifecycle

```sh
curl -X POST http://127.0.0.1:8080/v1/models/register \
  -H 'Content-Type: application/json' \
  -d '{"path":"/absolute/path/to/model.gguf"}'

curl -X POST http://127.0.0.1:8080/v1/runners/start
curl http://127.0.0.1:8080/v1/runners/status
curl -X POST http://127.0.0.1:8080/v1/runners/stop
```

Activate a named profile with `{"name":"profile-name"}` at
`POST /v1/models/activate`.

`GET /v1/models` intentionally returns profile/model IDs, never local
filesystem paths.

## Chat completions

Send one to 64 `system`, `user`, or `assistant` messages:

```sh
curl http://127.0.0.1:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{
    "messages": [{"role":"user","content":"Say hello in one sentence."}],
    "max_tokens": 64,
    "temperature": 0,
    "chat_template_kwargs": {"enable_thinking": false}
  }'
```

Set `"stream": true` and use `curl --no-buffer` for SSE. A stream ends with
`data: [DONE]`. It retains the single queue worker until the response ends.

```sh
curl --no-buffer http://127.0.0.1:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{"messages":[{"role":"user","content":"Count to three."}],"stream":true}'
```

The supported subset is intentionally small: `model`, messages, `max_tokens`,
`temperature`, `stream`, `stream_options.include_usage`, and `chat_template_kwargs`.
If `model` is supplied, it must match the active profile ID (or registered model
basename); an inactive/unknown model returns HTTP 404 with `model_not_found`.
Listing a configured model does not load it or route requests to it automatically.
`stream_options` requires `stream=true`; usage events depend on upstream support. See
[Operations](OPERATIONS.md) for configurable limits and error behavior.

## Client compatibility contract

CI uses the real Python OpenAI SDK **2.8.1** against the actual service binary and
a managed deterministic mock runner. It covers model listing, regular completions,
SSE completion chunks, usage events, unknown-model errors and invalid token limits.
No cloud requests or actual API key are used. This follows the client patterns in
[the official Python documentation](https://developers.openai.com/api/reference/python).

For a local client, set `base_url="http://127.0.0.1:8080/v1"` and a dummy
`api_key="local-only"`. The service has no API-key authentication; the dummy value
only satisfies the SDK constructor. Remain on localhost.

This is a tested subset, not full OpenAI compatibility. Unsupported request fields
are rejected rather than silently ignored. Tool messages, multimodal content,
Responses API, embeddings and structured output are future work. CI fixtures do
not prove every llama.cpp build/model emits the same response shape; pin and record
real runtime/model revisions before a stable release.
