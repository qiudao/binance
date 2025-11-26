#!/bin/bash

# 查看每日仓位数据的脚本

if [ ! -f daily_position.csv ]; then
    echo "❌ 找不到 daily_position.csv，请先运行: make daily-position"
    exit 1
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  📊 每日 BTC 仓位比例分析"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 统计信息
awk -F',' '
NR==1 {next}
{
    total++
    if ($7 == "Long") {
        long_count++
        long_sum += $6
    } else if ($7 == "Short") {
        short_count++
        short_sum += -$6
    } else {
        flat_count++
    }
}
END {
    print "📈 总体统计:"
    print "  总天数:", total
    print "  Long 天数:", long_count, sprintf("(%.1f%%)", long_count/total*100)
    print "  Short 天数:", short_count, sprintf("(%.1f%%)", short_count/total*100)
    print "  空仓天数:", flat_count, sprintf("(%.1f%%)", flat_count/total*100)
    print ""
    print "  平均 Long 倍数:", sprintf("%.2fx", long_sum/long_count)
    print "  平均 Short 倍数:", sprintf("%.2fx", short_sum/short_count)
}
' daily_position.csv

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  📅 最近 30 天仓位变化"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

awk -F',' '
NR==1 {next}
{
    date=$1
    ratio=$6
    side=$7
    
    if (side == "Long") {
        printf "  %s  %6.2fx  🟢 %s\n", date, ratio, side
    } else if (side == "Short") {
        printf "  %s  %6.2fx  🔴 %s\n", date, -ratio, side
    } else {
        printf "  %s  %6.2fx  ⚪ %s\n", date, ratio, side
    }
}
' daily_position.csv | tail -30

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
