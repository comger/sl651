# 水文设备统一运维平台技术设计方案

## 1. 技术选型审查

### 1.1 需求分析

| 需求项 | 说明 | 优先级 |
|--------|------|--------|
| 私有化部署 | 支持客户本地环境部署 | 高 |
| 接入能力 | 支持小于1万台设备 | 高 |
| 跨平台 | 支持Windows/Linux/macOS | 高 |
| 轻量化 | 资源占用低，部署简单 | 高 |
| 单体应用 | 简化部署和运维 | 高 |
| 高性能 | 支持高并发数据处理 | 高 |
| 服务解耦 | 接入服务与转发服务分离 | 高 |
| 信创兼容 | 支持国产化环境 | 中 |

### 1.2 技术选型对比

#### 1.2.1 编程语言选型

| 语言 | 跨平台 | 性能 | 资源占用 | 部署复杂度 | 信创支持 | 开发效率 | 综合评分 |
|------|--------|------|----------|------------|----------|----------|----------|
| Go | 优秀 | 优秀 | 低 | 低 | 优秀 | 优秀 | 9.5 |
| Rust | 优秀 | 优秀 | 低 | 低 | 优秀 | 良好 | 9.0 |
| Java | 良好 | 良好 | 高 | 中 | 良好 | 优秀 | 8.0 |
| Node.js | 优秀 | 中 | 中 | 低 | 良好 | 优秀 | 8.5 |
| C++ | 优秀 | 优秀 | 中 | 高 | 优秀 | 中 | 8.0 |

**推荐选择：Go (Golang)**

#### 1.2.2 选择Go的理由

| 优势 | 说明 |
|------|------|
| 高性能 | 原生支持并发，Goroutine轻量高效 |
| 跨平台 | 一次编译，多平台运行，完美支持信创环境 |
| 轻量化 | 单一可执行文件，无运行时依赖 |
| 开发效率高 | 语法简洁，标准库丰富，开发周期短 |
| 生态完善 | 丰富的第三方库，社区活跃 |
| 部署简单 | 静态编译，无需安装依赖 |
| 内存安全 | 垃圾回收机制，避免内存泄漏 |
| 并发模型 | CSP并发模型，适合高并发场景 |

### 1.3 技术栈选型

| 组件 | 技术选型 | 说明 |
|------|----------|------|
| 编程语言 | Go 1.21+ | 高性能、跨平台、并发友好 |
| Web框架 | Gin | 高性能HTTP框架 |
| 数据库 | SQLite | 嵌入式、零配置 |
| ORM | GORM | 强大的ORM框架 |
| 序列化 | encoding/json | 标准库JSON序列化 |
| 网络协议 | net, crypto/tls | 标准库网络支持 |
| 日志 | zap | 高性能结构化日志 |
| 配置管理 | Viper | 灵活的配置管理 |
| 监控 | Prometheus | 指标收集和监控 |
| 消息队列 | channel | 内置channel，轻量高效 |

## 2. 迭代版本规划

### 2.1 版本迭代路线图

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                         版本迭代路线图                                    │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  V1.0          V1.1          V1.2          V1.3          V2.0            │
│  │             │             │             │             │             │
│  ▼             ▼             ▼             ▼             ▼             │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐     │
│  │设备管理  │  │心跳管理  │  │故障诊断  │  │参数配置  │  │完整功能  │     │
│  │         │→ │         │→ │         │→ │         │→ │         │     │
│  │基础接入  │  │状态监控  │  │异常检测  │  │远程读写  │  │性能优化  │     │
│  │数据采集  │  │告警通知  │  │日志分析  │  │批量操作  │  │集群部署  │     │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘  └─────────┘     │
│                                                                         │
│  ──────────────────────────────────────────────────────────────────    │
│  2周           2周           2周           2周           4周           │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.2 V1.0 设备管理版本

#### 2.2.1 功能范围

| 模块 | 功能 | 说明 |
|------|------|------|
| 设备接入 | 设备注册 | 支持手动添加和自动发现设备 |
| 设备接入 | 设备认证 | 支持密码认证和证书认证 |
| 设备接入 | 协议解析 | 支持SL651协议ASCII/HEX/BCD编码 |
| 设备接入 | 连接管理 | 支持TCP/TLS连接 |
| 数据采集 | 实时采集 | 支持定时报和加报 |
| 数据采集 | 数据解析 | 解析SL651报文数据 |
| 数据存储 | 数据持久化 | 存储设备数据和状态 |
| 数据转发 | 基础转发 | 支持HTTP和MQTT转发 |

#### 2.2.2 技术实现

```go
package device

import (
    "context"
    "sync"
    "time"
)

type DeviceManager struct {
    devices map[string]*Device
    mutex   sync.RWMutex
    storage Storage
}

type Device struct {
    ID          string
    Name        string
    Protocol    string
    Address     string
    Port        int
    Credentials Credentials
    Status      DeviceStatus
    LastSeen    time.Time
    Config      DeviceConfig
}

type DeviceStatus string

const (
    StatusOnline  DeviceStatus = "online"
    StatusOffline DeviceStatus = "offline"
    StatusError   DeviceStatus = "error"
)

func (dm *DeviceManager) RegisterDevice(ctx context.Context, device *Device) error {
    dm.mutex.Lock()
    defer dm.mutex.Unlock()
    
    dm.devices[device.ID] = device
    return dm.storage.SaveDevice(ctx, device)
}

func (dm *DeviceManager) GetDevice(ctx context.Context, id string) (*Device, error) {
    dm.mutex.RLock()
    defer dm.mutex.RUnlock()
    
    device, exists := dm.devices[id]
    if !exists {
        return nil, ErrDeviceNotFound
    }
    return device, nil
}
```

#### 2.2.3 API接口

| 路径 | 方法 | 描述 |
|------|------|------|
| /api/v1/devices | GET | 获取设备列表 |
| /api/v1/devices | POST | 注册设备 |
| /api/v1/devices/{id} | GET | 获取设备详情 |
| /api/v1/devices/{id} | PUT | 更新设备信息 |
| /api/v1/devices/{id} | DELETE | 删除设备 |
| /api/v1/devices/{id}/data | GET | 获取设备数据 |

### 2.3 V1.1 心跳管理版本

#### 2.3.1 功能范围

| 模块 | 功能 | 说明 |
|------|------|------|
| 心跳管理 | 心跳检测 | 定时检测设备心跳 |
| 心跳管理 | 状态监控 | 实时监控设备在线状态 |
| 心跳管理 | 超时处理 | 心跳超时自动标记离线 |
| 告警通知 | 离线告警 | 设备离线发送告警 |
| 告警通知 | 恢复通知 | 设备恢复上线通知 |
| 统计分析 | 在线统计 | 统计设备在线率 |
| 统计分析 | 趋势分析 | 分析设备在线趋势 |

