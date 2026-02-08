package simulator

import (
	"os"

	"gopkg.in/yaml.v3"
)

type BusinessCenterConfig struct {
	Addr string `yaml:"addr"`
	ID   byte   `yaml:"id"`
}

type Config struct {
	Centers []BusinessCenterConfig `yaml:"centers"`
	Devices []DeviceConfig         `yaml:"devices"`
}

type ReportConfig struct {
	Type     string `yaml:"type"`     // "login", "hourly", "periodic"
	Interval int    `yaml:"interval"` // seconds
}

type ElementConfig struct {
	Tag      byte    `yaml:"tag"`
	Name     string  `yaml:"name"`
	BaseLine float64 `yaml:"baseline"`
	Strategy string  `yaml:"strategy"` // "random", "sine", "incremental", "static"
	Decimals int     `yaml:"decimals"`
}

type DeviceConfig struct {
	ID       string          `yaml:"id"`
	Password string          `yaml:"password"`
	Reports  []ReportConfig  `yaml:"reports"`
	Elements []ElementConfig `yaml:"elements"`
}

func DefaultConfig() *Config {
	return &Config{
		Centers: []BusinessCenterConfig{
			{Addr: "localhost:8080", ID: 0x01},
		},
		Devices: []DeviceConfig{
			{
				ID:       "1090330853",
				Password: "1234",
				Reports: []ReportConfig{
					{Type: "login", Interval: 60},
					{Type: "periodic", Interval: 30},
					{Type: "scheduled", Interval: 300},
					{Type: "hourly", Interval: 3600},
				},
				Elements: []ElementConfig{
					{Tag: 0x39, Name: "water_level_1", BaseLine: 10.0, Strategy: "sine", Decimals: 3},
					{Tag: 0x38, Name: "voltage", BaseLine: 12.5, Strategy: "random", Decimals: 2},
				},
			},
		},
	}
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

func SaveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
