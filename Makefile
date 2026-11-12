.PHONY: fmt check-format test race vet verify run build release compatibility

fmt:
	gofmt -w $$(find cmd internal -name '*.go' -type f)

check-format:
	@test -z "$$(gofmt -l $$(find cmd internal -name '*.go' -type f))"

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

verify: check-format test vet

run:
	go run ./cmd/myLlama

build:
	go build -trimpath -o bin/myLlama ./cmd/myLlama
	go build -trimpath -o bin/myLlamactl ./cmd/myLlamactl

release:
	python3 scripts/release.py

# Install tests/compat/requirements.txt in a virtual environment first.
compatibility: build
	python3 tests/compat/test_sdk.py