#### 2.3.2 技术实现

```go
package heartbeat

import (
    "context"
    "time"
)

type HeartbeatManager struct {
    deviceManager *device.DeviceManager
    notifier      *notifier.Notifier
    interval      time.Duration
    timeout       time.Duration
}

func (hm *HeartbeatManager) Start(ctx context.Context) {
    ticker := time.NewTicker(hm.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            hm.checkHeartbeats(ctx)
        }
    }
}

func (hm *HeartbeatManager) checkHeartbeats(ctx context.Context) {
    devices, err := hm.deviceManager.ListDevices(ctx)
    if err != nil {
        return
    }
    
    now := time.Now()
    for _, dev := range devices {
        if now.Sub(dev.LastSeen) > hm.timeout {
            if dev.Status != device.StatusOffline {
                hm.handleDeviceOffline(ctx, dev)
            }
        }
    }
}

func (hm *HeartbeatManager) handleDeviceOffline(ctx context.Context, dev *device.Device) {
    dev.Status = device.StatusOffline
    hm.deviceManager.UpdateDevice(ctx, dev)
    
    alert := &notifier.Alert{
        Type:    notifier.AlertTypeOffline,
        Device:  dev.ID,
        Message: "Device offline due to heartbeat timeout",
    }
    hm.notifier.SendAlert(ctx, alert)
}
```

#### 2.3.3 API接口

| 路径 | 方法 | 描述 |
|------|------|------|
| /api/v1/heartbeat/status | GET | 获取心跳状态 |
| /api/v1/heartbeat/config | GET | 获取心跳配置 |
| /api/v1/heartbeat/config | PUT | 更新心跳配置 |
| /api/v1/statistics/online | GET | 获取在线统计 |
| /api/v1/statistics/trend | GET | 获取趋势分析 |

### 2.4 V1.2 故障诊断版本

#### 2.4.1 功能范围

| 模块 | 功能 | 说明 |
|------|------|------|
| 故障检测 | 异常检测 | 检测设备数据异常 |
| 故障检测 | 通信异常 | 检测通信连接异常 |
| 故障检测 | 数据校验 | 校验数据完整性 |
| 故障诊断 | 故障定位 | 定位故障原因 |
| 故障诊断 | 故障分类 | 分类故障类型 |
| 日志管理 | 日志收集 | 收集设备日志 |
| 日志管理 | 日志分析 | 分析日志信息 |
| 日志管理 | 日志查询 | 查询历史日志 |

#### 2.4.2 技术实现

```go
package diagnosis

import (
    "context"
    "time"
)

type DiagnosisManager struct {
    deviceManager *device.DeviceManager
    logStorage    LogStorage
    analyzer      *Analyzer
}

type Fault struct {
    ID        string
    DeviceID  string
    Type      FaultType
    Severity  Severity
    Message   string
    Timestamp time.Time
    Details   map[string]interface{}
}

type FaultType string

const (
    FaultTypeCommunication FaultType = "communication"
    FaultTypeData         FaultType = "data"
    FaultTypeHardware     FaultType = "hardware"
)

type Severity string

const (
    SeverityInfo     Severity = "info"
    SeverityWarning  Severity = "warning"
    SeverityError    Severity = "error"
    SeverityCritical Severity = "critical"
)

func (dm *DiagnosisManager) DetectFaults(ctx context.Context, data *device.DeviceData) ([]*Fault, error) {
    var faults []*Fault
    
    if err := dm.analyzer.ValidateData(data); err != nil {
        fault := &Fault{
            ID:        generateID(),
            DeviceID:  data.DeviceID,
            Type:      FaultTypeData,
            Severity:  SeverityError,
            Message:   "Data validation failed",
            Timestamp: time.Now(),
            Details:   map[string]interface{}{"error": err.Error()},
        }
        faults = append(faults, fault)
    }
    
    return faults, nil
}

func (dm *DiagnosisManager) AnalyzeLogs(ctx context.Context, deviceID string, start, end time.Time) ([]*Fault, error) {
    logs, err := dm.logStorage.QueryLogs(ctx, deviceID, start, end)
    if err != nil {
        return nil, err
    }
    
    return dm.analyzer.AnalyzeLogs(logs), nil
}
```

#### 2.4.3 API接口

| 路径 | 方法 | 描述 |
|------|------|------|
| /api/v1/faults | GET | 获取故障列表 |
| /api/v1/faults/{id} | GET | 获取故障详情 |
| /api/v1/faults/devices/{id} | GET | 获取设备故障 |
| /api/v1/logs | GET | 获取日志列表 |
| /api/v1/logs/{id} | GET | 获取日志详情 |
| /api/v1/diagnosis/analyze | POST | 分析诊断 |

### 2.5 V1.3 参数配置版本

#### 2.5.1 功能范围

| 模块 | 功能 | 说明 |
|------|------|------|
| 参数读取 | 读取参数 | 读取设备参数 |
| 参数读取 | 批量读取 | 批量读取多个参数 |
| 参数设置 | 设置参数 | 设置设备参数 |
| 参数设置 | 批量设置 | 批量设置多个参数 |
| 参数管理 | 参数模板 | 参数模板管理 |
| 参数管理 | 参数备份 | 参数备份恢复 |
| 参数管理 | 参数校验 | 参数合法性校验 |

#### 2.5.2 技术实现

