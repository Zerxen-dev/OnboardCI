package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	userRegex   = regexp.MustCompile(`^[a-zA-Z0-9._-]+(:[a-zA-Z0-9._-]+)?$`)
	cpusRegex   = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?$`)
	memoryRegex = regexp.MustCompile(`^[0-9]+[bkmgBKMG]?$`)
)

// DockerConfig holds sandbox execution settings.
type DockerConfig struct {
	Image   string        `yaml:"image"`
	Timeout time.Duration `yaml:"timeout"`
	User    string        `yaml:"user"`
	CPUs    string        `yaml:"cpus"`
	Memory  string        `yaml:"memory"`
}

// Config represents .onboardci.yml definition.
type Config struct {
	Setup  []string     `yaml:"setup"`
	Test   []string     `yaml:"test"`
	Env    []string     `yaml:"env"`
	Docker DockerConfig `yaml:"docker"`
}

// rawDockerConfig helps unmarshal custom string timeouts (e.g. "10m").
type rawDockerConfig struct {
	Image   string `yaml:"image"`
	Timeout string `yaml:"timeout"`
	User    string `yaml:"user"`
	CPUs    string `yaml:"cpus"`
	Memory  string `yaml:"memory"`
}

type rawConfig struct {
	Setup  []string        `yaml:"setup"`
	Test   []string        `yaml:"test"`
	Env    []string        `yaml:"env"`
	Docker rawDockerConfig `yaml:"docker"`
}

// DefaultConfig returns safe baseline configuration values.
func DefaultConfig() *Config {
	return &Config{
		Setup: []string{},
		Test:  []string{},
		Env:   []string{},
		Docker: DockerConfig{
			Image:   "node:20-alpine",
			Timeout: 10 * time.Minute,
			User:    "1000:1000",
			CPUs:    "2.0",
			Memory:  "2g",
		},
	}
}

// LoadConfig attempts to locate and parse .onboardci.yml or a custom config path.
// If no file exists, it returns a default configuration with an empty path.
func LoadConfig(cwd string, customPath string) (*Config, string, error) {
	cfg := DefaultConfig()

	var targetPath string
	if customPath != "" {
		if filepath.IsAbs(customPath) {
			targetPath = customPath
		} else {
			targetPath = filepath.Join(cwd, customPath)
		}
	} else {
		candidates := []string{
			filepath.Join(cwd, ".onboardci.yml"),
			filepath.Join(cwd, ".onboardci.yaml"),
		}
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				targetPath = candidate
				break
			}
		}
	}

	if targetPath == "" {
		return cfg, "", nil
	}

	data, err := os.ReadFile(targetPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read config file %s: %w", targetPath, err)
	}

	var raw rawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, "", fmt.Errorf("malformed yaml in %s: %w", targetPath, err)
	}

	cfg.Setup = raw.Setup
	cfg.Test = raw.Test
	cfg.Env = raw.Env

	if raw.Docker.Image != "" {
		cfg.Docker.Image = raw.Docker.Image
	}
	if raw.Docker.User != "" {
		cfg.Docker.User = raw.Docker.User
	}
	if raw.Docker.CPUs != "" {
		cfg.Docker.CPUs = raw.Docker.CPUs
	}
	if raw.Docker.Memory != "" {
		cfg.Docker.Memory = raw.Docker.Memory
	}
	if raw.Docker.Timeout != "" {
		parsedTimeout, err := time.ParseDuration(raw.Docker.Timeout)
		if err != nil {
			return nil, "", fmt.Errorf("invalid docker timeout duration %q: %w", raw.Docker.Timeout, err)
		}
		cfg.Docker.Timeout = parsedTimeout
	}

	if err := validateConfig(cfg); err != nil {
		return nil, "", fmt.Errorf("invalid configuration in %s: %w", targetPath, err)
	}

	return cfg, targetPath, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Docker.User != "" && !userRegex.MatchString(cfg.Docker.User) {
		return fmt.Errorf("invalid docker user %q: must match [user[:group]] format without shell metacharacters", cfg.Docker.User)
	}
	if cfg.Docker.CPUs != "" && !cpusRegex.MatchString(cfg.Docker.CPUs) {
		return fmt.Errorf("invalid docker cpus %q: must be a numeric value (e.g. 2.0)", cfg.Docker.CPUs)
	}
	if cfg.Docker.Memory != "" && !memoryRegex.MatchString(cfg.Docker.Memory) {
		return fmt.Errorf("invalid docker memory %q: must be a valid memory limit (e.g. 2g, 512m)", cfg.Docker.Memory)
	}
	if cfg.Docker.Image != "" {
		if strings.ContainsAny(cfg.Docker.Image, " \t\n\r;$\"&|<>'`\\") {
			return fmt.Errorf("invalid docker image %q: contains illegal characters", cfg.Docker.Image)
		}
	}
	return nil
}
