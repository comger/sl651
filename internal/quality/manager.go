package quality

import (
	"context"
	"math"
	"sl651-platform/internal/model"
	"sl651-platform/internal/storage"
	"sync"
	"time"
)

type Manager struct {
	storage *storage.Storage
	stats   map[string]*DeviceStats
	mutex   sync.Mutex
}

type DeviceStats struct {
	LastLatencies []float64
	TotalCount    int
	ErrorCount    int
	LastSeen      time.Time
}

func NewManager(storage *storage.Storage) *Manager {
	return &Manager{
		storage: storage,
		stats:   make(map[string]*DeviceStats),
	}
}

func (m *Manager) Evaluate(ctx context.Context, deviceID string, rawData string, parsedData map[string]interface{}, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	s, ok := m.stats[deviceID]
	if !ok {
		s = &DeviceStats{LastLatencies: []float64{}}
		m.stats[deviceID] = s
	}

	metric := &model.QualityMetric{
		DeviceID: deviceID,
		Time:     time.Now(),
	}

	// 1. Error Rate & Completeness
	s.TotalCount++
	if err != nil {
		s.ErrorCount++
	}
	metric.ErrorRate = (float64(s.ErrorCount) / float64(s.TotalCount)) * 100

	// Completeness: In a real scenario, check if all expected fields are present
	// For simplicity, if parsed successfully, completeness is high
	if err == nil {
		metric.Completeness = 100.0
	} else {
		metric.Completeness = 0.0
	}

	// 2. Latency & Jitter
	if obsTimeStr, ok := parsedData["observation_time"].(string); ok {
		obsTime, err := time.Parse("2006-01-02 15:04", obsTimeStr)
		if err == nil {
			latency := time.Since(obsTime).Seconds()
			if latency < 0 {
				latency = 0
			}
			metric.Latency = latency

			// Jitter calculation (variation in latency)
			if len(s.LastLatencies) > 0 {
				last := s.LastLatencies[len(s.LastLatencies)-1]
				metric.Jitter = math.Abs(latency - last)
			}
			s.LastLatencies = append(s.LastLatencies, latency)
			if len(s.LastLatencies) > 10 {
				s.LastLatencies = s.LastLatencies[1:]
			}
		}
	}

	// 3. Comprehensive Score
	// Score = 40% Completeness + 30% (1 - ErrorRate) + 30% (Latency factor)
	latencyFactor := math.Max(0, 100-metric.Latency/10) // 10s latency = 0 score for that part
	metric.Score = (metric.Completeness * 0.4) + ((100 - metric.ErrorRate) * 0.3) + (latencyFactor * 0.3)

	m.storage.SaveQualityMetric(ctx, metric)
}

func (m *Manager) GetLatestMetrics(ctx context.Context, deviceID string, limit int) ([]*model.QualityMetric, error) {
	return m.storage.GetLatestQualityMetrics(ctx, deviceID, limit)
}

func (m *Manager) GetStats(ctx context.Context, deviceID string, duration time.Duration) (map[string]interface{}, error) {
	end := time.Now()
	start := end.Add(-duration)
	return m.storage.GetQualityStats(ctx, deviceID, start, end)
}

func (m *Manager) GetStatsPerDevice(ctx context.Context, duration time.Duration) ([]map[string]interface{}, error) {
	end := time.Now()
	start := end.Add(-duration)
	return m.storage.GetQualityStatsPerDevice(ctx, start, end)
}
