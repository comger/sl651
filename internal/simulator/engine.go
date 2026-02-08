package simulator

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net"
	"sl651-platform/internal/sl651"
	"time"
)

type LogCallback func(deviceID string, fCode byte, message string, data string)

type Engine struct {
	cfg         *Config
	configPath  string
	protocol    *sl651.Protocol
	logCallback LogCallback
	cancel      context.CancelFunc
}

func NewEngine(cfg *Config, configPath string) *Engine {
	return &Engine{
		cfg:        cfg,
		configPath: configPath,
		protocol:   sl651.NewProtocol(),
	}
}

func (e *Engine) SetLogCallback(cb LogCallback) {
	e.logCallback = cb
}

func (e *Engine) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel

	for _, dev := range e.cfg.Devices {
		go e.runDevice(ctx, dev)
	}
}

func (e *Engine) Stop() {
	if e.cancel != nil {
		e.cancel()
	}
}

func (e *Engine) runDevice(ctx context.Context, dev DeviceConfig) {
	for _, report := range dev.Reports {
		go e.runReport(ctx, dev, report)
	}
}

func (e *Engine) runReport(ctx context.Context, dev DeviceConfig, report ReportConfig) {
	ticker := time.NewTicker(time.Duration(report.Interval) * time.Second)
	defer ticker.Stop()

	fCode := e.getFunctionCode(report.Type)

	// Initial send
	for _, center := range e.cfg.Centers {
		e.sendToCenter(dev, center, fCode)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, center := range e.cfg.Centers {
				e.sendToCenter(dev, center, fCode)
			}
		}
	}
}

func (e *Engine) getFunctionCode(reportType string) byte {
	switch reportType {
	case "login":
		return 0x2F
	case "hourly":
		return 0x34
	case "periodic":
		return 0x31
	case "scheduled":
		return 0x32
	default:
		return 0x32 // default to add report
	}
}

func (e *Engine) sendToCenter(dev DeviceConfig, center BusinessCenterConfig, fCode byte) {
	conn, err := net.DialTimeout("tcp", center.Addr, 5*time.Second)
	if err != nil {
		if e.logCallback != nil {
			e.logCallback(dev.ID, fCode, fmt.Sprintf("Connection to %s failed", center.Addr), err.Error())
		}
		return
	}
	defer conn.Close()

	body := make([]byte, 0)
	if fCode != 0x2F { // Login/Heartbeat usually doesn't have telemetry body in standard
		for _, el := range dev.Elements {
			val := e.generateValue(el)
			body = append(body, e.protocol.BuildTLV(el.Tag, val, el.Decimals)...)
		}
	}

	frame, err := e.protocol.BuildMessage(dev.ID, dev.Password, center.ID, fCode, body)
	if err != nil {
		if e.logCallback != nil {
			e.logCallback(dev.ID, fCode, "Build message failed", err.Error())
		}
		return
	}

	_, err = conn.Write(frame)
	if err != nil {
		if e.logCallback != nil {
			e.logCallback(dev.ID, fCode, "Send failed", err.Error())
		}
		return
	}

	if e.logCallback != nil {
		e.logCallback(dev.ID, fCode, "Sent success", fmt.Sprintf("%X", frame))
	}
}

func (e *Engine) generateValue(el ElementConfig) float64 {
	switch el.Strategy {
	case "random":
		return el.BaseLine + (rand.Float64()*2 - 1) // +/- 1.0
	case "sine":
		return el.BaseLine + 2*math.Sin(float64(time.Now().Unix()%3600)/60.0) // period 1 min
	case "incremental":
		return el.BaseLine + float64(time.Now().Unix()%100)
	case "static":
		return el.BaseLine
	default:
		return el.BaseLine
	}
}
