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

func (s *Simulator) AddDevice(id, name string, interval time.Duration) *Device {
	s.mu.Lock()
	defer s.mu.Unlock()

	device := &Device{
		ID:       id,
		Name:     name,
		Interval: interval,
		StopChan: make(chan struct{}),
		Protocol: sl651.NewProtocol(),
		Step:     0,
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

	log.Printf("[日志] Device %s (%s) started", device.ID, device.Name)

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
			if err := s.sendDeviceData(device); err != nil {
				log.Printf("[日志] Device %s failed to send data: %v", device.ID, err)
			}
		}
	}
}

func (s *Simulator) sendDeviceData(device *Device) error {
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
		// Aligned with reference meta and decimals
		body = append(body, device.Protocol.BuildTLV(0x22, 12.2, 1)...) // Rainfall: meta 19 (3 byte), dec 1
		body = append(body, device.Protocol.BuildTLV(0x27, 0.0, 3)...)  // Day Rainfall: meta 2b (5 byte), dec 3
		body = append(body, device.Protocol.BuildTLV(0x39, 25.2, 3)...) // Water Level: meta 23 (4 byte), dec 3
		body = append(body, device.Protocol.BuildTLV(0x37, 0.0, 3)...)  // Instant Flow: meta 1b (3 byte), dec 3
		body = append(body, device.Protocol.BuildTLV(0x30, 43.0, 3)...) // Cumulative Flow: meta 2b (5 byte), dec 3
		body = append(body, device.Protocol.BuildTLV(0x20, 0.0, 1)...)  // Total Rainfall: meta 19 (3 byte), dec 1
		body = append(body, device.Protocol.BuildTLV(0x1F, 0.0, 1)...)  // Period Rainfall: meta 19 (3 byte), dec 1
		body = append(body, device.Protocol.BuildTLV(0x26, 0.0, 1)...)  // Hourly Rainfall: meta 19 (3 byte), dec 1
		body = append(body, device.Protocol.BuildTLV(0x38, 12.5, 2)...) // Voltage: meta 12 (2 byte), dec 2
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
		log.Printf("[日志] Device %s sending %s (%02X) JSON: %s", device.ID, fCodeName, fCode, string(jsonBytes))
	} else {
		log.Printf("[日志] Device %s sending %s (%02X) HEX: %s", device.ID, fCodeName, fCode, hexData)
	}

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	if fCode == 0x2F {
		log.Printf("[日志] Device %s (2F): Mode M1, skipping response check.", device.ID)
		return nil
	}

	response := make([]byte, 1024)
	err = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return err
	}

	n, err := conn.Read(response)
	if err != nil {
		log.Printf("[日志] Device %s wait response timeout/error: %v", device.ID, err)
	} else {
		respHex := hex.EncodeToString(response[:n])
		parsedResp, _ := device.Protocol.Parse(respHex)
		if parsedResp != nil {
			jsonMap, _ := device.Protocol.ConvertToStandardData(parsedResp)
			jsonBytes, _ := json.Marshal(jsonMap)
			log.Printf("[日志] Device %s received response JSON: %s", device.ID, string(jsonBytes))
		} else {
			log.Printf("[日志] Device %s received response HEX: %x", device.ID, response[:n])
		}
	}

	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	serverAddr := "120.79.72.98:9100"
	serverAddr = "127.0.0.1:8080"

	if len(os.Args) > 1 {
		serverAddr = os.Args[1]
	}

	simulator := NewSimulator(serverAddr)
	simulator.AddDevice("1090330854", "演示站点2", 5*time.Second)
	simulator.AddDevice("1090330855", "演示站点2", 5*time.Second)
	simulator.AddDevice("1090330856", "演示站点2", 5*time.Second)

	simulator.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	simulator.Stop()
}
