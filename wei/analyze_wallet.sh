#!/bin/bash

# BitMEX 钱包资金分析报告
# 分析 wallet.csv 生成详细的资金变化报告

CSV_FILE="wallet.csv"

if [ ! -f "$CSV_FILE" ]; then
    echo "❌ 错误: 找不到 $CSV_FILE 文件"
    echo "请先运行: make download-wallet"
    exit 1
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  📊 BitMEX 钱包资金分析报告"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 1. 总览统计
echo "📈 总览统计"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
awk -F',' '
NR==2 {
    start_balance=$12
    start_time=$10
    start_date=substr($10,1,10)
}
NR>1 {
    end_balance=$12
    end_time=$10
    end_date=substr($10,1,10)
    records++
}
END {
    growth = end_balance - start_balance
    growth_pct = (growth / start_balance) * 100
    days = (systime() - mktime(gensub(/-|T|:|Z/, " ", "g", start_date " 00:00:00"))) / 86400

    printf "  时间范围: %s 至 %s\n", start_date, end_date
    printf "  交易天数: %.0f 天\n", days
    printf "  记录总数: %d 条\n", records
    printf "\n"
    printf "  起始余额: %.8f BTC\n", start_balance
    printf "  最终余额: %.8f BTC\n", end_balance
    printf "  资产增长: %.8f BTC (%.2f%%)\n", growth, growth_pct
}
' "$CSV_FILE"
echo ""

# 2. 资金流动分析
echo "💰 资金流动分析"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
awk -F',' '
NR>1 && $2=="Deposit" && $3=="Completed" {
    deposit_sum += $7
    deposit_count++
    deposits[deposit_count] = $10 " | " $7 " BTC"
}
NR>1 && $2=="Withdrawal" && $3=="Completed" {
    withdrawal_sum += $7
    withdrawal_count++
    withdrawals[withdrawal_count] = $10 " | " $7 " BTC"
}
END {
    net_flow = deposit_sum + withdrawal_sum

    printf "  充币记录: %d 笔\n", deposit_count
    for (i=1; i<=deposit_count; i++) {
        printf "    • %s\n", deposits[i]
    }
    printf "  充币总额: %.8f BTC\n\n", deposit_sum

    printf "  提币记录: %d 笔\n", withdrawal_count
    for (i=1; i<=withdrawal_count; i++) {
        printf "    • %s\n", withdrawals[i]
    }
    printf "  提币总额: %.8f BTC\n\n", withdrawal_sum

    printf "  净充值额: %.8f BTC\n", net_flow
    if (net_flow < 0) {
        printf "  (净提出 %.8f BTC)\n", -net_flow
    }
}
' "$CSV_FILE"
echo ""

# 3. 交易盈亏分析
echo "📊 交易盈亏分析"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
awk -F',' '
NR>1 && $2=="RealisedPNL" {
    pnl_sum += $7
    pnl_count++
    if ($7 > 0) {
        profit_sum += $7
        profit_count++
    } else {
        loss_sum += $7
        loss_count++
    }
}
NR>1 && $2=="Funding" {
    funding_sum += $7
    funding_count++
}
NR>1 && $2=="Conversion" {
    conversion_count++
}
NR>1 && $2=="SpotTrade" {
    spot_count++
}
END {
    win_rate = (profit_count / pnl_count) * 100

    printf "  已实现盈亏 (RealisedPNL):\n"
    printf "    交易次数: %d 笔\n", pnl_count
    printf "    盈利次数: %d 笔 (总计: %.8f BTC)\n", profit_count, profit_sum
    printf "    亏损次数: %d 笔 (总计: %.8f BTC)\n", loss_count, loss_sum
    printf "    胜率: %.2f%%\n", win_rate
    printf "    净盈亏: %.8f BTC\n\n", pnl_sum

    printf "  资金费 (Funding):\n"
    printf "    收付次数: %d 笔\n", funding_count
    printf "    净收入: %.8f BTC\n\n", funding_sum

    printf "  其他交易:\n"
    printf "    币种转换: %d 笔\n", conversion_count
    printf "    现货交易: %d 笔\n", spot_count
}
' "$CSV_FILE"
echo ""

