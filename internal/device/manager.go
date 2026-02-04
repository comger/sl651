package device

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"sl651-platform/internal/model"
	"sl651-platform/internal/sl651"
	"sl651-platform/internal/storage"
)

type Manager struct {
	storage  *storage.Storage
	protocol *sl651.Protocol
	devices  map[string]*model.Device
	mutex    sync.RWMutex
	dataChan chan *model.DeviceData
}

func NewManager(storage *storage.Storage, protocol *sl651.Protocol) *Manager {
	return &Manager{
		storage:  storage,
		protocol: protocol,
		devices:  make(map[string]*model.Device),
		dataChan: make(chan *model.DeviceData, 10000),
	}
}

func (dm *Manager) Start(ctx context.Context) {
	log.Println("Device manager started")

	go dm.processData(ctx)

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Device manager stopped")
			return
		case <-ticker.C:
			dm.cleanupOldDevices(ctx)
		}
	}
}

func (dm *Manager) RegisterDevice(ctx context.Context, device *model.Device) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	dm.devices[device.ID] = device
	return dm.storage.SaveDevice(ctx, device)
}

func (dm *Manager) GetDevice(ctx context.Context, id string) (*model.Device, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	device, exists := dm.devices[id]
	if !exists {
		return dm.storage.GetDevice(ctx, id)
	}
	return device, nil
}

func (dm *Manager) ListDevices(ctx context.Context) ([]*model.Device, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if len(dm.devices) > 0 {
		devices := make([]*model.Device, 0, len(dm.devices))
		for _, dev := range dm.devices {
			devices = append(devices, dev)
		}
		return devices, nil
	}

	return dm.storage.ListDevices(ctx)
}

func (dm *Manager) UpdateDevice(ctx context.Context, device *model.Device) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	dm.devices[device.ID] = device
	if err := dm.storage.UpdateDevice(ctx, device); err != nil {
		return err
	}
	// Recording history can be improved to only record on CHANGE, but for simplicity we record on every update if status is updated
	return dm.SaveStatusHistory(ctx, device.ID, device.Status)
}

func (dm *Manager) DeleteDevice(ctx context.Context, id string) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	delete(dm.devices, id)
	return dm.storage.DeleteDevice(ctx, id)
}

func (dm *Manager) ProcessMessage(ctx context.Context, msg *sl651.Message) error {
	device, err := dm.storage.CreateDeviceIfNotExists(ctx, msg.StationID)
	if err != nil {
		return fmt.Errorf("failed to get or create device: %w", err)
	}

	device.Status = model.StatusOnline
	device.LastSeen = msg.Timestamp

	if err := dm.UpdateDevice(ctx, device); err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}

	standardData, err := dm.protocol.ConvertToStandardData(msg)
	if err != nil {
		return fmt.Errorf("failed to convert data: %w", err)
	}

	deviceData := &model.DeviceData{
		ID:        dm.protocol.GenerateID(),
		DeviceID:  msg.StationID,
		Timestamp: msg.Timestamp,
		DataType:  model.DataType(dm.protocol.GetDataType(msg.FunctionCode)),
		Values:    make(model.DataPoints, 0),
		RawData:   msg.RawData,
		Quality:   model.QualityGood,
	}

	for tag, value := range standardData {
		dataPoint := model.DataPoint{
			Tag: tag,
		}

		switch v := value.(type) {
		case float64:
			dataPoint.Value.Float = &v
		case int64:
			dataPoint.Value.Int = &v
		case string:
			dataPoint.Value.String = &v
		case bool:
			dataPoint.Value.Bool = &v
		}

		deviceData.Values = append(deviceData.Values, dataPoint)
	}

	select {
	case dm.dataChan <- deviceData:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (dm *Manager) GetDataChannel() <-chan *model.DeviceData {
	return dm.dataChan
}

func (dm *Manager) processData(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-dm.dataChan:
			if err := dm.storage.SaveDeviceData(ctx, data); err != nil {
				log.Printf("Failed to save device data: %v", err)
			}
		}
	}
}

func (dm *Manager) cleanupOldDevices(ctx context.Context) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	now := time.Now()
	timeout := 5 * time.Minute

	for id, device := range dm.devices {
		if now.Sub(device.LastSeen) > timeout && device.Status == model.StatusOnline {
			device.Status = model.StatusOffline
			if err := dm.storage.UpdateDevice(ctx, device); err != nil {
				log.Printf("Failed to update device status: %v", err)
			}
			log.Printf("Device %s marked as offline", id)
		}
	}
}

func (dm *Manager) GetDeviceStatistics(ctx context.Context) (*DeviceStatistics, error) {
	total, err := dm.storage.GetDeviceCount(ctx)
	if err != nil {
		return nil, err
	}

	online, err := dm.storage.GetOnlineDeviceCount(ctx)
	if err != nil {
		return nil, err
	}

	offline, err := dm.storage.GetOfflineDeviceCount(ctx)
	if err != nil {
		return nil, err
	}

	return &DeviceStatistics{
		Total:   total,
		Online:  online,
		Offline: offline,
	}, nil
}

type DeviceStatistics struct {
	Total   int64 `json:"total"`
	Online  int64 `json:"online"`
	Offline int64 `json:"offline"`
}

func (dm *Manager) ImportFromCSV(ctx context.Context, csvData [][]string) error {
	for _, record := range csvData {
		if len(record) < 6 {
			continue
		}

		msg, err := dm.protocol.ParseCSVRecord(record)
		if err != nil {
			log.Printf("Failed to parse CSV record: %v", err)
			continue
		}

		if err := dm.ProcessMessage(ctx, msg); err != nil {
			log.Printf("Failed to process message: %v", err)
			continue
		}
	}

	return nil
}

func (dm *Manager) GetDeviceData(ctx context.Context, deviceID string, limit, offset int) ([]*model.DeviceData, error) {
	return dm.storage.GetDeviceData(ctx, deviceID, limit, offset)
}

func (dm *Manager) GetAllDeviceData(ctx context.Context, stationIDs []string, limit int, offset int) ([]*model.DeviceData, error) {
	return dm.storage.GetAllDeviceData(ctx, stationIDs, limit, offset)
}

func (dm *Manager) GetLatestDeviceData(ctx context.Context, deviceID string) (*model.DeviceData, error) {
	return dm.storage.GetLatestDeviceData(ctx, deviceID)
}

func (dm *Manager) GetDeviceDataByTimeRange(ctx context.Context, deviceID string, start, end time.Time) ([]*model.DeviceData, error) {
	return dm.storage.GetDeviceDataByTimeRange(ctx, deviceID, start, end)
}

func (dm *Manager) GetTrendStatistics(ctx context.Context, start, end time.Time) ([]map[string]interface{}, error) {
	return dm.storage.GetTrendStatistics(ctx, start, end)
}

func (dm *Manager) SaveStatusHistory(ctx context.Context, deviceID string, status model.DeviceStatus) error {
	history := &model.StatusHistory{
		DeviceID: deviceID,
		Status:   status,
		Time:     time.Now(),
	}
	return dm.storage.SaveStatusHistory(ctx, history)
}
