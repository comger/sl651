@echo off
echo === SL651 Platform V1.0 测试脚本 ===
echo.

echo 1. 检查依赖...
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: Go 未安装
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
echo Go 版本: %GO_VERSION%

echo.
echo 2. 创建必要的目录...
if not exist data mkdir data
if not exist logs mkdir logs
if not exist bin mkdir bin

echo.
echo 3. 下载依赖...
go mod download

echo.
echo 4. 编译项目...
go build -o bin\sl651-platform.exe cmd\server\main.go
go build -o bin\importer.exe cmd\importer\main.go

if %errorlevel% neq 0 (
    echo 错误: 编译失败
    exit /b 1
)

echo 编译成功！

echo.
echo 5. 导入测试数据...
if exist testdata.csv (
    bin\importer.exe testdata.csv
    
    if %errorlevel% equ 0 (
        echo 数据导入成功！
    ) else (
        echo 警告: 数据导入失败
    )
) else (
    echo 警告: testdata.csv 文件不存在，跳过数据导入
)

echo.
echo 6. 启动服务...
echo 服务将在后台启动，日志输出到 logs\platform.log
start /B bin\sl651-platform.exe > logs\platform.log 2>&1

echo.
echo 7. 等待服务启动...
timeout /t 5 /nobreak

echo.
echo 8. 测试健康检查...
curl -s http://localhost:8080/api/v1/health
echo.

echo 9. 测试设备列表...
curl -s http://localhost:8080/api/v1/devices
echo.

echo 10. 测试设备统计...
curl -s http://localhost:8080/api/v1/statistics/devices
echo.

echo 11. 检查数据库...
if exist data\platform.db (
    echo 数据库文件存在: data\platform.db
) else (
    echo 警告: 数据库文件不存在
)

echo.
echo 12. 检查日志文件...
if exist logs\platform.log (
    echo 日志文件存在: logs\platform.log
    echo.
    echo 最后10行日志:
    powershell -Command "Get-Content logs\platform.log -Tail 10"
) else (
    echo 警告: 日志文件不存在
)

echo.
echo === 测试完成 ===
echo.
echo 服务正在运行...
echo 要停止服务，请关闭命令窗口或按 Ctrl+C
echo.
echo 访问 API 文档: http://localhost:8080/api/v1/health

pause
