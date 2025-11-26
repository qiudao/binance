# 每日 BTC 仓位比例分析

## 功能说明

这个工具用于计算和分析 BitMEX 账户每日的 BTC 仓位比例。

### 仓位比例定义

- **Long BTC 0.2x**: BTC 多头仓位价值是账户总余额的 0.2 倍
- **Short BTC 0.8x**: BTC 空头仓位价值是账户总余额的 0.8 倍(显示为 -0.8)
- **Flat**: 无持仓(0x)

### 计算公式

对于 BitMEX 反向合约(如 XBTUSD):

```
仓位价值(BTC) = |持仓数量| / 当前价格
仓位比例 = 仓位价值 / 账户余额

Long: 正值 (+)
Short: 负值 (-)
```

## 使用方法

### 1. 生成每日仓位数据

```bash
make daily-position
```

这将:
- 读取 `executions.csv`, `wallet.csv`, `klines_XBTUSD_1d.csv`
- 计算从 2020-05-01 至今每天的仓位比例
- 生成 `daily_position.csv` 文件

### 2. 查看仓位报告

```bash
make view-position
```

显示:
- 总体统计(Long/Short/空仓天数和比例)
- 平均仓位倍数
- 最近 30 天仓位变化

### 3. 直接查看 CSV 文件

```bash
cat daily_position.csv
```

CSV 格式:
```
Date,PositionQty,Price,Balance,PositionValue,PositionRatio,Side
2025-11-25,14610400,88228.60,26.80720515,165.59709663,6.1773,Long
```

字段说明:
- `Date`: 日期
- `PositionQty`: 持仓数量(合约张数)
- `Price`: 当日收盘价
- `Balance`: 账户余额(BTC)
- `PositionValue`: 仓位价值(BTC)
- `PositionRatio`: 仓位比例(倍数)
- `Side`: 方向(Long/Short/Flat)

## 统计示例

```
📈 总体统计:
  总天数: 2034
  Long 天数: 1380 (67.8%)
  Short 天数: 418 (20.6%)
  空仓天数: 236 (11.6%)

  平均 Long 倍数: 7.26x
  平均 Short 倍数: 1.08x
```

## 历史极值

```
🟢 最大 Long 仓位:
   日期: 2023-10-24
   倍数: 51.18x
   BTC价格: $33,122
   账户余额: 8.91 BTC

🔴 最大 Short 仓位:
   日期: 2020-07-17
   倍数: 29.41x
   BTC价格: $9,130
   账户余额: 1.19 BTC
```

## 注意事项

1. **过滤 Funding 记录**: 程序自动过滤资金费率结算记录,只统计真实交易
2. **反向合约**: 正确处理 BitMEX 反向合约的仓位价值计算
3. **数据完整性**: 需要完整的成交记录、钱包记录和 K 线数据

## 文件说明

- `daily_position.go`: 核心计算程序
- `daily_position.csv`: 输出的每日仓位数据
- `view_position.sh`: 查看报告的脚本
- `Makefile`: 集成的命令入口

## 后续扩展

可以基于生成的 CSV 数据:
- 绘制仓位比例趋势图
- 分析仓位与盈亏的关系
- 研究仓位管理策略
- 集成到 Web 界面的时间旅行功能中
