#!/bin/bash

cd /home/ubuntu/work/binance/wei

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🚀 启动 BitMEX Trading Dashboard"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 停止旧进程
echo "1. 清理旧进程..."
lsof -ti :8080 | xargs -r kill -9 2>/dev/null
sleep 1
echo "   ✓ 端口 8080 已清理"
echo ""

# 检查数据文件
echo "2. 检查数据文件..."
for file in klines_XBTUSD_1d.csv orders.csv executions.csv wallet.csv; do
    if [ -f "$file" ]; then
        echo "   ✓ $file"
    else
        echo "   ✗ $file - 缺失!"
    fi
done
echo ""

# 启动服务器
echo "3. 启动 Web 服务器..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "   📍 访问地址:"
echo ""
echo "      主界面: http://localhost:8080/"
echo "      测试页: http://localhost:8080/test.html"
echo ""
echo "   🔍 新功能: 点击右上角 '🔍 Diagnose' 按钮"
echo "              可以诊断所有问题并显示详细报告"
echo ""
echo "   ⏹  停止服务器: 按 Ctrl+C"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

go run web_server.go
