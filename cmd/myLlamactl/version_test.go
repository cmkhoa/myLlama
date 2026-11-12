package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestVersionWorksWithoutService(t *testing.T) {
	var out, stderr bytes.Buffer
	if err := run([]string{"--version"}, &out, &stderr); err != nil {
		t.Fatal(err)
	}
	if !json.Valid(out.Bytes()) {
		t.Fatal(out.String())
	}
}
