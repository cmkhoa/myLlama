package config

import "testing"

func TestLogSettings(t *testing.T) {
	t.Setenv("MYLLAMA_LOG_PATH", "/tmp/myLlama.log")
	t.Setenv("MYLLAMA_LOG_MAX_BYTES", "4096")
	t.Setenv("MYLLAMA_LOG_BACKUPS", "2")
	cfg, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LogPath != "/tmp/myLlama.log" || cfg.LogMaxBytes != 4096 || cfg.LogBackups != 2 {
		t.Fatal("log settings not loaded")
	}
	t.Setenv("MYLLAMA_LOG_BACKUPS", "101")
	if _, err := FromEnv(); err == nil {
		t.Fatal("expected invalid backup count")
	}
}
