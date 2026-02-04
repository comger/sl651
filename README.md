# SL651 Platform V1.0

基于SL651协议的水文设备统一运维平台 V1.0 版本

## 功能特性

- 设备管理：设备注册、查询、更新、删除
- SL651协议解析：支持ASCII/HEX/BCD编码
- 数据采集：实时数据采集和存储
- 数据转发：支持HTTP转发
- REST API：完整的设备管理和数据查询接口
- 数据持久化：SQLite数据库存储

## 技术栈

- Go 1.21+
- Gin Web框架
- GORM ORM
- SQLite数据库
- Viper配置管理

## 项目结构

```
sl651-platform/
├── cmd/
│   ├── server/          # 主服务
│   └── importer/        # CSV数据导入工具
├── internal/
│   ├── config/          # 配置管理
│   ├── database/        # 数据库初始化
│   ├── device/          # 设备管理
│   ├── forward/         # 数据转发
│   ├── http/            # HTTP服务
│   ├── model/           # 数据模型
│   ├── sl651/           # SL651协议解析
│   └── storage/         # 数据存储
├── config.yaml           # 配置文件
├── go.mod
└── README.md
```

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置

编辑 `config.yaml` 文件：

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"

database:
  path: "./data/platform.db"

access:
  max_connections: 10000

forward:
  max_queue_size: 10000
  worker_count: 4

logging:
  level: "info"
  file: "./logs/platform.log"
```

### 3. 编译项目

```bash
make build
```

或手动编译：

```bash
go build -o bin/sl651-platform cmd/server/main.go
go build -o bin/importer cmd/importer/main.go
```

### 4. 导入测试数据

```bash
make import
```

或手动导入：

```bash
go run cmd/importer/main.go testdata.csv
```

### 5. 启动服务

```bash
make run
```

或手动启动：

```bash
go run cmd/server/main.go
```

服务将在 http://localhost:8080 启动

### 6. 运行测试

#### Linux/Mac

```bash
make test-full
```

或手动运行：

```bash
./test.sh
```

#### Windows

```cmd
test.bat
```

## API接口

### 设备管理

| 路径 | 方法 | 描述 |
|------|------|------|
| /api/v1/devices | GET | 获取设备列表 |
| /api/v1/devices | POST | 注册设备 |
| /api/v1/devices/{id} | GET | 获取设备详情 |
| /api/v1/devices/{id} | PUT | 更新设备信息 |
| /api/v1/devices/{id} | DELETE | 删除设备 |
| /api/v1/devices/{id}/data | GET | 获取设备数据 |
| /api/v1/devices/{id}/data/latest | GET | 获取最新设备数据 |

### 统计信息

| 路径 | 方法 | 描述 |
|------|------|------|
| /api/v1/statistics/devices | GET | 获取设备统计 |
| /api/v1/statistics/forward | GET | 获取转发统计 |

### 健康检查

| 路径 | 方法 | 描述 |
|------|------|------|
| /api/v1/health | GET | 健康检查 |

## API示例

### 获取设备列表

```bash
curl http://localhost:8080/api/v1/devices
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "1090330853",
      "name": "站点-1090330853",
      "device_type": "RTU",
      "protocol": "SL651",
      "address": "127.0.0.1",
      "port": 8080,
      "status": "online",
      "last_seen": "2026-02-02T09:41:01Z",
      "created_at": "2026-02-02T09:41:01Z",
      "updated_at": "2026-02-02T09:41:01Z"
    }
  ],
  "timestamp": 1706869261000
}
```

### 获取设备数据

```bash
curl http://localhost:8080/api/v1/devices/1090330853/data?limit=10
```

### 获取设备统计

```bash
curl http://localhost:8080/api/v1/statistics/devices
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 1,
    "online": 1,
    "offline": 0
  },
  "timestamp": 1706869261000
}
```

## SL651协议支持

### 支持的功能码

- 0x2F: 链路维护
- 0x31: 定时报
- 0x32: 加报
- 0x34: 小时报

### 支持的数据类型

- 实时数据
- 告警数据
- 状态数据

### 支持的编码格式

- ASCII
- HEX
- BCD

## 数据格式

### CSV数据格式

testdata.csv 文件格式：

```
站点ID,站点名称,报文类型,功能码,报文内容,数据采集时间,数据接收时间
1090330853,站点名称,链路维护,2f,7e7e011090330853a0002f000802009f26020209410003bde3,2026-02-02 09:41:01,2026-02-02 09:41:01
```

### 设备数据格式

```json
{
  "id": "device_data_id",
  "device_id": "1090330853",
  "timestamp": "2026-02-02T09:41:01Z",
  "data_type": "realtime",
  "quality": "good",
  "values": [
    {
      "tag": "water_level",
      "value": {
        "float": 123.45
      },
      "unit": "m"
    }
  ]
}
```

## 性能指标

- 内存占用：<100MB（空载）
- CPU占用：<10%（空载）
- 启动时间：<5s
- 设备连接数：10000（最大并发）
- 数据吞吐量：10000 msg/s
- 响应延迟：<100ms（API）

## 部署

### Windows部署

```powershell
# 编译
go build -o sl651-platform.exe cmd/server/main.go

# 运行
.\sl651-platform.exe
```

### Linux部署

```bash
# 编译
go build -o sl651-platform cmd/server/main.go

# 运行
./sl651-platform
```

### Docker部署

```bash
# 构建镜像
docker build -t sl651-platform:v1.0 .

# 运行容器
docker run -d -p 8080:8080 -v $(pwd)/data:/app/data -v $(pwd)/logs:/app/logs sl651-platform:v1.0
```

## 测试

### 运行测试

```bash
go test ./...
```

### 导入测试数据

```bash
go run cmd/importer/main.go testdata.csv
```

## 故障排查

### 数据库锁定

如果遇到数据库锁定错误，请确保只有一个实例在运行。

### 端口占用

如果端口8080被占用，请修改config.yaml中的端口配置。

### 内存不足

如果遇到内存不足，请调整config.yaml中的max_connections和max_queue_size参数。

## 版本历史

### V1.0 (2024-01-01)

- 设备管理
- SL651协议解析
- 数据采集和存储
- 基础数据转发
- REST API

## 许可证

MIT License

## 联系方式

如有问题，请联系开发团队。
