# BitMEX 钱包历史记录下载工具

这个工具用于下载BitMEX账户的所有钱包历史记录（资金变动记录），并保存为CSV格式。可以追踪账户总资产随时间的变化。

## 功能特性

- ✅ 完整下载所有钱包历史记录
- ✅ 增量更新模式（只下载新记录）
- ✅ 自动分页处理
- ✅ 防止重复记录
- ✅ 同时提供 Satoshi 和 BTC 单位
- ✅ 详细记录信息（17个字段）
- ✅ API速率限制保护
- ✅ **显示每个时间点的账户总资产余额**

## 与交易记录的区别

| 特性 | 交易记录 (Executions) | 钱包历史 (Wallet History) |
|------|---------------------|-------------------------|
| **数据内容** | 交易执行的详细信息 | 所有导致资金变动的操作 |
| **包含信息** | 买卖价格、数量、手续费等 | 存款、提款、已实现盈亏、手续费、转账等 |
| **关键指标** | 成交价格、成交量 | **账户总余额（WalletBalance）** |
| **用途** | 分析交易明细 | **追踪账户总资产变化** |

## 使用方法

### 方式一：使用 Makefile（推荐）

```bash
# 查看帮助
make help

# 首次下载所有钱包历史记录
make wallet-download

# 增量更新（只下载新记录）
make wallet-update

# 列出所有CSV文件
make wallet-list

# 备份最新的CSV文件
make wallet-backup

# 删除所有CSV文件
make wallet-clean
```

### 方式二：直接运行 Go 程序

#### 1. 首次下载（全量）

首次使用时，下载所有钱包历史记录：

```bash
go run bitmex_wallet.go
```

这将创建一个新的CSV文件，文件名格式为：`bitmex_wallet_YYYYMMDD_HHMMSS.csv`

#### 2. 增量更新

后续有新记录时，使用增量更新模式只下载新记录：

```bash
go run bitmex_wallet.go --update
```

增量更新功能：
- 自动查找最新的CSV文件
- 读取最后一条记录的时间戳
- 只下载该时间之后的新记录
- 追加到现有文件（不创建新文件）
- 自动过滤重复记录

## CSV文件字段说明

CSV文件包含以下字段（同时提供 Satoshi 和 BTC 单位）：

| 字段名 | 说明 |
|--------|------|
| **TransactID** | 交易ID |
| **TransactType** | 交易类型（见下表） |
| **TransactStatus** | 交易状态（Completed/Pending/Canceled） |
| **Account** | 账户编号 |
| **Currency** | 货币（XBt=Bitcoin） |
| **Amount_Satoshi** | 变动金额（Satoshi单位） |
| **Amount_BTC** | 变动金额（BTC单位，保留8位小数） |
| **Fee_Satoshi** | 手续费（Satoshi单位） |
| **Fee_BTC** | 手续费（BTC单位） |
| **Timestamp** | 时间戳 |
| **WalletBalance_Satoshi** | ⭐ **钱包余额（Satoshi）** |
| **WalletBalance_BTC** | ⭐ **钱包余额（BTC）** - 关键字段 |
| **MarginBalance_Satoshi** | 保证金余额（Satoshi） |
| **MarginBalance_BTC** | 保证金余额（BTC） |
| **Address** | 地址（充提币时使用） |
| **Tx** | 交易哈希（充提币时使用） |
| **Text** | 备注信息 |

### 交易类型说明

| TransactType | 含义 |
|--------------|------|
| **RealisedPNL** | 已实现盈亏（平仓产生） |
| **UnrealisedPNL** | 未实现盈亏调整 |
| **Deposit** | 存款（充币） |
| **Withdrawal** | 提款（提币） |
| **Transfer** | 转账 |
| **AffiliatePayout** | 推荐佣金 |
| **RealisedPNL** | 交易手续费（通常包含在RealisedPNL中） |

## 示例输出

