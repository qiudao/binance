# 🔍 诊断功能使用指南

## 新功能：一键诊断按钮

现在在 Web 界面的右上角有一个 **🔍 Diagnose** 按钮，可以一键诊断所有问题！

---

## 🚀 快速开始

### 1. 启动服务器

```bash
cd /home/ubuntu/work/binance/wei
./START_WEB.sh
```

或者：
```bash
make web-server
```

### 2. 打开浏览器

访问: **http://localhost:8080/**

### 3. 点击诊断按钮

在页面右上角，点击 **🔍 Diagnose** 按钮

---

## 📊 诊断报告包含

### 1️⃣ Browser Environment（浏览器环境）
- 浏览器类型和版本
- 在线状态
- Cookie 是否启用
- 屏幕分辨率
- 窗口大小

### 2️⃣ API Connectivity（API 连接）
- ✅ K线 API 测试（显示数据条数和响应时间）
- ✅ 订单 API 测试（未成交订单数量）
- ✅ 成交 API 测试（成交记录数量）
- ✅ 账户 API 测试（余额信息）

### 3️⃣ Data Availability（数据可用性）
- K线数据是否加载
- 数据时间范围
- 数据条数

### 4️⃣ TradingView Library（图表库）
- ✅ 库是否加载
- 版本信息
- 图表实例是否创建
- 蜡烛图系列是否创建

### 5️⃣ Console Logs（控制台日志）
- 最近的 Console 输出
- 错误信息

### 6️⃣ Recommendations（建议）
- 根据检测到的问题给出具体建议
- 快速修复方案

---

## 🎯 诊断报告特点

### ✅ 所有检查项都有颜色标识
- 🟢 **绿色**: 正常工作
- 🔴 **红色**: 有错误
- 🟡 **黄色**: 警告
- ⚪ **灰色**: 信息

### 📋 一键复制报告
点击底部的 **📋 Copy Full Report** 按钮，可以：
- 复制完整的诊断报告到剪贴板
- 分享给开发者或技术支持
- 保存到文件用于后续分析

---

## 🐛 常见问题诊断

### 问题1: K线图不显示

**诊断步骤：**
1. 点击 🔍 Diagnose
2. 查看 **2️⃣ API Connectivity** 部分
3. 检查 "Klines API" 是否显示绿色 ✓

**如果显示红色 ✗:**
- 数据文件可能缺失
- 运行 `make download-klines` 下载数据
- 重启服务器

### 问题2: TradingView 库未加载

**诊断步骤：**
1. 点击 🔍 Diagnose
2. 查看 **4️⃣ TradingView Library** 部分
3. 检查 "Library Loaded" 是否为 ✓ Yes

**如果显示 ✗ No:**
- 网络连接问题
- CDN 加载失败
- 刷新页面重试
- 或检查浏览器 Console (F12) 查看具体错误

### 问题3: 数据加载慢

**诊断步骤：**
1. 点击 🔍 Diagnose
2. 查看 **2️⃣ API Connectivity** 中的响应时间
3. 如果显示 `(XXXms)`，数字越小越好

**如果响应时间 > 1000ms:**
- 数据量可能很大
- 服务器负载高
- 这是正常的，等待加载完成即可

---

## 📸 诊断报告示例

```
═══════════════════════════════════════
BitMEX Trading Dashboard - Diagnostic Report
═══════════════════════════════════════

1️⃣ BROWSER ENVIRONMENT
────────────────────────────────────────
Browser: Chrome ✓
Online Status: Online ✓
Cookies Enabled: Yes ✓
Screen Resolution: 1920x1080
Window Size: 1680x937

2️⃣ API CONNECTIVITY
────────────────────────────────────────
Klines API: ✓ 3714 records (45ms)
Orders API: ✓ 32 pending orders
Executions API: ✓ 100 executions
Account API: ✓ Balance: 26.8078 BTC

3️⃣ DATA AVAILABILITY
────────────────────────────────────────
K-line Data: ✓ 3714 candles available
Details: First: 9/26/2015, Last: 11/25/2025

4️⃣ TRADINGVIEW LIBRARY
────────────────────────────────────────
Library Loaded: ✓ Yes
Version: 4.2.1
Chart Instance: ✓ Created
Candlestick Series: ✓ Created

6️⃣ RECOMMENDATIONS
────────────────────────────────────────
✅ All systems operational!
```

---

## 💡 使用技巧

### 技巧1: 首次访问时立即诊断
```
1. 打开页面
2. 立即点击 🔍 Diagnose
3. 查看是否所有项都是绿色 ✓
4. 如果有红色 ✗，根据建议修复
```

### 技巧2: 遇到问题时诊断
```
如果 K线图不显示、数据不加载等：
1. 不要刷新页面
2. 先点击 🔍 Diagnose
3. 查看具体哪个环节出错
4. 根据建议解决问题
```

### 技巧3: 分享诊断报告
```
如果需要技术支持：
1. 点击 🔍 Diagnose
2. 点击 📋 Copy Full Report
3. 粘贴到邮件或问题报告中
4. 开发者可以快速定位问题
```

---

## 🔧 高级调试

### 配合浏览器开发者工具使用

1. **诊断面板** (在页面内)
   - 快速查看系统状态
   - 用户友好的界面

2. **开发者工具** (F12 → Console)
   - 查看详细的技术日志
   - 看到实时的错误信息
   - 适合技术人员

**最佳实践：**
- 先用诊断面板快速判断
- 如果有问题，打开 F12 查看详细错误
- 两者结合使用效果最好

---

## 📞 仍需帮助？

如果诊断后仍有问题：

1. **复制诊断报告**
   - 点击 📋 Copy Full Report

2. **复制浏览器 Console 日志**
   - 按 F12
   - 切换到 Console 标签
   - 右键 → Save as...

3. **提供信息**
   - 诊断报告
   - Console 日志
   - 问题描述
   - 截图

---

## ✨ 祝使用愉快！

现在你可以轻松诊断所有问题，不需要命令行就能排查！🎉
