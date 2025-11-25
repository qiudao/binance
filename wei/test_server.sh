#!/bin/bash

echo "启动 Web 服务器测试..."
echo ""

# 在后台启动服务器
go run web_server.go > server.log 2>&1 &
SERVER_PID=$!

echo "服务器 PID: $SERVER_PID"
echo "等待服务器启动..."
sleep 3

# 测试 API 端点
echo ""
echo "测试 API 端点:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 测试 K线数据
echo "1. 测试 K线数据 API..."
curl -s "http://localhost:8080/api/klines?symbol=XBTUSD&timeframe=1d" | head -c 200
echo "..."
echo ""

# 测试订单数据
echo "2. 测试订单数据 API..."
curl -s "http://localhost:8080/api/orders/pending" | head -c 200
echo "..."
echo ""

# 测试成交数据
echo "3. 测试成交数据 API..."
curl -s "http://localhost:8080/api/executions" | head -c 200
echo "..."
echo ""

# 测试仓位数据
echo "4. 测试仓位数据 API..."
curl -s "http://localhost:8080/api/positions" | head -c 200
echo "..."
echo ""

# 测试账户数据
echo "5. 测试账户数据 API..."
curl -s "http://localhost:8080/api/account"
echo ""
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "服务器日志:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat server.log
echo ""

# 停止服务器
echo "停止服务器..."
kill $SERVER_PID 2>/dev/null
sleep 1

echo ""
echo "✓ 测试完成！"
echo ""
echo "要启动服务器，运行: make web-server"
echo "或查看完整日志: cat server.log"
