package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Resource struct {
	Type    string `yaml:"type"`
	Name    string `yaml:"name"`
	Path    string `yaml:"path,omitempty"`    // Dosya yolu
	Content string `yaml:"content,omitempty"` // Dosya içeriği
	State   string `yaml:"state,omitempty"`   // running/stopped (servisler için)
	Enabled bool   `yaml:"enabled,omitempty"`
}

type Host struct {
	Name     string `yaml:"name"`
	Address  string `yaml:"address"` // Örn: 192.168.1.100:22
	User     string `yaml:"user"`
	Password string `yaml:"password,omitempty"` // Şimdilik basit tutuyoruz
}

type Config struct {
	Resources []Resource `yaml:"resources"`
	Hosts     []Host     `yaml:"hosts,omitempty"` // Sunucu listesi
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
