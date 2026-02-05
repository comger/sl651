package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sl651-platform/internal/simulator"
)

func main() {
	configPath := flag.String("config", "simulator_config.yaml", "path to simulator config file")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	cfg, err := simulator.LoadConfig(*configPath)
	if err != nil {
		log.Printf("[警告] Failed to load config from %s: %v. Creating default.", *configPath, err)
		cfg = simulator.DefaultConfig()
		if err := simulator.SaveConfig(*configPath, cfg); err != nil {
			log.Fatalf("[致命] Failed to save default config: %v", err)
		}
	}

	engine := simulator.NewEngine(cfg, *configPath)

	// Print logs to console
	engine.SetLogCallback(func(deviceID string, fCode byte, message string, data string) {
		log.Printf("[发送] Device %s -> %02X Data: %s", deviceID, fCode, data)
	})

	engine.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	engine.Stop()
}
