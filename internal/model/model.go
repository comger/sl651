package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type Device struct {
	ID          string       `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"not null"`
	DeviceType  DeviceType   `json:"device_type" gorm:"not null"`
	Protocol    Protocol     `json:"protocol" gorm:"not null"`
	Address     string       `json:"address" gorm:"not null"`
	Port        int          `json:"port" gorm:"not null"`
	Credentials Credentials  `json:"credentials" gorm:"type:text"`
	Status      DeviceStatus `json:"status" gorm:"not null"`
	LastSeen    time.Time    `json:"last_seen" gorm:"not null"`
	Config      DeviceConfig `json:"config" gorm:"type:text"`
	CreatedAt   time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

func (c *Credentials) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c Credentials) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *DeviceConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c DeviceConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

type DeviceType string

const (
	DeviceTypeRTU        DeviceType = "RTU"
	DeviceTypeSensor     DeviceType = "Sensor"
	DeviceTypeController DeviceType = "Controller"
	DeviceTypeGateway    DeviceType = "Gateway"
)

type Protocol string

const (
	ProtocolSL651  Protocol = "SL651"
	ProtocolModbus Protocol = "Modbus"
	ProtocolCustom Protocol = "Custom"
)

type Credentials struct {
	AuthType    AuthType `json:"auth_type"`
	Username    string   `json:"username,omitempty"`
	Password    string   `json:"password,omitempty"`
	Certificate string   `json:"certificate,omitempty"`
}

type AuthType string

const (
	AuthTypeNone        AuthType = "none"
	AuthTypePassword    AuthType = "password"
	AuthTypeCertificate AuthType = "certificate"
)

type DeviceStatus string

const (
	StatusOnline  DeviceStatus = "online"
	StatusOffline DeviceStatus = "offline"
	StatusError   DeviceStatus = "error"
)

type DeviceConfig struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
	DataInterval      int `json:"data_interval"`
	RetryCount        int `json:"retry_count"`
	Timeout           int `json:"timeout"`
}

type DataPoints []DataPoint

func (v *DataPoints) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, v)
}

func (v DataPoints) Value() (driver.Value, error) {
	return json.Marshal(v)
}

type DeviceData struct {
	ID           string      `json:"id" gorm:"primaryKey"`
	DeviceID     string      `json:"device_id" gorm:"not null;index"`
	Timestamp    time.Time   `json:"timestamp" gorm:"not null;index"`
	DataType     DataType    `json:"data_type" gorm:"not null"`
	FunctionCode string      `json:"function_code" gorm:"not null;size:2"`
	Values       DataPoints  `json:"values" gorm:"type:text"`
	RawData      string      `json:"raw_data" gorm:"type:text"`
	Quality      DataQuality `json:"quality" gorm:"not null"`
	Direction    string      `json:"direction" gorm:"not null;default:uplink"`
	CreatedAt    time.Time   `json:"created_at" gorm:"autoCreateTime"`
}

type DataPoint struct {
	Tag   string    `json:"tag"`
	Value DataValue `json:"value"`
	Unit  string    `json:"unit"`
}

type DataValue struct {
	Float  *float64 `json:"float,omitempty"`
	Int    *int64   `json:"int,omitempty"`
	String *string  `json:"string,omitempty"`
	Bool   *bool    `json:"bool,omitempty"`
}

type DataType string

const (
	DataTypeRealtime DataType = "realtime"
	DataTypeAlarm    DataType = "alarm"
	DataTypeStatus   DataType = "status"
)

type DataQuality string

const (
	QualityGood      DataQuality = "good"
	QualityUncertain DataQuality = "uncertain"
	QualityBad       DataQuality = "bad"
)

type Tenant struct {
	ID          string       `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"not null;uniqueIndex"`
	Description string       `json:"description"`
	Config      TenantConfig `json:"config" gorm:"type:text"`
	Status      TenantStatus `json:"status" gorm:"not null"`
	CreatedAt   time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

func (c *TenantConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

func (c TenantConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

type TenantConfig struct {
	DataRetentionDays int `json:"data_retention_days"`
	MaxDevices        int `json:"max_devices"`
	MaxRules          int `json:"max_rules"`
}

type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusDeleted   TenantStatus = "deleted"
)

type ForwardRule struct {
	ID           string        `json:"id" gorm:"primaryKey"`
	TenantID     string        `json:"tenant_id" gorm:"not null;index"`
	Name         string        `json:"name" gorm:"not null"`
	Enabled      bool          `json:"enabled" gorm:"not null"`
	Filter       RuleFilter    `json:"filter" gorm:"type:text"`
	Transform    RuleTransform `json:"transform" gorm:"type:text"`
	Destinations Destinations  `json:"destinations" gorm:"type:text"`
	RetryPolicy  RetryPolicy   `json:"retry_policy" gorm:"type:text"`
	CreatedAt    time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

type Destinations []Destination

func (d *Destinations) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (d Destinations) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (f *RuleFilter) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, f)
}

func (f RuleFilter) Value() (driver.Value, error) {
	return json.Marshal(f)
}

func (t *RuleTransform) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, t)
}

func (t RuleTransform) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (d *Destination) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (d Destination) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (r *RetryPolicy) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, r)
}

func (r RetryPolicy) Value() (driver.Value, error) {
	return json.Marshal(r)
}

type RuleFilter struct {
	DeviceIDs  []string    `json:"device_ids"`
	DataTypes  []DataType  `json:"data_types"`
	Conditions []Condition `json:"conditions"`
}

