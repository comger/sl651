package diagnosis

import (
	"context"
	"fmt"
	"sl651-platform/internal/model"
	"sl651-platform/internal/storage"
	"time"
)

type Manager struct {
	storage *storage.Storage
}

func NewManager(storage *storage.Storage) *Manager {
	return &Manager{
		storage: storage,
	}
}

func (m *Manager) HandleCommError(ctx context.Context, deviceID string, faultType model.FaultType, severity model.Severity, msg string, details string) error {
	return m.HandleFault(ctx, deviceID, "F-T-01-01", faultType, severity, msg, details)
}

func (m *Manager) HandleFault(ctx context.Context, deviceID string, faultCode string, faultType model.FaultType, severity model.Severity, msg string, details string) error {
	log := &model.FaultLog{
		DeviceID:  deviceID,
		FaultCode: faultCode,
		Type:      faultType,
		Severity:  severity,
		Message:   msg,
		Details:   details,
		Time:      time.Now(),
	}
	return m.storage.SaveFaultLog(ctx, log)
}

func (m *Manager) AnalyzeData(ctx context.Context, deviceID string, data map[string]interface{}) error {
	// 1. Solar Power System (Category P)
	if val, ok := data["voltage"].(float64); ok {
		if val < 11.0 {
			m.HandleFault(ctx, deviceID, "F-P-01-01", model.FaultTypeHardware, model.SeverityError, "Battery Undervoltage", fmt.Sprintf("Voltage: %.2fV (Threshold < 11V)", val))
		} else if val > 14.8 {
			m.HandleFault(ctx, deviceID, "F-P-01-02", model.FaultTypeHardware, model.SeverityWarning, "Battery Overvoltage", fmt.Sprintf("Voltage: %.2fV (Threshold > 14.8V)", val))
		}
	}

	// 2. Perception System (Category S)
	if val, ok := data["water_level"].(float64); ok {
		if val < 0 || val > 50.0 { // Assuming 50m is physical limit for this site type
			m.HandleFault(ctx, deviceID, "F-S-01-01", model.FaultTypeData, model.SeverityError, "Water Level Out of Range", fmt.Sprintf("Level: %.2f (Limit: 0-50m)", val))
		}
	}

	if val, ok := data["rainfall"].(float64); ok {
		if val < 0 {
			m.HandleFault(ctx, deviceID, "F-S-02-01", model.FaultTypeData, model.SeverityError, "Rainfall Anomaly", fmt.Sprintf("Value: %.2f (Cannot be negative)", val))
		}
	}

	return nil
}

func (m *Manager) LogSystemEvent(ctx context.Context, level, source, message string) error {
	log := &model.SystemLog{
		Level:   level,
		Source:  source,
		Message: message,
		Time:    time.Now(),
	}
	return m.storage.SaveSystemLog(ctx, log)
}

func (m *Manager) GetFaultLogs(ctx context.Context, deviceID string, limit int) ([]*model.FaultLog, error) {
	return m.storage.GetFaultLogs(ctx, deviceID, limit)
}

func (m *Manager) GetSystemLogs(ctx context.Context, limit int) ([]*model.SystemLog, error) {
	return m.storage.GetSystemLogs(ctx, limit)
}
