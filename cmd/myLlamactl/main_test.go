package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
