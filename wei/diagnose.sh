#!/bin/bash

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🔍 BitMEX Trading Dashboard 诊断工具"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 1. 检查数据文件
echo "1️⃣  检查数据文件..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
for file in klines_XBTUSD_1d.csv orders.csv executions.csv wallet.csv; do
    if [ -f "$file" ]; then
        lines=$(wc -l < "$file")
        size=$(ls -lh "$file" | awk '{print $5}')
        echo "✓ $file - $lines 行, $size"
    else
        echo "✗ $file - 文件不存在!"
    fi
done
echo ""

# 2. 检查端口占用
echo "2️⃣  检查端口 8080..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if lsof -i :8080 > /dev/null 2>&1; then
    echo "⚠️  端口 8080 被占用:"
    lsof -i :8080
    echo ""
    read -p "是否要停止占用的进程? (y/N) " answer
    if [[ "$answer" == "y" || "$answer" == "Y" ]]; then
        pkill -f web_server
        echo "✓ 已停止进程"
    fi
else
    echo "✓ 端口 8080 可用"
fi
echo ""

# 3. 启动服务器
echo "3️⃣  启动 Web 服务器..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
go run web_server.go > server_test.log 2>&1 &
SERVER_PID=$!
echo "服务器 PID: $SERVER_PID"
echo "等待启动..."
sleep 3

# 4. 测试 API
echo ""
echo "4️⃣  测试 API 端点..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 测试主页
echo -n "主页 (index.html): "
if curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/" | grep -q "200"; then
    echo "✓ OK (200)"
else
    echo "✗ 失败"
fi

# 测试 K线 API
echo -n "K线 API: "
response=$(curl -s "http://localhost:8080/api/klines?symbol=XBTUSD&timeframe=1d")
count=$(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(len(data))" 2>/dev/null)
if [ ! -z "$count" ] && [ "$count" -gt 0 ]; then
    echo "✓ OK ($count 条数据)"
else
    echo "✗ 失败或无数据"
fi

# 测试订单 API
echo -n "订单 API: "
response=$(curl -s "http://localhost:8080/api/orders/pending")
count=$(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(len(data))" 2>/dev/null)
if [ ! -z "$count" ]; then
    echo "✓ OK ($count 条数据)"
else
    echo "✗ 失败"
fi

# 测试账户 API
echo -n "账户 API: "
response=$(curl -s "http://localhost:8080/api/account")
if echo "$response" | python3 -c "import sys, json; json.load(sys.stdin)" > /dev/null 2>&1; then
    echo "✓ OK"
    echo "$response" | python3 -m json.tool | head -10
else
    echo "✗ 失败"
fi

echo ""

# 5. 显示服务器日志
echo "5️⃣  服务器日志..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat server_test.log
echo ""

# 6. 提供下一步建议
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  📋 诊断完成"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📍 服务器正在运行，访问以下地址："
echo ""
echo "   主界面: http://localhost:8080/"
echo "   测试页: http://localhost:8080/test.html"
echo ""
echo "🔧 在浏览器中:"
echo "   1. 打开上述网址"
echo "   2. 按 F12 打开开发者工具"
echo "   3. 查看 Console 标签的输出"
echo ""
echo "⚠️  服务器在后台运行 (PID: $SERVER_PID)"
echo "   停止服务器: kill $SERVER_PID"
echo "   或运行: pkill -f web_server"
echo ""
