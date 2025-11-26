BEGIN {FS=","}
NR>1 && $5==symbol {
    price = $8 + 0  # 成交价格
    qty = $7 + 0    # 成交数量
    
    if ($6 == "Buy") {
        totalQty += qty
        totalCost += qty / price
    } else {
        totalQty -= qty
        totalCost -= qty / price
    }
    
    lastPrice = price
}
END {
    if (totalQty != 0 && totalCost != 0) {
        entryPrice = totalQty / totalCost
        printf "持仓数量: %d\n", totalQty
        printf "入场均价: $%.2f\n", entryPrice
        printf "最后价格: $%.2f\n", lastPrice
        
        # 计算盈亏
        absQty = (totalQty < 0) ? -totalQty : totalQty
        pnl = (1.0/entryPrice - 1.0/lastPrice) * absQty
        pnlPercent = (lastPrice/entryPrice - 1.0) * 100
        
        # Short仓位反向
        if (totalQty < 0) {
            pnl = -pnl
            pnlPercent = -pnlPercent
        }
        
        printf "未实现盈亏: %.8f BTC (%.2f%%)\n", pnl, pnlPercent
    }
}
