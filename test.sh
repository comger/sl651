#!/bin/bash

echo "=== SL651 Platform V1.0 测试脚本 ==="

echo ""
echo "1. 检查依赖..."
if ! command -v go &> /dev/null; then
    echo "错误: Go 未安装"
    exit 1
fi

echo "Go 版本: $(go version)"

echo ""
echo "2. 创建必要的目录..."
mkdir -p data logs

echo ""
echo "3. 下载依赖..."
go mod download

echo ""
echo "4. 编译项目..."
go build -o bin/sl651-platform cmd/server/main.go
go build -o bin/importer cmd/importer/main.go

if [ $? -ne 0 ]; then
    echo "错误: 编译失败"
    exit 1
fi

echo "编译成功！"

echo ""
echo "5. 导入测试数据..."
if [ -f "testdata.csv" ]; then
    ./bin/importer testdata.csv
    
    if [ $? -eq 0 ]; then
        echo "数据导入成功！"
    else
        echo "警告: 数据导入失败"
    fi
else
    echo "警告: testdata.csv 文件不存在，跳过数据导入"
fi

echo ""
echo "6. 启动服务..."
echo "服务将在后台启动，日志输出到 logs/platform.log"
./bin/sl651-platform > logs/platform.log 2>&1 &
SERVER_PID=$!

echo "服务进程 ID: $SERVER_PID"

echo ""
echo "7. 等待服务启动..."
sleep 3

echo ""
echo "8. 测试健康检查..."
HEALTH_CHECK=$(curl -s http://localhost:8080/api/v1/health)
echo "健康检查响应: $HEALTH_CHECK"

echo ""
echo "9. 测试设备列表..."
DEVICES=$(curl -s http://localhost:8080/api/v1/devices)
echo "设备列表: $DEVICES"

echo ""
echo "10. 测试设备统计..."
STATS=$(curl -s http://localhost:8080/api/v1/statistics/devices)
echo "设备统计: $STATS"

echo ""
echo "11. 检查数据库..."
if [ -f "data/platform.db" ]; then
    echo "数据库文件存在: data/platform.db"
    DB_SIZE=$(du -h data/platform.db | cut -f1)
    echo "数据库大小: $DB_SIZE"
else
    echo "警告: 数据库文件不存在"
fi

echo ""
echo "12. 检查日志文件..."
if [ -f "logs/platform.log" ]; then
    echo "日志文件存在: logs/platform.log"
    LOG_SIZE=$(du -h logs/platform.log | cut -f1)
    echo "日志大小: $LOG_SIZE"
    echo ""
    echo "最后10行日志:"
    tail -10 logs/platform.log
else
    echo "警告: 日志文件不存在"
fi

echo ""
echo "=== 测试完成 ==="
echo ""
echo "服务正在运行，进程 ID: $SERVER_PID"
echo "要停止服务，请运行: kill $SERVER_PID"
echo ""
echo "访问 API 文档: http://localhost:8080/api/v1/health"
