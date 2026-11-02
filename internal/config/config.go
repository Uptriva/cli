package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Token   string `yaml:"token"`
	BaseURL string `yaml:"base_url"`
}

func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "uptriva", "config.yaml"), nil
}

func Load() (*Config, error) {
	c := &Config{BaseURL: "https://api.uptriva.com/api/v1/uptime"}
	p, err := Path()
	if err != nil {
		return c, nil
	}
	b, err := os.ReadFile(p)
	if err == nil {
		_ = yaml.Unmarshal(b, c)
	}
	if t := os.Getenv("UPTRIVA_TOKEN"); t != "" {
		c.Token = t
	}
	if u := os.Getenv("UPTRIVA_API_URL"); u != "" {
		c.BaseURL = u
	}
	return c, nil
}

func (c *Config) Save() error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o600)
}