```go
package parameter

import (
    "context"
    "time"
)

type ParameterManager struct {
    deviceManager *device.DeviceManager
    protocol      *sl651.Protocol
    storage       Storage
}

type Parameter struct {
    ID        string
    DeviceID  string
    Name      string
    Value     interface{}
    Unit      string
    Type      ParameterType
    ReadOnly  bool
    UpdatedAt time.Time
}

type ParameterType string

const (
    ParameterTypeInt    ParameterType = "int"
    ParameterTypeFloat  ParameterType = "float"
    ParameterTypeString ParameterType = "string"
    ParameterTypeBool   ParameterType = "bool"
)

func (pm *ParameterManager) ReadParameter(ctx context.Context, deviceID, paramName string) (*Parameter, error) {
    dev, err := pm.deviceManager.GetDevice(ctx, deviceID)
    if err != nil {
        return nil, err
    }
    
    value, err := pm.protocol.ReadParameter(dev, paramName)
    if err != nil {
        return nil, err
    }
    
    param := &Parameter{
        ID:        generateID(),
        DeviceID:  deviceID,
        Name:      paramName,
        Value:     value,
        Type:      pm.getParameterType(value),
        ReadOnly:  pm.isReadOnly(paramName),
        UpdatedAt: time.Now(),
    }
    
    return param, nil
}

func (pm *ParameterManager) SetParameter(ctx context.Context, deviceID, paramName string, value interface{}) error {
    dev, err := pm.deviceManager.GetDevice(ctx, deviceID)
    if err != nil {
        return err
    }
    
    if err := pm.protocol.SetParameter(dev, paramName, value); err != nil {
        return err
    }
    
    param := &Parameter{
        ID:        generateID(),
        DeviceID:  deviceID,
        Name:      paramName,
        Value:     value,
        Type:      pm.getParameterType(value),
        ReadOnly:  pm.isReadOnly(paramName),
        UpdatedAt: time.Now(),
    }
    
    return pm.storage.SaveParameter(ctx, param)
}

func (pm *ParameterManager) BatchReadParameters(ctx context.Context, deviceID string, paramNames []string) ([]*Parameter, error) {
    var params []*Parameter
    for _, name := range paramNames {
        param, err := pm.ReadParameter(ctx, deviceID, name)
        if err != nil {
            return nil, err
        }
        params = append(params, param)
    }
    return params, nil
}

func (pm *ParameterManager) BatchSetParameters(ctx context.Context, deviceID string, params map[string]interface{}) error {
    dev, err := pm.deviceManager.GetDevice(ctx, deviceID)
    if err != nil {
        return err
    }
    
    for name, value := range params {
        if err := pm.protocol.SetParameter(dev, name, value); err != nil {
            return err
        }
    }
    
    return nil
}
```

#### 2.5.3 API接口

| 路径 | 方法 | 描述 |
|------|------|------|
| /api/v1/parameters | GET | 获取参数列表 |
| /api/v1/parameters/{id} | GET | 获取参数详情 |
| /api/v1/devices/{id}/parameters | GET | 获取设备参数 |
| /api/v1/devices/{id}/parameters | PUT | 设置设备参数 |
| /api/v1/devices/{id}/parameters/batch | POST | 批量设置参数 |
| /api/v1/templates | GET | 获取参数模板 |
| /api/v1/templates | POST | 创建参数模板 |

### 2.6 V2.0 完整功能版本

#### 2.6.1 功能范围

| 模块 | 功能 | 说明 |
|------|------|------|
| 性能优化 | 性能调优 | 优化系统性能 |
| 性能优化 | 内存优化 | 优化内存使用 |
| 性能优化 | 并发优化 | 优化并发处理 |
| 集群部署 | 负载均衡 | 支持多实例部署 |
| 集群部署 | 数据同步 | 数据同步机制 |
| 集群部署 | 故障转移 | 自动故障转移 |
| 高级功能 | 插件系统 | 支持自定义插件 |
| 高级功能 | 规则引擎 | 灵活的转发规则 |
| 高级功能 | 数据分析 | 高级数据分析 |

## 3. 系统架构设计

### 3.1 总体架构

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                         水文设备统一运维平台（单体应用）                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                        应用层（Go）                               │  │
│  ├──────────────────┬──────────────────┬────────────────────────────┤  │
│  │   接入服务模块    │   转发服务模块    │      管理控制模块           │  │
│  │                  │                  │                            │  │
│  │  - 设备管理      │  - 规则引擎      │  - Web API                 │  │
│  │  - 心跳管理      │  - 插件管理      │  - WebSocket               │  │
│  │  - 故障诊断      │  - 数据转发      │  - 配置管理                 │  │
│  │  - 参数配置      │  - 监控统计      │  - 用户管理                 │  │
│  └──────────────────┴──────────────────┴────────────────────────────┘  │
│                              │                                           │
│                              ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                        核心服务层                                  │  │
│  ├──────────┬──────────┬──────────┬──────────┬──────────┬───────────┤  │
│  │ 消息总线  │ 配置中心  │ 数据存储  │ 日志服务  │ 监控服务  │ 安全服务   │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────┴───────────┘  │
│                              │                                           │
│                              ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                        基础设施层                                  │  │
│  ├──────────┬──────────┬──────────┬──────────┬──────────┬───────────┤  │
│  │ 网络层    │ 加密层    │ 文件系统  │ 时间服务  │ 协程池    │ 资源管理   │  │
│  └──────────┴──────────┴──────────┴──────────┴──────────┴───────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
                              │
         ┌────────────────────┼────────────────────┐
         ▼                    ▼                    ▼
   ┌──────────┐        ┌──────────┐        ┌──────────┐
   │ SL651设备 │        │ 管理终端  │        │ 第三方平台 │
   │ (TCP/串口)│        │ (Web/CLI) │        │ (HTTP/MQTT)│
   └──────────┘        └──────────┘        └──────────┘
