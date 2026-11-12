package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSONChatPreservesCompletionAndUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path=%s", r.URL.Path)
		}
		w.Write([]byte(`{"choices":[{"message":{"content":"hello"}}],"usage":{"completion_tokens":1}}`))
	}))
	defer server.Close()
	var out, stderr bytes.Buffer
	if err := run([]string{"--address", server.URL, "--json", "chat", "hi"}, &out, &stderr); err != nil {
		t.Fatal(err)
	}
	if !json.Valid(out.Bytes()) || !strings.Contains(out.String(), `"completion_tokens":1`) {
		t.Fatal(out.String())
	}
}

func TestInitDoesNotOverwriteConfiguration(t *testing.T) {
	directory := t.TempDir()
	var out, stderr bytes.Buffer
	args := []string{"init", "--directory", directory}
	if err := run(args, &out, &stderr); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "profiles.json")
	if err := os.WriteFile(path, []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run(args, &out, &stderr); err == nil {
		t.Fatal("expected overwrite refusal")
	}
	got, _ := os.ReadFile(path)
	if string(got) != "keep me" {
		t.Fatal("configuration overwritten")
	}
}

func TestDoctorReportsConfigurationFailureAsJSON(t *testing.T) {
	t.Setenv("MYLLAMA_LISTEN_ADDRESS", "0.0.0.0:8080")
	var out, stderr bytes.Buffer
	err := run([]string{"--json", "doctor"}, &out, &stderr)
	if err == nil || !json.Valid(out.Bytes()) || !strings.Contains(out.String(), `"ok":false`) {
		t.Fatalf("report=%s, error=%v", out.String(), err)
	}
}
