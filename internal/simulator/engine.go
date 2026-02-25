package simulator

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"

	"sl651-platform/internal/sl651"
)

type LogCallback func(deviceID string, fCode byte, message string, data string)

type Device struct {
	Config   DeviceConfig
	StopChan chan struct{}
	Running  bool
	mu       sync.Mutex
	Protocol *sl651.Protocol
	Step     int
}

type Engine struct {
	config      *SimulatorConfig
	configPath  string
	devices     []*Device
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	logCallback LogCallback
}

func NewEngine(config *SimulatorConfig, path string) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	e := &Engine{
		config:     config,
		configPath: path,
		ctx:        ctx,
		cancel:     cancel,
	}

	for _, dc := range config.Devices {
		e.devices = append(e.devices, &Device{
			Config:   dc,
			StopChan: make(chan struct{}),
			Protocol: sl651.NewProtocol(),
		})
	}
	return e
}

func (e *Engine) SetLogCallback(cb LogCallback) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.logCallback = cb
}

func (e *Engine) Start() {
	log.Printf("[Engine] Starting with %d devices on %s", len(e.devices), e.config.ServerAddr)
	for _, device := range e.devices {
		go e.runDevice(device)
	}
}

func (e *Engine) Stop() {
	log.Println("[Engine] Stopping and saving state...")
	e.cancel()
	e.mu.RLock()
	defer e.mu.RUnlock()

	for i, d := range e.devices {
		for j, ch := range d.Config.Channels {
			if ch.Algorithm.Type == "step" {
				e.config.Devices[i].Channels[j].Algorithm.Base = ch.Algorithm.Base
			}
		}
	}

	if err := SaveConfig(e.configPath, e.config); err != nil {
		log.Printf("[Engine] Failed to save config: %v", err)
	}

	for _, device := range e.devices {
		device.Stop()
	}
	log.Println("[Engine] Stopped and state saved")
}

func (d *Device) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.Running {
		d.Running = false
		close(d.StopChan)
	}
}

func (e *Engine) runDevice(device *Device) {
	device.mu.Lock()
	device.Running = true
	device.mu.Unlock()

	log.Printf("[Engine] Device %s (%s) started", device.Config.ID, device.Config.Name)

	if device.Config.LoginEnabled {
		if err := e.sendRawMessage(device, 0x2F, nil); err != nil {
			log.Printf("[Engine] Login report failed for %s: %v", device.Config.ID, err)
		}
	}

	interval := time.Duration(device.Config.Interval) * time.Second
	if interval < time.Second {
		interval = time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Hour ticker
	hourTicker := time.NewTicker(time.Minute)
	defer hourTicker.Stop()
	lastHour := time.Now().Hour()

	for {
		select {
		case <-device.StopChan:
			return
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			if err := e.sendPeriodicData(device); err != nil {
				log.Printf("[Engine] Device %s periodic send failed: %v", device.Config.ID, err)
			}
		case now := <-hourTicker.C:
			if device.Config.HourReportEnabled && now.Hour() != lastHour {
				lastHour = now.Hour()
				if err := e.sendRawMessage(device, 0x34, e.generateBody(device, now)); err != nil {
					log.Printf("[Engine] Hour report failed for %s: %v", device.Config.ID, err)
				}
			}
		}
	}
}

func (e *Engine) generateBody(device *Device, obsTime time.Time) []byte {
	body := make([]byte, 0)
	body = append(body, device.Protocol.BuildTimeTLV(obsTime)...)

	for i := range device.Config.Channels {
		ch := &device.Config.Channels[i]
		val := generateValue(ch.Algorithm, obsTime)

		if ch.Algorithm.Type == "step" {
			ch.Algorithm.Base = val
		}

		body = append(body, device.Protocol.BuildTLV(ch.Tag, val, ch.Decimals)...)
	}
	return body
}

func generateValue(p SimulationParams, t time.Time) float64 {
	switch p.Type {
	case "constant":
		return p.Base
	case "random":
		return p.Min + rand.Float64()*(p.Max-p.Min)
	case "sin":
		period := p.Scale
		if period == 0 {
			period = 86400
		}
		seconds := float64(t.Hour()*3600 + t.Minute()*60 + t.Second())
		phase := (seconds / period) * 2 * math.Pi
		normalizedSin := (math.Sin(phase) + 1) / 2
		return p.Min + normalizedSin*(p.Max-p.Min)
	case "step":
		return p.Base + p.Step
	default:
		return p.Base
	}
}

func (e *Engine) sendPeriodicData(device *Device) error {
	body := e.generateBody(device, time.Now())
	return e.sendRawMessage(device, 0x32, body)
}

func (e *Engine) sendRawMessage(device *Device, fCode byte, body []byte) error {
	conn, err := net.DialTimeout("tcp", e.config.ServerAddr, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	data, err := device.Protocol.BuildMessage(device.Config.ID, fCode, body)
	if err != nil {
		return err
	}

	hexData := hex.EncodeToString(data)
	parsed, _ := device.Protocol.Parse(hexData)
	var jsonData string
	if parsed != nil {
		jsonMap, _ := device.Protocol.ConvertToStandardData(parsed)
		jb, _ := json.Marshal(jsonMap)
		jsonData = string(jb)
	}

	e.mu.RLock()
	cb := e.logCallback
	e.mu.RUnlock()

	if cb != nil {
		cb(device.Config.ID, fCode, "Sending message", jsonData)
	}

	_, err = conn.Write(data)
	return err
}
