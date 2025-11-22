# Binance K线数据获取工具

使用 Golang 获取币安交易所历史 1 分钟 K 线数据。

## 功能

- 获取指定交易对的历史 K 线数据
- 支持 1 分钟时间间隔
- 可指定时间范围或数量限制

## 使用方法

### 运行程序

```bash
go run main.go
```

### 参数说明

在 `main.go` 中可以修改以下参数：

- `symbol`: 交易对，例如 "BTCUSDT", "ETHUSDT"
- `interval`: K线间隔
  - `1m` - 1分钟
  - `5m` - 5分钟
  - `15m` - 15分钟
  - `1h` - 1小时
  - `1d` - 1天
- `limit`: 返回的K线数量（最大1000）
- `startTime`: 开始时间（毫秒时间戳）
- `endTime`: 结束时间（毫秒时间戳）

### 示例

#### 1. 获取最近 100 条 1 分钟 K 线

```go
klines, err := GetKlines("BTCUSDT", "1m", 0, 0, 100)
```

#### 2. 获取指定时间范围的 K 线

```go
endTime := time.Now().UnixMilli()
startTime := endTime - 24*60*60*1000 // 过去24小时
klines, err := GetKlines("BTCUSDT", "1m", startTime, endTime, 0)
```

## K线数据结构

每条 K 线包含以下字段：

- `OpenTime`: 开盘时间
- `Open`: 开盘价
- `High`: 最高价
- `Low`: 最低价
- `Close`: 收盘价
- `Volume`: 成交量
- `CloseTime`: 收盘时间
- `QuoteAssetVolume`: 成交额
- `NumberOfTrades`: 成交笔数
- `TakerBuyBaseAssetVolume`: 主动买入成交量
- `TakerBuyQuoteAssetVolume`: 主动买入成交额

## API 限制

- 单次请求最多返回 1000 条数据
- 注意 API 访问频率限制
# binance
