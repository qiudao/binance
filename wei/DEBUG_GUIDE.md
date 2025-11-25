# 🐛 Web 界面调试指南

## 问题：浏览器中看不到数据

### 步骤1: 确认服务器正常运行

```bash
cd /home/ubuntu/work/binance/wei

# 停止可能运行的旧进程
pkill -f web_server.go

# 启动服务器
make web-server
```

你应该看到：
```
2025/11/25 17:17:40 正在加载数据...
2025/11/25 17:17:40 ✓ 加载 klines_XBTUSD_1d.csv: 3714 条记录
2025/11/25 17:17:40 ✓ 加载 orders.csv: 43034 条记录
2025/11/25 17:17:41 ✓ 加载 executions.csv: 171578 条记录
2025/11/25 17:17:41 🌐 Web服务器启动成功!
2025/11/25 17:17:41    访问: http://localhost:8080
```

---

### 步骤2: 测试 API 是否工作

**在另一个终端运行：**

```bash
# 测试 K线 API
curl "http://localhost:8080/api/klines?symbol=XBTUSD&timeframe=1d" | head -c 500

# 测试账户 API
curl "http://localhost:8080/api/account"
```

如果返回 JSON 数据，说明 API 正常。

---

### 步骤3: 在浏览器中打开调试

1. **打开浏览器** (Chrome/Firefox)
2. **访问**: http://localhost:8080
3. **打开开发者工具**: 按 `F12` 或 `Ctrl+Shift+I`
4. **切换到 Console 标签页**

**你应该看到：**
```
=== BitMEX Trading Dashboard 初始化 ===
1. 检查 TradingView 库...
✓ TradingView 库已加载
2. 初始化图表...
✓ 图表初始化成功
3. 加载数据...
加载K线数据: XBTUSD 1d
  - 请求 API...
  - 收到 3714 条K线数据
  - 设置图表数据...
  ✓ K线图渲染完成
  - 加载订单标记...
  ✓ 订单标记完成
✓ 数据加载完成
4. 设置事件监听...
✓ 事件监听已设置
=== Dashboard 初始化完成! ===
```

---

### 步骤4: 检查 Network 请求

在开发者工具中：
1. 切换到 **Network** 标签页
2. 刷新页面 (`F5`)
3. 检查是否有以下请求：
   - `index.html` - 状态 200
   - `style.css` - 状态 200
   - `api.js`, `chart.js`, `orders.js`, `app.js` - 状态 200
   - `lightweight-charts.standalone.production.js` - 状态 200
   - `api/klines?symbol=XBTUSD&timeframe=1d` - 状态 200

如果有请求失败 (红色):
- 点击查看详情
- 查看 Response 或 Console 的错误信息

---

### 步骤5: 使用测试页面

访问简化的测试页面：
```
http://localhost:8080/test.html
```

这个页面会：
- 显示每个 API 的测试结果
- 不依赖 TradingView 图表库
- 直接显示数据是否正常

---

## 常见问题

### ❌ 问题1: TradingView 库未加载

**症状**: Console 显示 `LightweightCharts is undefined`

**原因**: CDN 加载失败或网络问题

**解决**:
```html
<!-- 方案1: 使用其他 CDN -->
<script src="https://cdn.jsdelivr.net/npm/lightweight-charts/dist/lightweight-charts.standalone.production.js"></script>

<!-- 方案2: 下载到本地 -->
# 下载库到本地
wget https://unpkg.com/lightweight-charts@4.2.1/dist/lightweight-charts.standalone.production.js -O web/lib/lightweight-charts.js

# 修改 index.html 引用
<script src="lib/lightweight-charts.js"></script>
```

---

### ❌ 问题2: K线图区域是空白

**症状**: 页面显示了，但图表区域空白

**原因**: CSS 高度问题或图表未正确渲染

**检查**:
```javascript
// 在 Console 中运行
console.log(chart);  // 应该显示图表对象
console.log(candlestickSeries);  // 应该显示系列对象
```

**解决**:
```bash
# 刷新页面
# 或调整窗口大小触发重绘
```

---

### ❌ 问题3: API 返回空数据

**症状**: Console 显示 "收到 0 条K线数据"

**检查数据文件**:
```bash
# 确认文件存在且有数据
ls -lh klines_XBTUSD_1d.csv
wc -l klines_XBTUSD_1d.csv

# 查看前几行
head klines_XBTUSD_1d.csv
```

**如果文件不存在或为空**:
```bash
# 重新下载
make download-klines
```

---

### ❌ 问题4: 端口被占用

**症状**: `bind: address already in use`

**解决**:
```bash
# 查找占用端口的进程
lsof -i :8080

# 杀死进程
kill -9 <PID>

# 或使用其他端口
# 编辑 web_server.go 第 157 行，改为 port := "3000"
```

---

## 调试命令速查

```bash
# 停止所有服务器
pkill -f web_server

# 启动服务器（前台）
make web-server

# 启动服务器（后台）
go run web_server.go > server.log 2>&1 &

# 查看日志
tail -f server.log

# 测试 API
curl "http://localhost:8080/api/klines?symbol=XBTUSD&timeframe=1d" | python3 -m json.tool | head

# 测试主页
curl -I "http://localhost:8080/"
```

---

## 完整重启流程

```bash
# 1. 停止所有
pkill -f web_server

# 2. 确认数据文件
ls -lh klines_XBTUSD_1d.csv orders.csv executions.csv wallet.csv

# 3. 启动服务器
make web-server

# 4. 在浏览器打开 (新标签页)
http://localhost:8080

# 5. 打开开发者工具 F12，查看 Console
```

---

## 如果还是不行...

提供以下信息：

1. **服务器日志**:
```bash
cat server.log
```

2. **浏览器 Console 截图** (F12 → Console)

3. **Network 请求状态** (F12 → Network → 刷新页面)

4. **数据文件信息**:
```bash
ls -lh *.csv
head -3 klines_XBTUSD_1d.csv
```
