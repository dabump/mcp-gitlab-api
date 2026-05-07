package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	GitLab GitLabConfig `yaml:"gitlab"`
}

type ServerConfig struct {
	Transport string `yaml:"transport"`
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Endpoint  string `yaml:"endpoint"`
}

type GitLabConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

func Load(path string) (Config, error) {
	cfg := Config{
		Server: ServerConfig{
			Transport: "stdio",
			Host:      "127.0.0.1",
			Port:      8080,
			Endpoint:  "/mcp",
		},
		GitLab: GitLabConfig{URL: "https://gitlab.com"},
	}

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("parse config: %w", err)
		}
	}

	if v := strings.TrimSpace(os.Getenv("GITLAB_URL")); v != "" {
		cfg.GitLab.URL = v
	}
	if v := strings.TrimSpace(os.Getenv("GITLAB_TOKEN")); v != "" {
		cfg.GitLab.Token = v
	}

	cfg.GitLab.URL = strings.TrimRight(strings.TrimSpace(cfg.GitLab.URL), "/")
	cfg.GitLab.Token = strings.TrimSpace(cfg.GitLab.Token)
	cfg.Server.Transport = strings.ToLower(strings.TrimSpace(cfg.Server.Transport))
	cfg.Server.Host = strings.TrimSpace(cfg.Server.Host)
	cfg.Server.Endpoint = strings.TrimSpace(cfg.Server.Endpoint)

	if cfg.Server.Transport == "" {
		cfg.Server.Transport = "stdio"
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "127.0.0.1"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Endpoint == "" {
		cfg.Server.Endpoint = "/mcp"
	}

	if cfg.GitLab.URL == "" {
		return Config{}, errors.New("gitlab.url is required")
	}
	if cfg.GitLab.Token == "" {
		return Config{}, errors.New("gitlab.token is required")
	}
	if cfg.Server.Transport != "stdio" && cfg.Server.Transport != "http" {
		return Config{}, errors.New("server.transport must be stdio or http")
	}
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return Config{}, errors.New("server.port must be between 1 and 65535")
	}

	return cfg, nil
}
