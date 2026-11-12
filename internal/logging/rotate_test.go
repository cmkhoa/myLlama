package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRotationBoundsRetentionAndPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.log")
	w, err := Open(path, 8, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range []string{"1111\n", "2222\n", "3333\n", "4444\n"} {
		if _, err := w.Write([]byte(record)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{path: "4444\n", path + ".1": "3333\n", path + ".2": "2222\n"} {
		got, err := os.ReadFile(name)
		if err != nil || string(got) != want {
			t.Fatalf("%s = %q, %v", name, got, err)
		}
		info, _ := os.Stat(name)
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("mode = %v", info.Mode())
		}
	}
	if _, err := os.Stat(path + ".3"); !os.IsNotExist(err) {
		t.Fatalf("unexpected extra backup: %v", err)
	}
}

func TestOversizedRecordDoesNotGrowFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.log")
	w, err := Open(path, 4, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	if _, err := w.Write([]byte("too large")); err == nil {
		t.Fatal("expected size error")
	}
	if _, err := w.Write([]byte("ok\n")); err != nil {
		t.Fatal(err)
	}
}
