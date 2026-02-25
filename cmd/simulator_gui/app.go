package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"sl651-platform/internal/simulator"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	engine     *simulator.Engine
	config     *simulator.SimulatorConfig
	configPath string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Load config from current directory or executable location
	exePath, _ := os.Executable()
	basePath := filepath.Dir(exePath)
	configName := "simulator_config.yaml"

	// Check current working dir first (for dev mode)
	if _, err := os.Stat(configName); err == nil {
		a.configPath, _ = filepath.Abs(configName)
	} else {
		a.configPath = filepath.Join(basePath, configName)
	}

	cfg, err := simulator.LoadConfig(a.configPath)
	if err != nil {
		log.Printf("[Wails] Failed to load config, using default: %v", err)
		cfg = simulator.DefaultConfig()
		_ = simulator.SaveConfig(a.configPath, cfg)
	}
	a.config = cfg
	a.engine = simulator.NewEngine(cfg, a.configPath)

	// Set log callback to emit Wails events
	a.engine.SetLogCallback(func(deviceID string, fCode byte, message string, data string) {
		runtime.EventsEmit(a.ctx, "simulator:log", map[string]interface{}{
			"device_id": deviceID,
			"fcode":     fmt.Sprintf("%02X", fCode),
			"message":   message,
			"data":      data,
			"time":      time.Now().Format("15:04:05"),
		})
	})
}

// GetConfig returns the current simulator configuration
func (a *App) GetConfig() *simulator.SimulatorConfig {
	return a.config
}

// SaveConfig saves the configuration to disk
func (a *App) SaveConfig(config *simulator.SimulatorConfig) error {
	a.config = config
	return simulator.SaveConfig(a.configPath, config)
}

// StartSimulation starts the simulator engine
func (a *App) StartSimulation() {
	if a.engine != nil {
		a.engine.Start()
	}
}

// StopSimulation stops the simulator engine
func (a *App) StopSimulation() {
	if a.engine != nil {
		a.engine.Stop()
	}
}

// AddDevice adds a new device to the configuration
func (a *App) AddDevice(name string) {
	newDev := simulator.DeviceConfig{
		ID:       fmt.Sprintf("109%07d", time.Now().Unix()%10000000),
		Name:     name,
		Interval: 60,
		Channels: []simulator.ChannelConfig{},
	}
	a.config.Devices = append(a.config.Devices, newDev)
	_ = simulator.SaveConfig(a.configPath, a.config)
}

// DeleteDevice removes a device from the configuration
func (a *App) DeleteDevice(index int) {
	if index >= 0 && index < len(a.config.Devices) {
		a.config.Devices = append(a.config.Devices[:index], a.config.Devices[index+1:]...)
		_ = simulator.SaveConfig(a.configPath, a.config)
	}
}
