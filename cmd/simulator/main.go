package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"sl651-platform/internal/sl651"
)

type Device struct {
	ID       string
	Name     string
	Interval time.Duration
	StopChan chan struct{}
	Running  bool
	mu       sync.Mutex
	Protocol *sl651.Protocol
	Step     int
	Mode     string // "normal", "solar_fail", "network_retry"
}

type Simulator struct {
	devices    []*Device
	serverAddr string
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewSimulator(serverAddr string) *Simulator {
	ctx, cancel := context.WithCancel(context.Background())
	return &Simulator{
		serverAddr: serverAddr,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (s *Simulator) AddDevice(id, name string, interval time.Duration, mode string) *Device {
	s.mu.Lock()
	defer s.mu.Unlock()

	device := &Device{
		ID:       id,
		Name:     name,
		Interval: interval,
		StopChan: make(chan struct{}),
		Protocol: sl651.NewProtocol(),
		Step:     0,
		Mode:     mode,
	}
	s.devices = append(s.devices, device)
	return device
}

func (s *Simulator) Start() {
	log.Printf("[日志] Starting simulator with %d devices on %s", len(s.devices), s.serverAddr)
	for _, device := range s.devices {
		go s.runDevice(device)
	}
}

func (s *Simulator) Stop() {
	log.Println("[日志] Stopping simulator...")
	s.cancel()
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, device := range s.devices {
		device.Stop()
	}
	log.Println("[日志] Simulator stopped")
}

func (d *Device) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.Running {
		d.Running = false
		close(d.StopChan)
	}
}

func (s *Simulator) runDevice(device *Device) {
	device.mu.Lock()
	device.Running = true
	device.mu.Unlock()

	log.Printf("[日志] Device %s (%s) started in [%s] mode", device.ID, device.Name, device.Mode)

	ticker := time.NewTicker(device.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-device.StopChan:
			log.Printf("[日志] Device %s stopped", device.ID)
			return
		case <-s.ctx.Done():
			log.Printf("[日志] Device %s stopped by context", device.ID)
			return
		case <-ticker.C:
			// Network Retry Logic: 1 in 5 chance to skip sending (simulating offline)
			if device.Mode == "network_retry" && rand.Intn(5) == 0 {
				log.Printf("[警告] Device %s (网络波动站) simulating connection interruption...", device.ID)
				time.Sleep(device.Interval * 2) // Wait longer
				log.Printf("[日志] Device %s (网络波动站) reconnecting and re-sending historical data...", device.ID)

				// Re-send with historical timestamp
				historicalTime := time.Now().Add(-1 * time.Hour)
				if err := s.sendDeviceDataWithTime(device, historicalTime); err != nil {
					log.Printf("[错误] Reissue failed: %v", err)
				}
			}

			if err := s.sendDeviceData(device); err != nil {
				log.Printf("[日志] Device %s failed to send data: %v", device.ID, err)
			}
		}
	}
}

func (s *Simulator) sendDeviceData(device *Device) error {
	return s.sendDeviceDataWithTime(device, time.Now())
}

func (s *Simulator) sendDeviceDataWithTime(device *Device, obsTime time.Time) error {
	conn, err := net.DialTimeout("tcp", s.serverAddr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Sequence of Function Codes
	fCodes := []byte{0x2F, 0x32, 0x31, 0x34}
	fCode := fCodes[device.Step%len(fCodes)]
	device.Step++

	body := make([]byte, 0)

	// Complex Telemetry Reports (31, 32, 34)
	if fCode != 0x2F {
		voltage := 12.5 + (rand.Float64() * 1.5)
		if device.Mode == "solar_fail" {
			voltage = 10.5 // Constant low voltage
		} else if device.Mode == "normal" {
			voltage = 13.2 // Very steady
		}

		waterLevel := 10.0 + (rand.Float64() * 20.0)
		if device.Mode == "normal" {
			waterLevel = 15.6
		}

		rainfall := rand.Float64() * 5.0

		// Add Observation Time for re-sending simulation
		body = append(body, device.Protocol.BuildTimeTLV(obsTime)...)

		body = append(body, device.Protocol.BuildTLV(0x22, rainfall, 1)...)
		body = append(body, device.Protocol.BuildTLV(0x27, 0.0, 3)...)
		body = append(body, device.Protocol.BuildTLV(0x39, waterLevel, 3)...)
		body = append(body, device.Protocol.BuildTLV(0x37, 0.0, 3)...)
		body = append(body, device.Protocol.BuildTLV(0x30, 43.0, 3)...)
		body = append(body, device.Protocol.BuildTLV(0x20, 0.0, 1)...)
		body = append(body, device.Protocol.BuildTLV(0x1F, 0.0, 1)...)
		body = append(body, device.Protocol.BuildTLV(0x26, 0.0, 1)...)
		body = append(body, device.Protocol.BuildTLV(0x38, voltage, 2)...)
	}

	data, err := device.Protocol.BuildMessage(device.ID, fCode, body)
	if err != nil {
		return err
	}

	hexData := hex.EncodeToString(data)
	fCodeName := device.Protocol.GetFunctionCodeName(fmt.Sprintf("%02X", fCode))

	// Parse back for logging as JSON
	parsed, _ := device.Protocol.Parse(hexData)
	if parsed != nil {
		jsonMap, _ := device.Protocol.ConvertToStandardData(parsed)
		jsonBytes, _ := json.Marshal(jsonMap)
		log.Printf("[日志] Device %s sending %s (%02X) [%s] JSON: %s", device.ID, fCodeName, fCode, device.Mode, string(jsonBytes))
	}

	_, err = conn.Write(data)
	return err
}

func main() {
	rand.Seed(time.Now().UnixNano())
	serverAddr := "127.0.0.1:8080"

	simulator := NewSimulator(serverAddr)
	simulator.AddDevice("1090330853", "演示-完全正常", 10*time.Second, "normal")
	simulator.AddDevice("1090330854", "演示-供电异常", 10*time.Second, "solar_fail")
	simulator.AddDevice("1090330855", "演示-补发测试", 10*time.Second, "network_retry")
	simulator.AddDevice("1090330856", "常规站点", 15*time.Second, "normal")

	simulator.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	simulator.Stop()
}
