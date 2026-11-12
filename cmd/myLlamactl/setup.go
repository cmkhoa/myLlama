package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"myLlama/internal/config"
	"myLlama/internal/runner"
)

func runInit(args []string, out, stderr io.Writer) error {
	base, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(stderr)
	directory := flags.String("directory", filepath.Join(base, "myLlama"), "configuration directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("usage: myLlamactl init [--directory PATH]")
	}
	path, err := filepath.Abs(*directory)
	if err != nil {
		return err
	}
	files := map[string]string{
		"profiles.json": "{\n  \"profiles\": {}\n}\n",
		"myLlama.env":   "# Source this file before starting myLlama.\nMYLLAMA_PROFILES_PATH='" + strings.ReplaceAll(filepath.Join(path, "profiles.json"), "'", "'\\''") + "'\nexport MYLLAMA_PROFILES_PATH\n# MYLLAMA_IDLE_TIMEOUT=15m\n# MYLLAMA_LOG_PATH=/absolute/path/to/myLlama.log\n",
	}
	for name := range files {
		if _, err := os.Lstat(filepath.Join(path, name)); !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("refusing to overwrite %s", filepath.Join(path, name))
		}
	}
	if err := os.MkdirAll(path, 0o700); err != nil {
		return err
	}
	created := []string{}
	for name, body := range files {
		file, err := os.OpenFile(filepath.Join(path, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		_, writeErr := io.WriteString(file, body)
		closeErr := file.Close()
		if err := errors.Join(writeErr, closeErr); err != nil {
			return err
		}
		created = append(created, filepath.Join(path, name))
	}
	return json.NewEncoder(out).Encode(map[string]any{"directory": path, "created": created, "next": "Source myLlama.env, then run myLlamactl doctor and start myLlama. Register a model or add a profile before starting the runner."})
}

type check struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

func runDoctor(args []string, out, stderr io.Writer, asJSON bool) error {
	flags := flag.NewFlagSet("doctor", flag.ContinueOnError)
	flags.SetOutput(stderr)
	model := flags.String("model", "", "optional local GGUF path to validate")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("usage: myLlamactl doctor [--model PATH]")
	}
	checks := []check{}
	cfg, err := config.FromEnv()
	checks = append(checks, check{"configuration", err == nil, "valid startup settings"})
	if err != nil {
		checks[0].Detail = err.Error()
	} else {
		path, err := exec.LookPath(cfg.Runner.Executable)
		detail := path
		if err != nil {
			detail = "Install llama.cpp or set MYLLAMA_LLAMA_SERVER_PATH to its executable: " + err.Error()
		}
		checks = append(checks, check{"llama-server", err == nil, detail})
		manager := runner.NewManager(cfg.Runner, nil)
		for _, profile := range cfg.Profiles {
			err := manager.AddProfile(profile)
			detail := "valid profile"
			if err != nil {
				detail = err.Error()
			}
			checks = append(checks, check{"profile " + profile.Name, err == nil, detail})
		}
		if *model != "" {
			_, err := manager.Register(*model)
			detail := "nonempty GGUF path; model contents and memory fit are not verified"
			if err != nil {
				detail = err.Error()
			}
			checks = append(checks, check{"model", err == nil, detail})
		}
		for _, address := range []string{cfg.ListenAddress, cfg.Runner.Address} {
			listener, err := net.Listen("tcp", address)
			detail := "port available (point-in-time check)"
			if err != nil {
				detail = "port unavailable; stop the existing service or change its address: " + err.Error()
			} else {
				listener.Close()
			}
			checks = append(checks, check{"port " + address, err == nil, detail})
		}
	}
	ok := true
	for _, c := range checks {
		ok = ok && c.OK
	}
	if asJSON {
		if err := json.NewEncoder(out).Encode(map[string]any{"ok": ok, "checks": checks}); err != nil {
			return err
		}
	} else {
		for _, c := range checks {
			state := "PASS"
			if !c.OK {
				state = "FAIL"
			}
			fmt.Fprintf(out, "%s %s: %s\n", state, c.Name, c.Detail)
		}
	}
	if !ok {
		return errors.New("doctor found problems; review the checks above")
	}
	return nil
}