type Condition struct {
	Field    string    `json:"field"`
	Operator Operator  `json:"operator"`
	Value    DataValue `json:"value"`
}

type Operator string

const (
	OperatorEq       Operator = "eq"
	OperatorNe       Operator = "ne"
	OperatorGt       Operator = "gt"
	OperatorLt       Operator = "lt"
	OperatorGte      Operator = "gte"
	OperatorLte      Operator = "lte"
	OperatorContains Operator = "contains"
)

type RuleTransform struct {
	Mappings []FieldMapping `json:"mappings"`
	Template *string        `json:"template,omitempty"`
}

type FieldMapping struct {
	Source    string         `json:"source"`
	Target    string         `json:"target"`
	Transform *TransformFunc `json:"transform,omitempty"`
}

type TransformFunc string

const (
	TransformFuncNone     TransformFunc = "none"
	TransformFuncToString TransformFunc = "to_string"
	TransformFuncToInt    TransformFunc = "to_int"
	TransformFuncToFloat  TransformFunc = "to_float"
	TransformFuncCustom   TransformFunc = "custom"
)

type Destination struct {
	DestType DestType            `json:"dest_type"`
	URL      string              `json:"url"`
	Auth     *DestinationAuth    `json:"auth,omitempty"`
	Headers  []map[string]string `json:"headers"`
}

type DestType string

const (
	DestTypeHttp     DestType = "http"
	DestTypeMqtt     DestType = "mqtt"
	DestTypeDatabase DestType = "database"
	DestTypeMySQL    DestType = "mysql"
	DestTypePostgres DestType = "postgres"
	DestTypeSqlite   DestType = "sqlite"
	DestTypeCustom   DestType = "custom"
)

type DestinationAuth struct {
	AuthType AuthType `json:"auth_type"`
	Username *string  `json:"username,omitempty"`
	Password *string  `json:"password,omitempty"`
	Token    *string  `json:"token,omitempty"`
}

type RetryPolicy struct {
	MaxRetries        int     `json:"max_retries"`
	RetryInterval     int     `json:"retry_interval"`
	BackoffMultiplier float64 `json:"backoff_multiplier"`
}

type ForwardLog struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	RuleID         string    `json:"rule_id" gorm:"not null;index"`
	DeviceID       string    `json:"device_id" gorm:"not null"`
	DataID         string    `json:"data_id" gorm:"not null"`
	Status         string    `json:"status" gorm:"not null"`
	TargetType     string    `json:"target_type" gorm:"not null"`
	DestinationURL string    `json:"destination_url" gorm:"type:text"`
	Payload        string    `json:"payload" gorm:"type:text"`
	ErrorMessage   string    `json:"error_message"`
	RetryCount     int       `json:"retry_count" gorm:"not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"not null;index"`
}

type StatusHistory struct {
	ID       uint         `json:"id" gorm:"primaryKey"`
	DeviceID string       `json:"device_id" gorm:"index"`
	Status   DeviceStatus `json:"status" gorm:"not null"`
	Time     time.Time    `json:"time" gorm:"index"`
}

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

type FaultType string

const (
	FaultTypeComm     FaultType = "communication"
	FaultTypeData     FaultType = "data"
	FaultTypeHardware FaultType = "hardware"
)

type FaultLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	DeviceID  string    `json:"device_id" gorm:"index"`
	FaultCode string    `json:"fault_code" gorm:"index"` // Format: F-A-BB-CC
	Type      FaultType `json:"type" gorm:"not null"`
	Severity  Severity  `json:"severity" gorm:"not null"`
	Message   string    `json:"message" gorm:"not null"`
	Details   string    `json:"details"`
	Time      time.Time `json:"time" gorm:"index"`
	Resolved  bool      `json:"resolved" gorm:"default:false"`
}

type SystemLog struct {
	ID      uint      `json:"id" gorm:"primaryKey"`
	Level   string    `json:"level" gorm:"index"`
	Source  string    `json:"source" gorm:"index"`
	Message string    `json:"message"`
	Time    time.Time `json:"time" gorm:"index"`
}

type QualityMetric struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	DeviceID     string    `json:"device_id" gorm:"index"`
	Completeness float64   `json:"completeness"` // 0-100%
	Latency      float64   `json:"latency"`      // Seconds
	ErrorRate    float64   `json:"error_rate"`   // 0-100%
	Jitter       float64   `json:"jitter"`       // Latency variation
	Score        float64   `json:"score"`        // Comprehensive score 0-100
	Time         time.Time `json:"time" gorm:"index"`
}

type CommandStatus string

const (
	CommandStatusPending CommandStatus = "pending"
	CommandStatusSent    CommandStatus = "sent"
	CommandStatusSuccess CommandStatus = "success"
	CommandStatusFailed  CommandStatus = "failed"
	CommandStatusTimeout CommandStatus = "timeout"
)

type DeviceCommand struct {
	ID           string        `json:"id" gorm:"primaryKey"`
	DeviceID     string        `json:"device_id" gorm:"index"`
	FunctionCode string        `json:"function_code"`
	Payload      string        `json:"payload"` // Hex string or JSON
	Status       CommandStatus `json:"status" gorm:"index"`
	Result       string        `json:"result"`
	CreatedAt    time.Time     `json:"created_at" gorm:"autoCreateTime"`
	SentAt       *time.Time    `json:"sent_at"`
	UpdatedAt    time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}
