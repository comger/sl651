package heartbeat

import (
	"context"
	"log"
	"time"

	"sl651-platform/internal/config"
	"sl651-platform/internal/device"
	"sl651-platform/internal/model"
	"sl651-platform/internal/notifier"
)

type HeartbeatManager struct {
	config        config.HeartbeatConfig
	deviceManager *device.Manager
	notifier      *notifier.Notifier
	interval      time.Duration
	timeout       time.Duration
}

func NewHeartbeatManager(cfg config.HeartbeatConfig, dm *device.Manager, n *notifier.Notifier) *HeartbeatManager {
	return &HeartbeatManager{
		config:        cfg,
		deviceManager: dm,
		notifier:      n,
		interval:      time.Duration(cfg.CheckInterval) * time.Second,
		timeout:       time.Duration(cfg.DefaultTimeout) * time.Second,
	}
}

func (hm *HeartbeatManager) Start(ctx context.Context) {
	log.Printf("Heartbeat manager started (interval: %v, timeout: %v)", hm.interval, hm.timeout)

	ticker := time.NewTicker(hm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Heartbeat manager stopped")
			return
		case <-ticker.C:
			hm.checkHeartbeats(ctx)
		}
	}
}

func (hm *HeartbeatManager) checkHeartbeats(ctx context.Context) {
	devices, err := hm.deviceManager.ListDevices(ctx)
	if err != nil {
		log.Printf("Heartbeat check failed to list devices: %v", err)
		return
	}

	now := time.Now()
	for _, dev := range devices {
		// Only check online devices
		if dev.Status != model.StatusOnline {
			continue
		}

		if now.Sub(dev.LastSeen) > hm.timeout {
			hm.handleDeviceOffline(ctx, dev)
		}
	}
}

func (hm *HeartbeatManager) handleDeviceOffline(ctx context.Context, dev *model.Device) {
	log.Printf("Device %s heartbeat timeout, marking as offline", dev.ID)

	dev.Status = model.StatusOffline
	if err := hm.deviceManager.UpdateDevice(ctx, dev); err != nil {
		log.Printf("Failed to update device %s status to offline: %v", dev.ID, err)
		return
	}

	alert := &notifier.Alert{
		Type:    notifier.AlertTypeOffline,
		Device:  dev.ID,
		Message: "Device offline due to heartbeat timeout",
	}
	hm.notifier.SendAlert(ctx, alert)
}

func (hm *HeartbeatManager) GetConfig() config.HeartbeatConfig {
	return hm.config
}

func (hm *HeartbeatManager) UpdateConfig(cfg config.HeartbeatConfig) {
	hm.config = cfg
	hm.interval = time.Duration(cfg.CheckInterval) * time.Second
	hm.timeout = time.Duration(cfg.DefaultTimeout) * time.Second
	log.Printf("Heartbeat manager config updated (interval: %v, timeout: %v)", hm.interval, hm.timeout)
}