```
=== BitMEX 钱包历史记录下载工具 ===

测试API连接...
✓ API连接成功!

运行模式: 全量下载

开始下载所有钱包历史记录...
正在获取记录 0-10000...
  已获取 8432 条记录，总计: 8432 条
✓ 已到达最后一页

总共下载了 8432 条钱包历史记录

第一条记录:
  时间: 2017-03-15T08:23:45.123Z
  类型: Deposit
  金额: 0.50000000 BTC
  钱包余额: 0.50000000 BTC

最后一条记录:
  时间: 2025-11-25T12:30:15.456Z
  类型: RealisedPNL
  金额: 0.00125000 BTC
  钱包余额: 2.35678901 BTC

正在保存到文件: bitmex_wallet_20251125_123456.csv
✓ 成功保存 8432 条记录到 bitmex_wallet_20251125_123456.csv
```

## 数据分析示例

使用CSV文件可以进行以下分析：

### 1. 查看账户总资产变化趋势
```bash
# 使用 awk 提取时间和钱包余额
awk -F',' 'NR>1 {print $10, $12}' bitmex_wallet_*.csv
```

### 2. 计算总盈亏
```bash
# 统计所有 RealisedPNL 类型的金额总和
awk -F',' 'NR>1 && $2=="RealisedPNL" {sum+=$7} END {print "总盈亏:", sum, "BTC"}' bitmex_wallet_*.csv
```

### 3. 统计充提币记录
```bash
# 统计存款
awk -F',' 'NR>1 && $2=="Deposit" {count++; sum+=$7} END {print "存款次数:", count, "总额:", sum, "BTC"}' bitmex_wallet_*.csv

# 统计提款
awk -F',' 'NR>1 && $2=="Withdrawal" {count++; sum+=$7} END {print "提款次数:", count, "总额:", sum, "BTC"}' bitmex_wallet_*.csv
```

### 4. 在 Excel 中分析
直接在 Excel 或 Google Sheets 中打开 CSV 文件，可以：
- 创建资产变化曲线图（Timestamp vs WalletBalance_BTC）
- 使用数据透视表按类型统计
- 计算每日/每月盈亏
- 分析手续费支出

## 注意事项

1. **API限流**：程序已内置500ms延迟，避免触发API速率限制
2. **数据完整性**：增量更新会自动过滤重复记录
3. **单位换算**：1 BTC = 100,000,000 Satoshi
4. **下载时间**：首次全量下载时间取决于记录数量（通常1-2分钟）
5. **增量更新**：通常几秒钟完成
6. **负数金额**：提款、手续费等会显示为负数
7. **WalletBalance**：这是最重要的字段，显示每个时间点的账户总资产

## 技术细节

- 使用BitMEX REST API v1
- 接口：`GET /api/v1/user/walletHistory`
- 支持HMAC-SHA256签名认证
- 每次请求最多10000条记录（API默认值）
- 自动处理分页
- 使用Go标准库，无外部依赖

## 常见问题

**Q: 钱包历史和交易记录有什么区别？**
A: 交易记录只包含交易执行信息，钱包历史包含所有资金变动，最重要的是可以看到每个时间点的账户总资产。

**Q: WalletBalance 和 MarginBalance 有什么区别？**
A: WalletBalance 是钱包总余额，MarginBalance 是可用于保证金的余额。

**Q: 为什么有些记录的金额是负数？**
A: 负数表示资金减少，如提款、手续费支出、交易亏损等。

**Q: 如何计算总盈亏？**
A: 可以筛选 TransactType 为 "RealisedPNL" 的记录，对 Amount 求和即可。

**Q: CSV可以在Excel中打开吗？**
A: 可以，CSV格式兼容所有主流表格软件（Excel、Google Sheets等）。

**Q: 如何绘制账户资产变化曲线？**
A: 在Excel中，使用 Timestamp 作为X轴，WalletBalance_BTC 作为Y轴创建折线图。

**Q: 增量更新会修改原文件吗？**
A: 是的，新记录会追加到文件末尾，不会覆盖原有数据。

**Q: 数据是实时的吗？**
A: 数据是API返回的历史记录，通常有几秒钟延迟。

## 相关文件

- `bitmex_wallet.go` - 钱包历史记录下载工具源代码
- `bitmex_data.go` - 交易记录下载工具源代码
- `Makefile` - 便捷命令集合
- `README_bitmex.md` - 交易记录工具说明文档

## 单位换算参考

| Satoshi | BTC |
|---------|-----|
| 1 | 0.00000001 |
| 100 | 0.000001 |
| 10,000 | 0.0001 |
| 1,000,000 | 0.01 |
| 100,000,000 | 1.0 |