```

### 3.2 模块架构

#### 3.2.1 接入服务模块

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                           接入服务模块                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                          连接管理层                                │  │
│  ├──────────────────┬──────────────────┬────────────────────────────┤  │
│  │  TCP连接池       │  串口连接池      │  TLS安全连接                │  │
│  │  (net)           │  (go-serial)     │  (crypto/tls)              │  │
│  └──────────────────┴──────────────────┴────────────────────────────┘  │
│                              │                                           │
│                              ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                          协议处理层                                │  │
│  ├──────────────────┬──────────────────┬────────────────────────────┤  │
│  │  SL651解析器      │  编解码器        │  协议路由器                 │  │
│  │  (ASCII/HEX/BCD) │  (encoding)      │  (指令分发)                 │  │
│  └──────────────────┴──────────────────┴────────────────────────────┘  │
│                              │                                           │
│                              ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                          业务逻辑层                                │  │
│  ├──────────────────┬──────────────────┬────────────────────────────┤  │
│  │  设备管理         │  心跳管理        │  故障诊断                   │  │
│  │  (设备注册/认证)  │  (状态监控)      │  (异常检测)                 │  │
│  ├──────────────────┼──────────────────┼────────────────────────────┤  │
│  │  参数配置         │  数据采集        │  数据处理                   │  │
│  │  (参数读写)      │  (实时采集)      │  (数据标准化)               │  │
│  └──────────────────┴──────────────────┴────────────────────────────┘  │
│                              │                                           │
│                              ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                          数据处理层                                │  │
│  ├──────────────────┬──────────────────┬────────────────────────────┤  │
│  │  数据采集         │  数据缓冲        │  数据标准化                 │  │
│  │  (实时采集)       │  (SQLite)        │  (JSON/CSV)                │  │
│  └──────────────────┴──────────────────┴────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 3.2.2 转发服务模块

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                           转发服务模块                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                          规则引擎层                                │  │
│  ├──────────────────┬──────────────────┬────────────────────────────┤  │
│  │  规则解析器       │  条件评估器      │  路由决策器                 │  │
│  │  (DSL解析)        │  (条件匹配)      │  (目标选择)                 │  │
│  └──────────────────┴──────────────────┴────────────────────────────┘  │
│                              │                                           │
│                              ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                          插件管理层                                │  │
│  ├──────────────────┬──────────────────┬────────────────────────────┤  │
│  │  插件加载器       │  插件管理器      │  热更新机制                 │  │
│  │  (plugin)        │  (生命周期)      │  (无重启更新)               │  │
│  └──────────────────┴──────────────────┴────────────────────────────┘  │
│                              │                                           │
│                              ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                          通信适配层                                │  │
│  ├──────────────────┬──────────────────┬────────────────────────────┤  │
│  │  HTTP适配器       │  MQTT适配器      │  数据库适配器                │  │
│  │  (net/http)      │  (mqtt)          │  (database/sql)             │  │
│  ├──────────────────┼──────────────────┼────────────────────────────┤  │
│  │  WebSocket适配器  │  TCP适配器       │  自定义协议适配器             │  │
│  │  (gorilla/websocket) │ (net)       │  (插件扩展)                  │  │
│  └──────────────────┴──────────────────┴────────────────────────────┘  │
│                              │                                           │
│                              ▼                                           │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                          监控统计层                                │  │
│  ├──────────────────┬──────────────────┬────────────────────────────┤  │
│  │  指标收集         │  告警引擎        │  日志管理                   │  │
│  │  (prometheus)    │  (规则告警)      │  (zap)                      │  │
│  └──────────────────┴──────────────────┴────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 3.3 服务解耦设计

#### 3.3.1 解耦架构

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                         服务解耦架构                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────┐                                                   │
│  │   接入服务       │                                                   │
│  │  (独立模块)     │                                                   │
│  │                 │                                                   │
│  │  - 设备管理     │                                                   │
│  │  - 心跳管理     │                                                   │
│  │  - 故障诊断     │                                                   │
│  │  - 参数配置     │                                                   │
│  └────────┬────────┘                                                   │
│           │                                                            │
│           │ 消息总线 (Channel)                                          │
│           │ - 数据消息                                                   │
│           │ - 控制消息                                                   │
│           │ - 事件消息                                                   │
│           │                                                            │
│  ┌────────▼────────┐                                                   │
│  │   转发服务       │                                                   │
│  │  (独立模块)     │                                                   │
│  │                 │                                                   │
│  │  - 规则引擎     │                                                   │
│  │  - 数据转发     │                                                   │
│  │  - 监控统计     │                                                   │
│  └─────────────────┘                                                   │
│                                                                         │
│  解耦原则：                                                              │
│  1. 模块独立：接入和转发模块独立编译，可单独测试                          │
│  2. 接口清晰：通过消息总线通信，接口明确                                 │
│  3. 数据隔离：各自维护独立的数据模型                                     │
│  4. 部署灵活：可独立部署或统一部署                                        │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 3.3.2 消息总线设计

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                         消息总线设计                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                        消息总线核心                                  │  │
│  │                        (go channel)                                 │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│                              │                                           │
│         ┌────────────────────┼────────────────────┐                     │
│         ▼                    ▼                    ▼                     │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐             │
│  │ 数据通道     │      │ 控制通道     │      │ 事件通道     │             │
│  │ (buffered)  │      │ (unbuffered)│      │ (broadcast) │             │
│  └─────────────┘      └─────────────┘      └─────────────┘             │
│         │                    │                    │                     │
│         ▼                    ▼                    ▼                     │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐             │
│  │ 设备数据     │      │ 参数配置     │      │ 设备上线     │             │
│  │ 实时监测     │      │ 参数读写     │      │ 设备下线     │             │
│  │ 告警信息     │      │ 设备控制     │      │ 故障告警     │             │
│  └─────────────┘      └─────────────┘      └─────────────┘             │
│                                                                         │
│  消息特性：                                                              │
│  - 高性能：零拷贝传输，内存占用低                                         │
│  - 类型安全：强类型消息定义                                               │
│  - 背压控制：buffered channel防止内存溢出                                 │
│  - 异步处理：完全异步，不阻塞                                             │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## 4. 接口设计

### 4.1 模块间接口

#### 4.1.1 消息定义

```go
package message

import "time"

type MessageType string

const (
    MessageTypeData     MessageType = "data"
    MessageTypeControl  MessageType = "control"
    MessageTypeEvent    MessageType = "event"
)

type Message struct {
    Type      MessageType
    Timestamp time.Time
    Payload   interface{}
}

type DataMessage struct {
    DeviceID  string
    Timestamp time.Time
    DataType  DataType
    Payload   []byte
    Metadata  Metadata
}

type ControlMessage struct {
    DeviceID  string
    Command   CommandType
    Payload   []byte
    Timeout   time.Duration
}

type EventMessage struct {
    DeviceID  string
    EventType EventType
    Timestamp time.Time
    Details   string
}

type DataType string

const (
    DataTypeRealtime DataType = "realtime"
    DataTypeAlarm    DataType = "alarm"
    DataTypeStatus   DataType = "status"
)

type CommandType string

const (
    CommandReadParam   CommandType = "read_param"
    CommandWriteParam  CommandType = "write_param"
)

type EventType string

const (
    EventTypeDeviceOnline   EventType = "device_online"
    EventTypeDeviceOffline  EventType = "device_offline"
    EventTypeDeviceError    EventType = "device_error"
    EventTypeConnectionLost EventType = "connection_lost"
)

type Metadata struct {
    Source    string
    Quality   string
    Tags      map[string]string
}
```

#### 4.1.2 消息总线接口

```go
package bus

import (
    "context"
    "sl651/message"
)

type MessageBus interface {
    PublishData(ctx context.Context, msg *message.DataMessage) error
    PublishControl(ctx context.Context, msg *message.ControlMessage) error
    PublishEvent(ctx context.Context, msg *message.EventMessage) error
    
    SubscribeData(ctx context.Context) <-chan *message.DataMessage
    SubscribeControl(ctx context.Context) <-chan *message.ControlMessage
    SubscribeEvent(ctx context.Context) <-chan *message.EventMessage
    
    Close() error
}

type ChannelMessageBus struct {
    dataChan     chan *message.DataMessage
    controlChan  chan *message.ControlMessage
    eventChan    chan *message.EventMessage
}

