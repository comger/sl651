package storage

import (
	"context"
	"fmt"
	"time"

	"sl651-platform/internal/model"

	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

func NewStorage(db *gorm.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) SaveDevice(ctx context.Context, device *model.Device) error {
	return s.db.WithContext(ctx).Save(device).Error
}

func (s *Storage) GetDevice(ctx context.Context, id string) (*model.Device, error) {
	var device model.Device
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (s *Storage) ListDevices(ctx context.Context) ([]*model.Device, error) {
	var devices []*model.Device
	err := s.db.WithContext(ctx).Find(&devices).Error
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (s *Storage) UpdateDevice(ctx context.Context, device *model.Device) error {
	return s.db.WithContext(ctx).Save(device).Error
}

func (s *Storage) DeleteDevice(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&model.Device{}, "id = ?", id).Error
}

func (s *Storage) SaveDeviceData(ctx context.Context, data *model.DeviceData) error {
	return s.db.WithContext(ctx).Save(data).Error
}

func (s *Storage) GetDeviceData(ctx context.Context, deviceID string, limit int, offset int) ([]*model.DeviceData, error) {
	var data []*model.DeviceData
	query := s.db.WithContext(ctx).Where("device_id = ?", deviceID).Order("timestamp DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Storage) GetAllDeviceData(ctx context.Context, stationIDs []string, limit int, offset int) ([]*model.DeviceData, error) {
	var data []*model.DeviceData
	query := s.db.WithContext(ctx).Order("timestamp DESC")

	if len(stationIDs) > 0 {
		query = query.Where("device_id IN ?", stationIDs)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Storage) GetDeviceDataByTimeRange(ctx context.Context, deviceID string, start, end time.Time) ([]*model.DeviceData, error) {
	var data []*model.DeviceData
	err := s.db.WithContext(ctx).
		Where("device_id = ? AND timestamp >= ? AND timestamp <= ?", deviceID, start, end).
		Order("timestamp ASC").
		Find(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Storage) GetLatestDeviceData(ctx context.Context, deviceID string) (*model.DeviceData, error) {
	var data model.DeviceData
	err := s.db.WithContext(ctx).
		Where("device_id = ?", deviceID).
		Order("timestamp DESC").
		First(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (s *Storage) SaveTenant(ctx context.Context, tenant *model.Tenant) error {
	return s.db.WithContext(ctx).Save(tenant).Error
}

func (s *Storage) GetTenant(ctx context.Context, id string) (*model.Tenant, error) {
	var tenant model.Tenant
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (s *Storage) ListTenants(ctx context.Context) ([]*model.Tenant, error) {
	var tenants []*model.Tenant
	err := s.db.WithContext(ctx).Find(&tenants).Error
	if err != nil {
		return nil, err
	}
	return tenants, nil
}

func (s *Storage) UpdateTenant(ctx context.Context, tenant *model.Tenant) error {
	return s.db.WithContext(ctx).Save(tenant).Error
}

func (s *Storage) DeleteTenant(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&model.Tenant{}, "id = ?", id).Error
}

func (s *Storage) SaveForwardRule(ctx context.Context, rule *model.ForwardRule) error {
	return s.db.WithContext(ctx).Save(rule).Error
}

func (s *Storage) GetForwardRule(ctx context.Context, id string) (*model.ForwardRule, error) {
	var rule model.ForwardRule
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *Storage) ListForwardRules(ctx context.Context, tenantID string) ([]*model.ForwardRule, error) {
	var rules []*model.ForwardRule
	query := s.db.WithContext(ctx)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.Find(&rules).Error
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (s *Storage) UpdateForwardRule(ctx context.Context, rule *model.ForwardRule) error {
	return s.db.WithContext(ctx).Save(rule).Error
}

func (s *Storage) DeleteForwardRule(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&model.ForwardRule{}, "id = ?", id).Error
}

func (s *Storage) SaveForwardLog(ctx context.Context, log *model.ForwardLog) error {
	return s.db.WithContext(ctx).Save(log).Error
}

func (s *Storage) GetForwardLogs(ctx context.Context, ruleID string, limit int, offset int) ([]*model.ForwardLog, error) {
	var logs []*model.ForwardLog
	query := s.db.WithContext(ctx).Where("rule_id = ?", ruleID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (s *Storage) GetDeviceCount(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.Device{}).Count(&count).Error
	return count, err
}

func (s *Storage) GetOnlineDeviceCount(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.Device{}).Where("status = ?", model.StatusOnline).Count(&count).Error
	return count, err
}

func (s *Storage) GetOfflineDeviceCount(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&model.Device{}).Where("status = ?", model.StatusOffline).Count(&count).Error
	return count, err
}

func (s *Storage) GetDataCount(ctx context.Context, deviceID string) (int64, error) {
	var count int64
	query := s.db.WithContext(ctx).Model(&model.DeviceData{})
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (s *Storage) GetDeviceByStationID(ctx context.Context, stationID string) (*model.Device, error) {
	var device model.Device
	err := s.db.WithContext(ctx).Where("id = ?", stationID).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (s *Storage) CreateDeviceIfNotExists(ctx context.Context, stationID string) (*model.Device, error) {
	device, err := s.GetDeviceByStationID(ctx, stationID)
	if err == nil {
		return device, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	newDevice := &model.Device{
		ID:         stationID,
		Name:       fmt.Sprintf("站点-%s", stationID),
		DeviceType: model.DeviceTypeRTU,
		Protocol:   model.ProtocolSL651,
		Address:    "127.0.0.1",
		Port:       8080,
		Credentials: model.Credentials{
			AuthType: model.AuthTypeNone,
		},
		Status:   model.StatusOffline,
		LastSeen: time.Now(),
		Config: model.DeviceConfig{
			HeartbeatInterval: 60,
			DataInterval:      60,
			RetryCount:        3,
			Timeout:           30,
		},
	}

	err = s.SaveDevice(ctx, newDevice)
	if err != nil {
		return nil, err
	}

	return newDevice, nil
}
func (s *Storage) SaveStatusHistory(ctx context.Context, history *model.StatusHistory) error {
	return s.db.WithContext(ctx).Create(history).Error
}

func (s *Storage) GetStatusHistory(ctx context.Context, deviceID string, start, end time.Time) ([]*model.StatusHistory, error) {
	var history []*model.StatusHistory
	query := s.db.WithContext(ctx).Order("time ASC")
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if !start.IsZero() {
		query = query.Where("time >= ?", start)
	}
	if !end.IsZero() {
		query = query.Where("time <= ?", end)
	}
	err := query.Find(&history).Error
	return history, err
}

func (s *Storage) GetTrendStatistics(ctx context.Context, start, end time.Time) ([]map[string]interface{}, error) {
	// Simple hourly aggregation for online rate
	var results []map[string]interface{}
	// This is a simplified version, in a real scenario we'd use a more complex SQL query or time-series DB
	// For now, return status history directly as snapshots
	history, err := s.GetStatusHistory(ctx, "", start, end)
	if err != nil {
		return nil, err
	}

	for _, h := range history {
		results = append(results, map[string]interface{}{
			"time":      h.Time,
			"device_id": h.DeviceID,
			"status":    h.Status,
		})
	}
	return results, nil
}
