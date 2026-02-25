package simulator

import (
	"os"

	"gopkg.in/yaml.v3"
)

type SimulatorConfig struct {
	ServerAddr string         `yaml:"server_addr"`
	Devices    []DeviceConfig `yaml:"devices"`
}

type DeviceConfig struct {
	ID                string          `yaml:"id"`
	Name              string          `yaml:"name"`
	Password          string          `yaml:"password"`
	Interval          int             `yaml:"interval_seconds"`
	HeartbeatEnabled  bool            `yaml:"heartbeat_enabled"`
	LoginEnabled      bool            `yaml:"login_enabled"`
	HourReportEnabled bool            `yaml:"hour_report_enabled"`
	Channels          []ChannelConfig `yaml:"channels"`
}

type ChannelConfig struct {
	Tag       byte             `yaml:"tag"`      // SL651 Tag (e.g., 0x39 for Water Level)
	Name      string           `yaml:"name"`     // Human readable name
	Decimals  int              `yaml:"decimals"` // Number of decimals
	Algorithm SimulationParams `yaml:"algorithm"`
}

type SimulationParams struct {
	Type  string  `yaml:"type"`  // "constant", "random", "sin", "step"
	Base  float64 `yaml:"base"`  // Base value
	Min   float64 `yaml:"min"`   // Min value for random/sin
	Max   float64 `yaml:"max"`   // Max value for random/sin
	Step  float64 `yaml:"step"`  // Step size for "step" algorithm
	Scale float64 `yaml:"scale"` // Scale factor for "sin" (period scaling)
}

func LoadConfig(path string) (*SimulatorConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config SimulatorConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func SaveConfig(path string, config *SimulatorConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func DefaultConfig() *SimulatorConfig {
	return &SimulatorConfig{
		ServerAddr: "127.0.0.1:8085",
		Devices: []DeviceConfig{
			{
				ID:                "1090330854",
				Name:              "标准水文站",
				Password:          "1234",
				Interval:          60,
				HeartbeatEnabled:  true,
				LoginEnabled:      true,
				HourReportEnabled: true,
				Channels: []ChannelConfig{
					{
						Tag:      0x39,
						Name:     "水位",
						Decimals: 3,
						Algorithm: SimulationParams{
							Type:  "sin",
							Base:  15.0,
							Min:   14.0,
							Max:   16.0,
							Scale: 3600,
						},
					},
					{
						Tag:      0x38,
						Name:     "电压",
						Decimals: 2,
						Algorithm: SimulationParams{
							Type: "random",
							Base: 12.5,
							Min:  12.0,
							Max:  13.0,
						},
					},
					{
						Tag:      0x22,
						Name:     "雨量",
						Decimals: 1,
						Algorithm: SimulationParams{
							Type: "step",
							Base: 0,
							Step: 0.5,
						},
					},
				},
			},
		},
	}
}
