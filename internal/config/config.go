// Package config reads myLlama's process-level configuration.
package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"myLlama/internal/api"
	"myLlama/internal/runner"
)

const defaultListenAddress = "127.0.0.1:8080"

// Config contains settings needed before a model runner can be started.
type Config struct {
	ListenAddress string
	Runner        runner.Config
	API           api.Config
	Profiles      []runner.Profile
	LogPath       string
	LogMaxBytes   int64
	LogBackups    int
}

// FromEnv loads safe defaults for the local HTTP API.
func FromEnv() (Config, error) {
	listenAddress := os.Getenv("MYLLAMA_LISTEN_ADDRESS")
	if listenAddress == "" {
		listenAddress = defaultListenAddress
	}

	host, _, err := net.SplitHostPort(listenAddress)
	if err != nil {
		return Config{}, fmt.Errorf("MYLLAMA_LISTEN_ADDRESS must be host:port: %w", err)
	}
	if host != "127.0.0.1" && host != "localhost" {
		return Config{}, fmt.Errorf("MYLLAMA_LISTEN_ADDRESS must bind to localhost")
	}

	runnerConfig := runner.DefaultConfig()
	apiConfig := api.DefaultConfig()
	runnerConfig.Executable = envOrDefault("MYLLAMA_LLAMA_SERVER_PATH", runnerConfig.Executable)
	runnerConfig.Address = envOrDefault("MYLLAMA_RUNNER_ADDRESS", runnerConfig.Address)
	if err := validateLoopbackAddress("MYLLAMA_RUNNER_ADDRESS", runnerConfig.Address); err != nil {
		return Config{}, err
	}
	if runnerConfig.ContextSize, err = positiveIntEnv("MYLLAMA_CONTEXT_SIZE", runnerConfig.ContextSize); err != nil {
		return Config{}, err
	}
	if runnerConfig.GPULayers, err = nonNegativeIntEnv("MYLLAMA_GPU_LAYERS", runnerConfig.GPULayers); err != nil {
		return Config{}, err
	}
	if runnerConfig.ReadinessTimeout, err = durationEnv("MYLLAMA_READINESS_TIMEOUT", runnerConfig.ReadinessTimeout); err != nil {
		return Config{}, err
	}
	if runnerConfig.ShutdownTimeout, err = durationEnv("MYLLAMA_SHUTDOWN_TIMEOUT", runnerConfig.ShutdownTimeout); err != nil {
		return Config{}, err
	}
	if runnerConfig.AutoRestart, err = boolEnv("MYLLAMA_AUTO_RESTART", runnerConfig.AutoRestart); err != nil {
		return Config{}, err
	}
	if runnerConfig.AutoRestartDelay, err = durationEnv("MYLLAMA_AUTO_RESTART_DELAY", runnerConfig.AutoRestartDelay); err != nil {
		return Config{}, err
	}
	if runnerConfig.MaxAutoRestarts, err = nonNegativeIntEnv("MYLLAMA_MAX_AUTO_RESTARTS", runnerConfig.MaxAutoRestarts); err != nil {
		return Config{}, err
	}
	if apiConfig.MaxRequestBytes, err = positiveInt64Env("MYLLAMA_MAX_REQUEST_BYTES", apiConfig.MaxRequestBytes); err != nil {
		return Config{}, err
	}
	if apiConfig.MaxInputBytes, err = positiveIntEnv("MYLLAMA_MAX_INPUT_BYTES", apiConfig.MaxInputBytes); err != nil {
		return Config{}, err
	}
	if apiConfig.MaxGeneratedTokens, err = positiveIntEnv("MYLLAMA_MAX_GENERATED_TOKENS", apiConfig.MaxGeneratedTokens); err != nil {
		return Config{}, err
	}
	if apiConfig.DefaultMaxTokens, err = positiveIntEnv("MYLLAMA_DEFAULT_MAX_TOKENS", apiConfig.DefaultMaxTokens); err != nil {
		return Config{}, err
	}
	if apiConfig.QueueCapacity, err = positiveIntEnv("MYLLAMA_QUEUE_CAPACITY", apiConfig.QueueCapacity); err != nil {
		return Config{}, err
	}
	if apiConfig.RequestTimeout, err = durationEnv("MYLLAMA_REQUEST_TIMEOUT", apiConfig.RequestTimeout); err != nil {
		return Config{}, err
	}
	if err := apiConfig.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid API limits: %w", err)
	}

	logMaxBytes, err := positiveInt64Env("MYLLAMA_LOG_MAX_BYTES", 10<<20)
	if err != nil {
		return Config{}, err
	}
	logBackups, err := positiveIntEnv("MYLLAMA_LOG_BACKUPS", 3)
	if err != nil {
		return Config{}, err
	}
	if logBackups > 100 {
		return Config{}, fmt.Errorf("MYLLAMA_LOG_BACKUPS must not exceed 100")
	}
	profiles, err := profilesFromEnv()
	if err != nil {
		return Config{}, err
	}
	for _, profile := range profiles {
		maxTokens := apiConfig.MaxGeneratedTokens
		if profile.MaxGeneratedTokens > 0 {
			maxTokens = profile.MaxGeneratedTokens
		}
		defaultTokens := apiConfig.DefaultMaxTokens
		if profile.DefaultMaxTokens > 0 {
			defaultTokens = profile.DefaultMaxTokens
		}
		if defaultTokens > maxTokens {
			return Config{}, fmt.Errorf("profile %q default max tokens must not exceed max generated tokens", profile.Name)
		}
	}
	return Config{ListenAddress: listenAddress, Runner: runnerConfig, API: apiConfig, Profiles: profiles, LogPath: os.Getenv("MYLLAMA_LOG_PATH"), LogMaxBytes: logMaxBytes, LogBackups: logBackups}, nil
}

func profilesFromEnv() ([]runner.Profile, error) {
	path := os.Getenv("MYLLAMA_PROFILES_PATH")
	if path == "" {
		return nil, nil
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read MYLLAMA_PROFILES_PATH: %w", err)
	}
	var document struct {
		Profiles map[string]runner.Profile `json:"profiles"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(contents)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("parse MYLLAMA_PROFILES_PATH: %w", err)
	}
	profiles := make([]runner.Profile, 0, len(document.Profiles))
	for name, profile := range document.Profiles {
		profile.Name = name
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func validateLoopbackAddress(key, address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil || port == "" {
		return fmt.Errorf("%s must be host:port", key)
	}
	if host != "127.0.0.1" && host != "localhost" {
		return fmt.Errorf("%s must bind to localhost", key)
	}
	return nil
}

func positiveIntEnv(key string, fallback int) (int, error) {
	value, err := nonNegativeIntEnv(key, fallback)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}
	return value, nil
}

func nonNegativeIntEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", key)
	}
	return parsed, nil
}

func positiveInt64Env(key string, fallback int64) (int64, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return parsed, nil
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive Go duration", key)
	}
	return parsed, nil
}

func boolEnv(key string, fallback bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false", key)
	}
	return parsed, nil
}
