# 报告目录

此目录用于存放所有生成的分析报告文件。

## 生成的文件

### 📄 文本报告
- `wallet_report.txt` - 钱包资金分析文本报告（使用 `make analyze` 生成）

### 📊 可视化图表 (PNG)
使用 `make plot` 生成以下图表：
- `balance_trend.png` - 账户余额变化趋势图
- `pnl_cumulative.png` - 累计盈亏趋势图
- `monthly_pnl.png` - 月度盈亏柱状图
- `transaction_types.png` - 交易类型占比饼图
- `drawdown_analysis.png` - 回撤分析图

### 🌐 交互式仪表板
- `wallet_dashboard.html` - HTML交互式仪表板（使用 `make dashboard` 生成）

## 使用方法

```bash
# 生成文本报告
make analyze

# 生成图表（需要 pandas + matplotlib）
make plot

# 生成HTML仪表板（需要 pandas）
make dashboard
```

## 注意

- 所有报告文件都是自动生成的，可以随时重新生成
- 报告基于 `wallet.csv` 数据文件
- 建议将此目录添加到 `.gitignore` 以避免提交生成的报告文件
