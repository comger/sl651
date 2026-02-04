package forward

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"sl651-platform/internal/model"
	"sl651-platform/internal/storage"
)

type Service struct {
	storage    *storage.Storage
	dataChan   chan *model.DeviceData
	rules      map[string]*model.ForwardRule
	rulesMutex sync.RWMutex
	workers    int
}

func NewService(storage *storage.Storage) *Service {
	return &Service{
		storage:  storage,
		dataChan: make(chan *model.DeviceData, 10000),
		rules:    make(map[string]*model.ForwardRule),
		workers:  4,
	}
}

func (fs *Service) Start(ctx context.Context) {
	log.Println("Forward service started")

	go fs.loadRules(ctx)

	for i := 0; i < fs.workers; i++ {
		go fs.worker(ctx, i)
	}

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Forward service stopped")
			return
		case <-ticker.C:
			fs.loadRules(ctx)
		}
	}
}

func (fs *Service) loadRules(ctx context.Context) {
	rules, err := fs.storage.ListForwardRules(ctx, "")
	if err != nil {
		log.Printf("Failed to load forward rules: %v", err)
		return
	}

	fs.rulesMutex.Lock()
	defer fs.rulesMutex.Unlock()

	fs.rules = make(map[string]*model.ForwardRule)
	for _, rule := range rules {
		if rule.Enabled {
			fs.rules[rule.ID] = rule
		}
	}

	log.Printf("Loaded %d forward rules", len(fs.rules))
}

func (fs *Service) worker(ctx context.Context, workerID int) {
	log.Printf("Forward worker %d started", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Forward worker %d stopped", workerID)
			return
		case data := <-fs.dataChan:
			fs.processData(ctx, data, workerID)
		}
	}
}

func (fs *Service) processData(ctx context.Context, data *model.DeviceData, workerID int) {
	fs.rulesMutex.RLock()
	defer fs.rulesMutex.RUnlock()

	for _, rule := range fs.rules {
		if fs.shouldForward(rule, data) {
			if err := fs.forwardData(ctx, rule, data); err != nil {
				log.Printf("Worker %d: Failed to forward data: %v", workerID, err)
				fs.logForwardResult(ctx, rule.ID, data.ID, "failed", err.Error(), 0)
			} else {
				fs.logForwardResult(ctx, rule.ID, data.ID, "success", "", 0)
			}
		}
	}
}

func (fs *Service) shouldForward(rule *model.ForwardRule, data *model.DeviceData) bool {
	if len(rule.Filter.DeviceIDs) > 0 {
		found := false
		for _, deviceID := range rule.Filter.DeviceIDs {
			if deviceID == data.DeviceID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(rule.Filter.DataTypes) > 0 {
		found := false
		for _, dataType := range rule.Filter.DataTypes {
			if dataType == data.DataType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (fs *Service) forwardData(ctx context.Context, rule *model.ForwardRule, data *model.DeviceData) error {
	switch rule.Destination.DestType {
	case model.DestTypeHttp:
		return fs.forwardHTTP(ctx, rule, data)
	default:
		return fmt.Errorf("unsupported destination type: %s", rule.Destination.DestType)
	}
}

func (fs *Service) forwardHTTP(ctx context.Context, rule *model.ForwardRule, data *model.DeviceData) error {
	payload := fs.transformData(rule, data)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", rule.Destination.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	for _, header := range rule.Destination.Headers {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (fs *Service) transformData(rule *model.ForwardRule, data *model.DeviceData) map[string]interface{} {
	result := make(map[string]interface{})

	result["device_id"] = data.DeviceID
	result["timestamp"] = data.Timestamp.Unix()
	result["data_type"] = string(data.DataType)
	result["quality"] = string(data.Quality)

	values := make(map[string]interface{})
	for _, point := range data.Values {
		values[point.Tag] = fs.transformValue(point.Value)
	}
	result["values"] = values

	for _, mapping := range rule.Transform.Mappings {
		if sourceValue, ok := values[mapping.Source]; ok {
			values[mapping.Target] = fs.applyTransform(mapping.Transform, sourceValue)
		}
	}

	if rule.Transform.Template != nil {
		result["formatted"] = fs.applyTemplate(*rule.Transform.Template, result)
	}

	return result
}

func (fs *Service) transformValue(value model.DataValue) interface{} {
	if value.Float != nil {
		return *value.Float
	}
	if value.Int != nil {
		return *value.Int
	}
	if value.String != nil {
		return *value.String
	}
	if value.Bool != nil {
		return *value.Bool
	}
	return nil
}

func (fs *Service) applyTransform(transformFunc *model.TransformFunc, value interface{}) interface{} {
	if transformFunc == nil {
		return value
	}

	switch *transformFunc {
	case model.TransformFuncNone:
		return value
	case model.TransformFuncToString:
		return fmt.Sprintf("%v", value)
	case model.TransformFuncToInt:
		switch v := value.(type) {
		case float64:
			return int64(v)
		case string:
			var result int64
			fmt.Sscanf(v, "%d", &result)
			return result
		default:
			return value
		}
	case model.TransformFuncToFloat:
		switch v := value.(type) {
		case int64:
			return float64(v)
		case string:
			var result float64
			fmt.Sscanf(v, "%f", &result)
			return result
		default:
			return value
		}
	default:
		return value
	}
}

func (fs *Service) applyTemplate(template string, data map[string]interface{}) string {
	result := template
	for k, v := range data {
		placeholder := fmt.Sprintf("{{%s}}", k)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", v))
	}
	return result
}

func (fs *Service) logForwardResult(ctx context.Context, ruleID, dataID, status, errorMessage string, retryCount int) {
	forwardLog := &model.ForwardLog{
		ID:           fmt.Sprintf("%d", time.Now().UnixNano()),
		RuleID:       ruleID,
		DeviceID:     "",
		DataID:       dataID,
		Status:       status,
		ErrorMessage: errorMessage,
		RetryCount:   retryCount,
	}

	if err := fs.storage.SaveForwardLog(ctx, forwardLog); err != nil {
		log.Printf("Failed to save forward log: %v", err)
	}
}

func (fs *Service) GetDataChannel() chan *model.DeviceData {
	return fs.dataChan
}

func (fs *Service) GetForwardStatistics(ctx context.Context) (*ForwardStatistics, error) {
	total, err := fs.storage.GetDataCount(ctx, "")
	if err != nil {
		return nil, err
	}

	return &ForwardStatistics{
		TotalData:   total,
		ActiveRules: int64(len(fs.rules)),
	}, nil
}

type ForwardStatistics struct {
	TotalData   int64 `json:"total_data"`
	ActiveRules int64 `json:"active_rules"`
}
