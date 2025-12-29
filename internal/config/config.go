package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ResourceConfig, YAML'dan okunan ham kaynak tanımıdır.
// Factory bu yapıyı kullanarak gerçek Resource nesnelerini üretir.
type ResourceConfig struct {
	ID         string                 `yaml:"id"`
	Type       string                 `yaml:"type"`
	Parameters map[string]interface{} `yaml:"params"`
}

type Host struct {
	Name           string `yaml:"name"`
	HostName       string `yaml:"hostname"`
	User           string `yaml:"user"`
	Port           int    `yaml:"port"`
	SSHKeyPath     string `yaml:"ssh_key_path"`
	BecomePassword string `yaml:"become_password"` // Sudo şifresi (Age ile şifrelenecek)
}

type Config struct {
	Vars      map[string]string  `yaml:"vars"`
	Hosts     []Host             `yaml:"hosts"`
	Resources [][]ResourceConfig `yaml:"resources"` // Katmanlı yapı
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config dosyası okunamadı: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("yaml parse hatası: %w", err)
	}

	return &cfg, nil
}