# 4. 按年度统计
echo "📅 年度统计"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
awk -F',' '
NR>1 {
    year = substr($10, 1, 4)
    year_balance[year] = $12

    if ($2 == "RealisedPNL") {
        year_pnl[year] += $7
    }
    if ($2 == "Funding") {
        year_funding[year] += $7
    }
}
END {
    n = asorti(year_balance, sorted_years)
    printf "  年份     年末余额          已实现盈亏      资金费收入\n"
    printf "  ──────────────────────────────────────────────────────\n"
    for (i=1; i<=n; i++) {
        y = sorted_years[i]
        printf "  %s  %12.8f BTC  %12.8f BTC  %12.8f BTC\n",
            y, year_balance[y], year_pnl[y], year_funding[y]
    }
}
' "$CSV_FILE"
echo ""

# 5. 关键事件时间线
echo "🎯 关键事件时间线"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
awk -F',' '
NR==2 {
    min_balance = $12
    max_balance = $12
    start_balance = $12
}
NR>1 {
    if ($12 > max_balance) {
        max_balance = $12
        max_time = substr($10, 1, 10)
    }
    if ($12 < min_balance && NR > 100) {  # 排除初始阶段
        min_balance = $12
        min_time = substr($10, 1, 10)
    }

    # 记录大额盈亏 (单笔 > 0.1 BTC)
    if ($2 == "RealisedPNL" && ($7 > 0.1 || $7 < -0.1)) {
        big_pnl[++big_count] = sprintf("%s | %+.8f BTC | 余额: %.8f BTC",
            substr($10, 1, 10), $7, $12)
    }

    # 记录充提币
    if ($2 == "Deposit" || $2 == "Withdrawal") {
        deposits[++dep_count] = sprintf("%s | %s %+.8f BTC",
            substr($10, 1, 10), $2, $7)
    }

    end_balance = $12
}
END {
    drawdown = ((max_balance - min_balance) / max_balance) * 100
    recovery = ((end_balance - min_balance) / (max_balance - min_balance)) * 100

    printf "  📈 峰值时刻: %s  余额: %.8f BTC\n", max_time, max_balance
    printf "  📉 谷底时刻: %s  余额: %.8f BTC\n", min_time, min_balance
    printf "  📊 最大回撤: %.2f%%\n", drawdown
    printf "  💪 回撤恢复: %.2f%%\n", recovery
    printf "\n"

    printf "  大额盈亏事件 (单笔 >0.1 BTC):\n"
    for (i=1; i<=5 && i<=big_count; i++) {
        printf "    • %s\n", big_pnl[i]
    }
    if (big_count > 5) printf "    ... 共 %d 笔\n", big_count
}
' "$CSV_FILE"
echo ""

# 6. ROI 计算
echo "💎 投资回报率 (ROI) 分析"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
awk -F',' '
NR==2 {
    start_balance = $12
}
NR>1 {
    end_balance = $12
}
NR>1 && $2=="Deposit" && $3=="Completed" {
    deposit_sum += $7
}
NR>1 && $2=="Withdrawal" && $3=="Completed" {
    withdrawal_sum += $7
}
NR>1 && $2=="RealisedPNL" {
    pnl_sum += $7
}
END {
    net_deposit = deposit_sum + withdrawal_sum
    profit = end_balance - start_balance - net_deposit

    # 简单ROI (不考虑时间)
    simple_roi = (profit / start_balance) * 100

    # 考虑充提币的ROI
    if (net_deposit > 0) {
        adjusted_roi = (profit / (start_balance + net_deposit)) * 100
    } else {
        # 如果是净提出,用初始+累计盈亏作为基数
        adjusted_roi = (end_balance / start_balance - 1) * 100
    }

    printf "  初始投入: %.8f BTC\n", start_balance
    printf "  净充提额: %+.8f BTC\n", net_deposit
    printf "  最终余额: %.8f BTC\n", end_balance
    printf "\n"
    printf "  交易盈亏: %.8f BTC\n", pnl_sum
    printf "  净利润: %.8f BTC\n", profit
    printf "\n"
    printf "  简单 ROI: %.2f%%\n", simple_roi
    printf "  调整 ROI: %.2f%% (考虑充提币)\n", adjusted_roi
}
' "$CSV_FILE"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  报告生成完成"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
