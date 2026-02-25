package storage

import (
	"context"
	"fmt"
	"log"
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
	db := s.db.WithContext(ctx).Debug() // Enable Debug for this call
	query := db.Order("timestamp DESC")

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
	log.Printf("[STORAGE DEBUG] GetAllDeviceData query IDs: %v, found: %d, err: %v", stationIDs, len(data), err)
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

func (s *Storage) GetForwardLogs(ctx context.Context, ruleID, deviceID, status string, limit int, offset int) ([]*model.ForwardLog, error) {
	var logs []*model.ForwardLog
	query := s.db.WithContext(ctx).Order("created_at DESC")

	if ruleID != "" {
		query = query.Where("rule_id = ?", ruleID)
	}
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

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

func (s *Storage) SaveFaultLog(ctx context.Context, log *model.FaultLog) error {
	return s.db.WithContext(ctx).Create(log).Error
}

func (s *Storage) GetFaultLogs(ctx context.Context, deviceID string, limit int) ([]*model.FaultLog, error) {
	var logs []*model.FaultLog
	query := s.db.WithContext(ctx).Order("time DESC")
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&logs).Error
	return logs, err
}

func (s *Storage) SaveSystemLog(ctx context.Context, log *model.SystemLog) error {
	return s.db.WithContext(ctx).Create(log).Error
}

func (s *Storage) GetSystemLogs(ctx context.Context, limit int) ([]*model.SystemLog, error) {
	var logs []*model.SystemLog
	query := s.db.WithContext(ctx).Order("time DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&logs).Error
	return logs, err
}

func (s *Storage) SaveQualityMetric(ctx context.Context, metric *model.QualityMetric) error {
	return s.db.WithContext(ctx).Create(metric).Error
}

func (s *Storage) GetLatestQualityMetrics(ctx context.Context, deviceID string, limit int) ([]*model.QualityMetric, error) {
	var metrics []*model.QualityMetric
	query := s.db.WithContext(ctx).Order("time DESC")
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&metrics).Error
	return metrics, err
}

func (s *Storage) GetQualityStats(ctx context.Context, deviceID string, start, end time.Time) (map[string]interface{}, error) {
	var result struct {
		AvgCompleteness float64 `gorm:"column:avg_completeness"`
		AvgLatency      float64 `gorm:"column:avg_latency"`
		AvgErrorRate    float64 `gorm:"column:avg_error_rate"`
		AvgJitter       float64 `gorm:"column:avg_jitter"`
		AvgScore        float64 `gorm:"column:avg_score"`
		Count           int64   `gorm:"column:count"`
	}

	query := s.db.WithContext(ctx).Model(&model.QualityMetric{}).
		Select("AVG(completeness) as avg_completeness, AVG(latency) as avg_latency, AVG(error_rate) as avg_error_rate, AVG(jitter) as avg_jitter, AVG(score) as avg_score, COUNT(*) as count")

	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if !start.IsZero() {
		query = query.Where("time >= ?", start)
	}
	if !end.IsZero() {
		query = query.Where("time <= ?", end)
	}

	err := query.Scan(&result).Error
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"avg_completeness": result.AvgCompleteness,
		"avg_latency":      result.AvgLatency,
		"avg_error_rate":   result.AvgErrorRate,
		"avg_jitter":       result.AvgJitter,
		"avg_score":        result.AvgScore,
		"count":            result.Count,
	}, nil
}

func (s *Storage) GetQualityStatsPerDevice(ctx context.Context, start, end time.Time) ([]map[string]interface{}, error) {
	var results []struct {
		DeviceID        string  `gorm:"column:device_id"`
		AvgCompleteness float64 `gorm:"column:avg_completeness"`
		AvgLatency      float64 `gorm:"column:avg_latency"`
		AvgErrorRate    float64 `gorm:"column:avg_error_rate"`
		AvgJitter       float64 `gorm:"column:avg_jitter"`
		AvgScore        float64 `gorm:"column:avg_score"`
		Count           int64   `gorm:"column:count"`
		LastTimeStr     string  `gorm:"column:last_time"`
	}

	query := s.db.WithContext(ctx).Model(&model.QualityMetric{}).
		Select("device_id, AVG(completeness) as avg_completeness, AVG(latency) as avg_latency, AVG(error_rate) as avg_error_rate, AVG(jitter) as avg_jitter, AVG(score) as avg_score, COUNT(*) as count, MAX(time) as last_time").
		Group("device_id")

	if !start.IsZero() {
		query = query.Where("time >= ?", start)
	}
	if !end.IsZero() {
		query = query.Where("time <= ?", end)
	}

	err := query.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	final := make([]map[string]interface{}, 0)
	for _, r := range results {
		lastTime, _ := time.Parse(time.RFC3339, r.LastTimeStr)
		if r.LastTimeStr != "" && lastTime.IsZero() {
			// Fallback for standard SQLite format
			lastTime, _ = time.Parse("2006-01-02 15:04:05", r.LastTimeStr)
		}

		final = append(final, map[string]interface{}{
			"device_id":        r.DeviceID,
			"avg_completeness": r.AvgCompleteness,
			"avg_latency":      r.AvgLatency,
			"avg_error_rate":   r.AvgErrorRate,
			"avg_jitter":       r.AvgJitter,
			"avg_score":        r.AvgScore,
			"count":            r.Count,
			"last_time":        lastTime,
		})
	}

	return final, nil
}

func (s *Storage) SaveCommand(ctx context.Context, cmd *model.DeviceCommand) error {
	return s.db.WithContext(ctx).Create(cmd).Error
}

func (s *Storage) GetCommand(ctx context.Context, id string) (*model.DeviceCommand, error) {
	var cmd model.DeviceCommand
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&cmd).Error
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}

func (s *Storage) UpdateCommandStatus(ctx context.Context, id string, status model.CommandStatus, result string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if status == model.CommandStatusSent {
		now := time.Now()
		updates["sent_at"] = &now
	}
	if result != "" {
		updates["result"] = result
	}
	return s.db.WithContext(ctx).Model(&model.DeviceCommand{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Storage) GetCommands(ctx context.Context, deviceID string, limit int) ([]*model.DeviceCommand, error) {
	var cmds []*model.DeviceCommand
	query := s.db.WithContext(ctx).Order("created_at DESC")
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&cmds).Error
	if err != nil {
		return nil, err
	}
	return cmds, nil
}

func (s *Storage) ListPendingCommands(ctx context.Context) ([]*model.DeviceCommand, error) {
	var cmds []*model.DeviceCommand
	err := s.db.WithContext(ctx).Where("status = ?", model.CommandStatusPending).Find(&cmds).Error
	if err != nil {
		return nil, err
	}
	return cmds, nil
}
