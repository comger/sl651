@echo off
echo === SL651 Platform V1.0 完整测试脚本 ===
echo.

echo 1. 检查编译文件...
if not exist bin\sl651-platform.exe (
    echo 错误: 主服务程序不存在
    exit /b 1
)
if not exist bin\simulator.exe (
    echo 错误: 模拟器程序不存在
    exit /b 1
)
echo 编译文件检查通过！

echo.
echo 2. 清理旧数据...
if exist data\platform.db del data\platform.db
echo 旧数据已清理

echo.
echo 3. 启动主服务器 (端口8081)...
echo 服务器将在新窗口中启动
start "SL651 Server" cmd /k "bin\sl651-platform.exe"

echo.
echo 4. 等待服务器启动...
timeout /t 3 /nobreak

echo.
echo 5. 启动遥测终端模拟器 (连接端口8081)...
echo 模拟器将在新窗口中启动
start "SL651 Simulator" cmd /k "bin\simulator.exe"

echo.
echo === 系统已启动 ===
echo.
echo 主服务器窗口: SL651 Server (端口8081)
echo 模拟器窗口: SL651 Simulator (连接端口8081)
echo.
echo 观察两个窗口的日志输出：
echo - 模拟器会发送包含明文信息的报文
echo - 服务器会接收并解析报文，显示明文信息
echo.
echo 按 Ctrl+C 停止测试
echo.

pause