func NewChannelMessageBus(dataBufferSize, controlBufferSize, eventBufferSize int) *ChannelMessageBus {
    return &ChannelMessageBus{
        dataChan:     make(chan *message.DataMessage, dataBufferSize),
        controlChan:  make(chan *message.ControlMessage, controlBufferSize),
        eventChan:    make(chan *message.EventMessage, eventBufferSize),
    }
}

func (b *ChannelMessageBus) PublishData(ctx context.Context, msg *message.DataMessage) error {
    select {
    case b.dataChan <- msg:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (b *ChannelMessageBus) PublishControl(ctx context.Context, msg *message.ControlMessage) error {
    select {
    case b.controlChan <- msg:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (b *ChannelMessageBus) PublishEvent(ctx context.Context, msg *message.EventMessage) error {
    select {
    case b.eventChan <- msg:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (b *ChannelMessageBus) SubscribeData(ctx context.Context) <-chan *message.DataMessage {
    return b.dataChan
}

func (b *ChannelMessageBus) SubscribeControl(ctx context.Context) <-chan *message.ControlMessage {
    return b.controlChan
}

func (b *ChannelMessageBus) SubscribeEvent(ctx context.Context) <-chan *message.EventMessage {
    return b.eventChan
}

func (b *ChannelMessageBus) Close() error {
    close(b.dataChan)
    close(b.controlChan)
    close(b.eventChan)
    return nil
}
```

### 4.2 外部API接口

#### 4.2.1 REST API设计

| 路径 | 方法 | 描述 | 权限 | 版本 |
|------|------|------|------|------|
| /api/v1/devices | GET | 获取设备列表 | read | V1.0 |
| /api/v1/devices | POST | 注册设备 | write | V1.0 |
| /api/v1/devices/{id} | GET | 获取设备详情 | read | V1.0 |
| /api/v1/devices/{id} | PUT | 更新设备信息 | write | V1.0 |
| /api/v1/devices/{id} | DELETE | 删除设备 | admin | V1.0 |
| /api/v1/devices/{id}/data | GET | 获取设备数据 | read | V1.0 |
| /api/v1/heartbeat/status | GET | 获取心跳状态 | read | V1.1 |
| /api/v1/heartbeat/config | GET | 获取心跳配置 | read | V1.1 |
| /api/v1/heartbeat/config | PUT | 更新心跳配置 | admin | V1.1 |
| /api/v1/statistics/online | GET | 获取在线统计 | read | V1.1 |
| /api/v1/statistics/trend | GET | 获取趋势分析 | read | V1.1 |
| /api/v1/faults | GET | 获取故障列表 | read | V1.2 |
| /api/v1/faults/{id} | GET | 获取故障详情 | read | V1.2 |
| /api/v1/faults/devices/{id} | GET | 获取设备故障 | read | V1.2 |
| /api/v1/logs | GET | 获取日志列表 | read | V1.2 |
| /api/v1/logs/{id} | GET | 获取日志详情 | read | V1.2 |
| /api/v1/diagnosis/analyze | POST | 分析诊断 | read | V1.2 |
| /api/v1/parameters | GET | 获取参数列表 | read | V1.3 |
| /api/v1/parameters/{id} | GET | 获取参数详情 | read | V1.3 |
| /api/v1/devices/{id}/parameters | GET | 获取设备参数 | read | V1.3 |
| /api/v1/devices/{id}/parameters | PUT | 设置设备参数 | write | V1.3 |
| /api/v1/devices/{id}/parameters/batch | POST | 批量设置参数 | write | V1.3 |
| /api/v1/templates | GET | 获取参数模板 | read | V1.3 |
| /api/v1/templates | POST | 创建参数模板 | write | V1.3 |
| /api/v1/tenants | GET | 获取租户列表 | admin | V1.0 |
| /api/v1/tenants | POST | 创建租户 | admin | V1.0 |
| /api/v1/tenants/{id}/rules | GET | 获取转发规则 | read | V1.0 |
| /api/v1/tenants/{id}/rules | POST | 创建转发规则 | write | V1.0 |
| /api/v1/tenants/{id}/rules/{rule_id} | DELETE | 删除转发规则 | write | V1.0 |
| /api/v1/monitoring/metrics | GET | 获取监控指标 | read | V1.0 |
| /api/v1/monitoring/alerts | GET | 获取告警列表 | read | V1.0 |

#### 4.2.2 API响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "timestamp": 1704067200000
}
```

#### 4.2.3 设备管理API示例

```go
package api

import (
    "github.com/gin-gonic/gin"
    "sl651/device"
)

type DeviceController struct {
    deviceManager *device.DeviceManager
}

func NewDeviceController(dm *device.DeviceManager) *DeviceController {
    return &DeviceController{
        deviceManager: dm,
    }
}

func (dc *DeviceController) RegisterRoutes(r *gin.Engine) {
    v1 := r.Group("/api/v1")
    {
        devices := v1.Group("/devices")
        {
            devices.GET("", dc.ListDevices)
            devices.POST("", dc.CreateDevice)
            devices.GET("/:id", dc.GetDevice)
            devices.PUT("/:id", dc.UpdateDevice)
            devices.DELETE("/:id", dc.DeleteDevice)
            devices.GET("/:id/data", dc.GetDeviceData)
        }
    }
}

func (dc *DeviceController) ListDevices(c *gin.Context) {
    devices, err := dc.deviceManager.ListDevices(c.Request.Context())
    if err != nil {
        c.JSON(500, gin.H{"code": 500, "message": err.Error()})
        return
    }
    c.JSON(200, gin.H{
        "code":      0,
        "message":   "success",
        "data":      devices,
        "timestamp": time.Now().UnixMilli(),
    })
}

func (dc *DeviceController) CreateDevice(c *gin.Context) {
    var req device.CreateDeviceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"code": 400, "message": err.Error()})
        return
    }
    
    dev, err := dc.deviceManager.RegisterDevice(c.Request.Context(), &req)
    if err != nil {
        c.JSON(500, gin.H{"code": 500, "message": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{
        "code":      0,
        "message":   "success",
        "data":      dev,
        "timestamp": time.Now().UnixMilli(),
    })
}
```

### 4.3 WebSocket接口

#### 4.3.1 连接端点

```
ws://localhost:8080/ws
```

#### 4.3.2 消息格式

```json
{
  "type": "subscribe",
  "topic": "device.data",
  "device_id": "device_001"
}
```

```json
{
  "type": "data",
  "device_id": "device_001",
  "timestamp": 1704067200000,
  "payload": {}
}
```

#### 4.3.3 事件类型

| 类型 | 描述 |
|------|------|
| subscribe | 订阅主题 |
| unsubscribe | 取消订阅 |
| data | 数据推送 |
| event | 事件通知 |
| error | 错误消息 |

## 5. 数据模型设计

### 5.1 核心数据模型

#### 5.1.1 设备模型

```go
package model

import "time"

type Device struct {
    ID          string        `json:"id" gorm:"primaryKey"`
    Name        string        `json:"name" gorm:"not null"`
    DeviceType  DeviceType    `json:"device_type" gorm:"not null"`
    Protocol    Protocol      `json:"protocol" gorm:"not null"`
    Address     string        `json:"address" gorm:"not null"`
    Port        int           `json:"port" gorm:"not null"`
    Credentials Credentials   `json:"credentials" gorm:"type:text"`
    Status      DeviceStatus  `json:"status" gorm:"not null"`
    LastSeen    time.Time     `json:"last_seen" gorm:"not null"`
    Config      DeviceConfig  `json:"config" gorm:"type:text"`
    CreatedAt   time.Time     `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
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
    ProtocolSL651 Protocol = "SL651"
    ProtocolModbus Protocol = "Modbus"
    ProtocolCustom Protocol = "Custom"
)

type Credentials struct {
    AuthType    AuthType `json:"auth_type"`
    Username    string    `json:"username,omitempty"`
    Password    string    `json:"password,omitempty"`
    Certificate string    `json:"certificate,omitempty"`
}

type AuthType string

const (
    AuthTypeNone       AuthType = "none"
    AuthTypePassword   AuthType = "password"
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
```

#### 5.1.2 数据模型

```go
type DeviceData struct {
    ID        string      `json:"id" gorm:"primaryKey"`
    DeviceID  string      `json:"device_id" gorm:"not null;index"`
    Timestamp time.Time   `json:"timestamp" gorm:"not null;index"`
    DataType  DataType    `json:"data_type" gorm:"not null"`
    Values    []DataPoint `json:"values" gorm:"type:text"`
    Quality   DataQuality `json:"quality" gorm:"not null"`
    CreatedAt time.Time   `json:"created_at" gorm:"autoCreateTime"`
}

type DataPoint struct {
    Tag   string     `json:"tag"`
    Value DataValue  `json:"value"`
    Unit  string     `json:"unit"`
}

type DataValue struct {
    Float  *float64 `json:"float,omitempty"`
    Int    *int64   `json:"int,omitempty"`
    String *string  `json:"string,omitempty"`
    Bool   *bool    `json:"bool,omitempty"`
}

type DataQuality string

const (
    QualityGood      DataQuality = "good"
    QualityUncertain DataQuality = "uncertain"
    QualityBad       DataQuality = "bad"
)
```

#### 5.1.3 租户模型

```go
type Tenant struct {
    ID          string        `json:"id" gorm:"primaryKey"`
    Name        string        `json:"name" gorm:"not null;uniqueIndex"`
    Description string        `json:"description"`
    Config      TenantConfig  `json:"config" gorm:"type:text"`
    Status      TenantStatus  `json:"status" gorm:"not null"`
    CreatedAt   time.Time     `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
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
```

#### 5.1.4 转发规则模型

```go
type ForwardRule struct {
    ID          string       `json:"id" gorm:"primaryKey"`
    TenantID    string       `json:"tenant_id" gorm:"not null;index"`
    Name        string       `json:"name" gorm:"not null"`
    Enabled     bool         `json:"enabled" gorm:"not null"`
    Filter      RuleFilter   `json:"filter" gorm:"type:text"`
    Transform   RuleTransform `json:"transform" gorm:"type:text"`
    Destination Destination  `json:"destination" gorm:"type:text"`
    RetryPolicy RetryPolicy  `json:"retry_policy" gorm:"type:text"`
    CreatedAt   time.Time    `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

type RuleFilter struct {
    DeviceIDs  []string    `json:"device_ids"`
    DataTypes  []DataType  `json:"data_types"`
    Conditions []Condition `json:"conditions"`
}

type Condition struct {
    Field    string     `json:"field"`
    Operator Operator   `json:"operator"`
    Value    DataValue  `json:"value"`
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
    Mappings  []FieldMapping `json:"mappings"`
    Template  *string        `json:"template,omitempty"`
}

type FieldMapping struct {
    Source    string        `json:"source"`
    Target    string        `json:"target"`
    Transform *TransformFunc `json:"transform,omitempty"`
}

type TransformFunc string

const (
    TransformFuncNone     TransformFunc = "none"
    TransformFuncToString TransformFunc = "to_string"
    TransformFuncToInt    TransformFunc = "to_int"
    TransformFuncToFloat TransformFunc = "to_float"
    TransformFuncCustom  TransformFunc = "custom"
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
```

#### 5.1.5 心跳模型

```go
type Heartbeat struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    DeviceID  string    `json:"device_id" gorm:"not null;index"`
    Timestamp time.Time `json:"timestamp" gorm:"not null;index"`
    Status    string    `json:"status" gorm:"not null"`
    Latency   int       `json:"latency"`
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type HeartbeatConfig struct {
    Interval int `json:"interval"`
    Timeout  int `json:"timeout"`
}
```

#### 5.1.6 故障模型

```go
type Fault struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    DeviceID  string    `json:"device_id" gorm:"not null;index"`
    Type      FaultType `json:"type" gorm:"not null"`
    Severity  Severity  `json:"severity" gorm:"not null"`
    Message   string    `json:"message" gorm:"not null"`
    Timestamp time.Time `json:"timestamp" gorm:"not null;index"`
    Details   string    `json:"details" gorm:"type:text"`
    Resolved  bool      `json:"resolved" gorm:"default:false"`
    ResolvedAt *time.Time `json:"resolved_at,omitempty"`
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}
```

#### 5.1.7 参数模型

```go
type Parameter struct {
    ID        string        `json:"id" gorm:"primaryKey"`
    DeviceID  string        `json:"device_id" gorm:"not null;index"`
    Name      string        `json:"name" gorm:"not null"`
    Value     string        `json:"value" gorm:"not null"`
    Unit      string        `json:"unit"`
    Type      ParameterType `json:"type" gorm:"not null"`
    ReadOnly  bool          `json:"read_only" gorm:"not null"`
    UpdatedAt time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
    CreatedAt time.Time     `json:"created_at" gorm:"autoCreateTime"`
}

type ParameterType string

const (
    ParameterTypeInt    ParameterType = "int"
    ParameterTypeFloat  ParameterType = "float"
    ParameterTypeString ParameterType = "string"
    ParameterTypeBool   ParameterType = "bool"
)
```

### 5.2 数据库设计

#### 5.2.1 SQLite表结构

```sql
CREATE TABLE devices (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    device_type TEXT NOT NULL,
    protocol TEXT NOT NULL,
    address TEXT NOT NULL,
    port INTEGER NOT NULL,
    credentials TEXT NOT NULL,
    status TEXT NOT NULL,
    last_seen INTEGER NOT NULL,
    config TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE device_data (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    data_type TEXT NOT NULL,
    values TEXT NOT NULL,
    quality TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

CREATE INDEX idx_device_data_device_id ON device_data(device_id);
CREATE INDEX idx_device_data_timestamp ON device_data(timestamp);

CREATE TABLE tenants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    config TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE forward_rules (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    enabled INTEGER NOT NULL,
    filter TEXT NOT NULL,
    transform TEXT NOT NULL,
    destination TEXT NOT NULL,
    retry_policy TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE forward_logs (
    id TEXT PRIMARY KEY,
    rule_id TEXT NOT NULL,
    device_id TEXT NOT NULL,
    data_id TEXT NOT NULL,
    status TEXT NOT NULL,
    error_message TEXT,
    retry_count INTEGER NOT NULL,
    created_at INTEGER NOT NULL
);

CREATE INDEX idx_forward_logs_rule_id ON forward_logs(rule_id);
CREATE INDEX idx_forward_logs_created_at ON forward_logs(created_at);

CREATE TABLE heartbeats (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    status TEXT NOT NULL,
    latency INTEGER,
    created_at INTEGER NOT NULL
);

CREATE INDEX idx_heartbeats_device_id ON heartbeats(device_id);
CREATE INDEX idx_heartbeats_timestamp ON heartbeats(timestamp);

CREATE TABLE faults (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL,
    type TEXT NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    details TEXT,
    resolved INTEGER NOT NULL DEFAULT 0,
    resolved_at INTEGER,
    created_at INTEGER NOT NULL
);

CREATE INDEX idx_faults_device_id ON faults(device_id);
CREATE INDEX idx_faults_timestamp ON faults(timestamp);

CREATE TABLE parameters (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    unit TEXT,
    type TEXT NOT NULL,
    read_only INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL
);

CREATE INDEX idx_parameters_device_id ON parameters(device_id);
```

## 6. 部署方案

### 6.1 单体部署架构

```text
┌─────────────────────────────────────────────────────────────────────────┐
│                         单体应用部署                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                    sl651-platform (单一可执行文件)                  │  │
│  │                                                                   │  │
│  │  ┌─────────────────────────────────────────────────────────────┐ │  │
│  │  │                      所有模块                                  │ │  │
│  │  │  - 接入服务                                                  │ │  │
│  │  │  - 转发服务                                                  │ │  │
│  │  │  - 管理服务                                                  │ │  │
│  │  │  - Web服务                                                   │ │  │
│  │  │  - 数据库 (SQLite)                                            │ │  │
│  │  └─────────────────────────────────────────────────────────────┘ │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  部署特点：                                                              │
│  1. 单一可执行文件，无需依赖                                              │
│  2. 内置SQLite数据库，零配置                                              │
│  3. 资源占用低（<100MB内存）                                              │
│  4. 部署简单，直接运行                                                    │
│  5. 适合中小规模部署（<1万台设备）                                         │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 6.2 部署配置

#### 6.2.1 配置文件 (config.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"

database:
  path: "./data/platform.db"
  max_open_conns: 10
  max_idle_conns: 5

access:
  max_connections: 10000
  connection_timeout: 30
  heartbeat_interval: 60
  retry_count: 3

forward:
  max_queue_size: 10000
  worker_count: 4
  retry_interval: 5

logging:
  level: "info"
  file: "./logs/platform.log"
  max_size: 100
  max_backups: 10
  max_age: 30

monitoring:
  enabled: true
  metrics_port: 9090
```

#### 6.2.2 环境变量

```bash
# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_MODE=release

# 数据库配置
DATABASE_PATH=./data/platform.db

# 日志配置
LOG_LEVEL=info
LOG_FILE=./logs/platform.log

# 监控配置
MONITORING_ENABLED=true
METRICS_PORT=9090
```

### 6.3 性能指标

| 指标 | 目标值 | 说明 |
|------|--------|------|
| 内存占用 | <100MB | 空载时 |
| CPU占用 | <10% | 空载时 |
| 启动时间 | <5s | 冷启动 |
| 设备连接数 | 10000 | 最大并发连接 |
| 数据吞吐量 | 10000 msg/s | 消息处理能力 |
| 响应延迟 | <100ms | API响应时间 |
| 数据转发延迟 | <500ms | 端到端延迟 |

### 6.4 跨平台支持

| 平台 | 支持状态 | 备注 |
|------|----------|------|
| Windows x64 | 完全支持 | Windows 10+ |
| Windows ARM64 | 完全支持 | Windows 11+ |
| Linux x64 | 完全支持 | Ubuntu 20.04+, CentOS 8+ |
| Linux ARM64 | 完全支持 | 树莓派等嵌入式设备 |
| macOS x64 | 完全支持 | macOS 11+ |
| macOS ARM64 | 完全支持 | Apple Silicon |
| 麒麟OS | 完全支持 | 国产操作系统 |
| 统信UOS | 完全支持 | 国产操作系统 |

### 6.5 部署步骤

#### 6.5.1 Windows部署

```powershell
# 1. 下载可执行文件
# sl651-platform.exe

# 2. 创建配置目录
mkdir data
mkdir logs

# 3. 创建配置文件
# config.yaml

# 4. 运行程序
.\sl651-platform.exe

# 5. 安装为服务（可选）
sc create Sl651Platform binPath= "C:\sl651\sl651-platform.exe" start= auto
sc start Sl651Platform
```

#### 6.5.2 Linux部署

```bash
# 1. 下载可执行文件
# sl651-platform

# 2. 添加执行权限
chmod +x sl651-platform

# 3. 创建配置目录
mkdir -p data logs

# 4. 创建配置文件
# config.yaml

# 5. 运行程序
./sl651-platform

# 6. 配置systemd服务（可选）
cat > /etc/systemd/system/sl651-platform.service << EOF
[Unit]
Description=SL651 Platform
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/sl651
ExecStart=/opt/sl651/sl651-platform
Restart=always

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable sl651-platform
systemctl start sl651-platform
```

## 7. 安全设计

### 7.1 认证与授权

#### 7.1.1 JWT认证

```go
package auth

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

type AuthService struct {
    secretKey string
}

func NewAuthService(secretKey string) *AuthService {
    return &AuthService{
        secretKey: secretKey,
    }
}

func (as *AuthService) GenerateToken(userID, role string) (string, error) {
    now := time.Now()
    claims := Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(now),
            NotBefore: jwt.NewNumericDate(now),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(as.secretKey))
}

func (as *AuthService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(as.secretKey), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, ErrInvalidToken
}
```

#### 7.1.2 权限控制

```go
type Permission string

const (
    PermissionRead  Permission = "read"
    PermissionWrite Permission = "write"
    PermissionAdmin Permission = "admin"
)

type User struct {
    ID          string      `json:"id"`
    Username    string      `json:"username"`
    Permissions []Permission `json:"permissions"`
}

func CheckPermission(user *User, required Permission) bool {
    for _, perm := range user.Permissions {
        if perm == required || perm == PermissionAdmin {
            return true
        }
    }
    return false
}
```

### 7.2 数据加密

#### 7.2.1 TLS配置

```go
package tls

import (
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
)

func CreateTLSConfig(certPath, keyPath, caPath string) (*tls.Config, error) {
    cert, err := tls.LoadX509KeyPair(certPath, keyPath)
    if err != nil {
        return nil, err
    }
    
    caCert, err := ioutil.ReadFile(caPath)
    if err != nil {
        return nil, err
    }
    
    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)
    
    config := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        ClientCAs:    caCertPool,
        MinVersion:   tls.VersionTLS12,
    }
    
    return config, nil
}
```

### 7.3 审计日志

```go
type AuditLog struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    UserID    string    `json:"user_id" gorm:"not null;index"`
    Action    string    `json:"action" gorm:"not null"`
    Resource  string    `json:"resource" gorm:"not null"`
    Result    bool      `json:"result" gorm:"not null"`
    Timestamp time.Time `json:"timestamp" gorm:"not null;index"`
    IPAddress string    `json:"ip_address"`
}
```

## 8. 监控与运维

### 8.1 监控指标

#### 8.1.1 系统指标

| 指标名称 | 类型 | 描述 |
|----------|------|------|
| system_cpu_usage | Gauge | CPU使用率 |
| system_memory_usage | Gauge | 内存使用量 |
| system_disk_usage | Gauge | 磁盘使用量 |
| system_network_in | Counter | 网络入流量 |
| system_network_out | Counter | 网络出流量 |

#### 8.1.2 业务指标

| 指标名称 | 类型 | 描述 |
|----------|------|------|
| device_connections | Gauge | 设备连接数 |
| device_online_count | Gauge | 在线设备数 |
| device_offline_count | Gauge | 离线设备数 |
| data_messages_total | Counter | 数据消息总数 |
| data_messages_per_second | Gauge | 消息速率 |
| forward_success_total | Counter | 转发成功数 |
| forward_failure_total | Counter | 转发失败数 |
| forward_latency | Histogram | 转发延迟 |
| heartbeat_success_total | Counter | 心跳成功数 |
| heartbeat_failure_total | Counter | 心跳失败数 |
| fault_total | Counter | 故障总数 |

### 8.2 告警规则

```yaml
alerts:
  - name: "设备离线告警"
    condition: "device_offline_count > 100"
    duration: "5m"
    severity: "warning"
    
  - name: "转发失败告警"
    condition: "forward_failure_total / forward_success_total > 0.1"
    duration: "10m"
    severity: "critical"
    
  - name: "内存使用告警"
    condition: "system_memory_usage > 80%"
    duration: "5m"
    severity: "warning"
    
  - name: "心跳失败告警"
    condition: "heartbeat_failure_total / heartbeat_success_total > 0.2"
    duration: "10m"
    severity: "warning"
```

## 9. 技术方案总结

### 9.1 方案优势

| 优势项 | 说明 |
|--------|------|
| 轻量化 | 单体应用，资源占用低，部署简单 |
| 高性能 | Go实现，原生并发，支持1万台设备 |
| 跨平台 | 支持Windows/Linux/macOS及国产系统 |
| 易维护 | 单一技术栈，代码统一，降低维护成本 |
| 解耦设计 | 接入与转发模块解耦，易于扩展 |
| 安全可靠 | 类型安全，TLS加密，完善的认证授权 |
| 开发高效 | Go语法简洁，标准库丰富，开发周期短 |
| 生态完善 | 丰富的第三方库，社区活跃 |

### 9.2 技术栈总结

| 层次 | 技术选型 | 说明 |
|------|----------|------|
| 编程语言 | Go 1.21+ | 高性能、跨平台、并发友好 |
| Web框架 | Gin | 高性能HTTP框架 |
| ORM | GORM | 强大的ORM框架 |
| 数据库 | SQLite | 嵌入式、零配置 |
| 序列化 | encoding/json | 标准库JSON序列化 |
| 日志 | zap | 高性能结构化日志 |
| 配置管理 | Viper | 灵活的配置管理 |
| 监控 | Prometheus | 指标收集和监控 |
| 认证 | JWT | JSON Web Token |
| 网络协议 | net, crypto/tls | 标准库网络支持 |

### 9.3 迭代版本总结

| 版本 | 功能 | 周期 |
|------|------|------|
| V1.0 | 设备管理、数据采集、基础转发 | 2周 |
| V1.1 | 心跳管理、状态监控、告警通知 | 2周 |
| V1.2 | 故障诊断、日志分析、异常检测 | 2周 |
| V1.3 | 参数配置、参数读写、模板管理 | 2周 |
| V2.0 | 性能优化、集群部署、高级功能 | 4周 |

### 9.4 部署建议

1. **小规模部署（<1000台设备）**：单体应用，单机部署
2. **中等规模部署（1000-5000台设备）**：单体应用，负载均衡
3. **大规模部署（5000-10000台设备）**：单体应用，集群部署

### 9.5 扩展性考虑

1. **水平扩展**：通过负载均衡支持多实例部署
2. **垂直扩展**：增加服务器资源提升处理能力
3. **模块扩展**：插件化架构支持功能扩展
4. **协议扩展**：支持自定义协议插件

## 10. 附录

### 10.1 参考资料

- SL651-2014 水文监测数据通信规约
- Go官方文档
- Gin Web框架文档
- GORM文档

### 10.2 术语表

| 术语 | 说明 |
|------|------|
| SL651 | 水文监测数据通信规约国家标准 |
| RTU | 远程终端单元 |
| TLS | 传输层安全协议 |
| JWT | JSON Web Token |
| MQTT | 消息队列遥测传输协议 |
| GORM | Go对象关系映射框架 |

### 10.3 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| 2.0 | 2024-01-01 | 调整为Go技术栈，删除远程控制和固件升级，按迭代版本规划 |
| 1.0 | 2024-01-01 | 初始版本 |
