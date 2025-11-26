package main

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// handleSnapshot 处理历史快照API请求
func handleSnapshot(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	snapshot := generateSnapshot(targetDate)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

// generateSnapshot 生成指定日期的快照数据
func generateSnapshot(targetDate time.Time) DailySnapshot {
	// 设置日期范围
	minDate := "2020-05-01"
	maxDate := time.Now().Format("2006-01-02")

	// 1. 获取截至该日期的余额
	balance := getBalanceAtDate(targetDate)

	// 2. 计算截至该日期的持仓
	positions := getPositionsAtDate(targetDate)

	// 3. 只保留BTC相关持仓
	btcPositions := []Position{}
	for _, pos := range positions {
		if strings.Contains(pos.Symbol, "XBT") || pos.Symbol == "XBTUSD" {
			btcPositions = append(btcPositions, pos)
		}
	}

	// 4. 计算未实现盈亏
	var unrealizedPNL float64
	for _, pos := range btcPositions {
		unrealizedPNL += pos.UnrealizedPNL
	}

	// 5. 获取当日订单
	todayOrders := getOrdersAtDate(targetDate)

	// 6. 获取最近50笔成交
	recentExecs := getRecentExecutions(targetDate, 50)

	// 7. 获取K线数据(该日期前90天)
	klineData := getKlinesBeforeDate(targetDate, 90)

	return DailySnapshot{
		Date:          targetDate.Format("2006-01-02"),
		Balance:       balance,
		TotalEquity:   balance + unrealizedPNL,
		UnrealizedPNL: unrealizedPNL,
		BTCPositions:  btcPositions,
		TodayOrders:   todayOrders,
		RecentExecs:   recentExecs,
		KlineData:     klineData,
		MinDate:       minDate,
		MaxDate:       maxDate,
	}
}

// getBalanceAtDate 获取截至指定日期的余额
func getBalanceAtDate(targetDate time.Time) float64 {
	file, err := os.Open("wallet.csv")
	if err != nil {
		log.Printf("Error opening wallet.csv: %v", err)
		return 0
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return 0
	}

	var balance float64
	for i := 1; i < len(records); i++ {
		timestamp, err := time.Parse(time.RFC3339, records[i][9])
		if err != nil {
			continue
		}

		// 只取该日期之前的记录
		if timestamp.After(targetDate) {
			break
		}

		balance, _ = strconv.ParseFloat(records[i][11], 64)
	}

	return balance
}

// getPositionsAtDate 计算截至指定日期的持仓
func getPositionsAtDate(targetDate time.Time) []Position {
	positions := []Position{}
	positionsMap := make(map[string]*Position)

	// 筛选截至目标日期的成交记录
	filteredExecs := []ExecutionData{}
	for _, exec := range executionsCache {
		execTime, err := time.Parse(time.RFC3339, exec.Timestamp)
		if err != nil {
			continue
		}

		if execTime.Before(targetDate) || execTime.Equal(targetDate) {
			filteredExecs = append(filteredExecs, exec)
		}
	}

	// 计算持仓数量
	for _, exec := range filteredExecs {
		pos, exists := positionsMap[exec.Symbol]
		if !exists {
			pos = &Position{Symbol: exec.Symbol}
			positionsMap[exec.Symbol] = pos
		}

		if exec.Side == "Buy" {
			pos.Qty += exec.Qty
		} else {
			pos.Qty -= exec.Qty
		}
	}

	// 计算入场均价和未实现盈亏
	for symbol, pos := range positionsMap {
		if pos.Qty == 0 {
			delete(positionsMap, symbol)
			continue
		}

		// 计算入场均价
		var totalCost float64
		var totalQty int

		for _, exec := range filteredExecs {
			if exec.Symbol != symbol {
				continue
			}

			if exec.Side == "Buy" {
				totalQty += exec.Qty
				totalCost += float64(exec.Qty) / exec.Price
			} else {
				totalQty -= exec.Qty
				totalCost -= float64(exec.Qty) / exec.Price
			}
		}

		if totalQty != 0 && totalCost != 0 {
			entryPrice := float64(totalQty) / totalCost
			if !math.IsNaN(entryPrice) && !math.IsInf(entryPrice, 0) && entryPrice > 0 {
				pos.EntryPrice = entryPrice
			}
		}

		// 获取该日期的收盘价作为当前价
		pos.CurrentPrice = getClosePriceAtDate(symbol, targetDate)

		// 设置方向
		if pos.Qty > 0 {
			pos.Side = "Long"
		} else {
			pos.Side = "Short"
		}

		// 计算未实现盈亏
		if pos.EntryPrice > 0 && pos.CurrentPrice > 0 {
			qty := float64(pos.Qty)
			pnl := qty * (1.0/pos.EntryPrice - 1.0/pos.CurrentPrice)

			pnlPercent := (pos.CurrentPrice/pos.EntryPrice - 1.0) * 100
			if pos.Side == "Short" {
				pnlPercent = -pnlPercent
			}

			if !math.IsNaN(pnl) && !math.IsInf(pnl, 0) {
				pos.UnrealizedPNL = pnl
			}
			if !math.IsNaN(pnlPercent) && !math.IsInf(pnlPercent, 0) {
				pos.UnrealizedPNLPercent = pnlPercent
			}
		}
	}

	for _, pos := range positionsMap {
		positions = append(positions, *pos)
	}

	return positions
}

// getClosePriceAtDate 获取指定日期的K线收盘价
func getClosePriceAtDate(symbol string, targetDate time.Time) float64 {
	key := symbol + "_1d"
	klines, exists := klinesCache[key]
	if !exists {
		return 0
	}

	targetTimestamp := targetDate.Unix()
	var closePrice float64

	// 查找该日期或之前的最近一条K线
	for i := len(klines) - 1; i >= 0; i-- {
		if klines[i].Time <= targetTimestamp {
			closePrice = klines[i].Close
			break
		}
	}

	return closePrice
}

// getOrdersAtDate 获取指定日期的订单
func getOrdersAtDate(targetDate time.Time) []OrderData {
	var orders []OrderData

	for _, order := range ordersCache {
		orderTime, err := time.Parse(time.RFC3339, order.Timestamp)
		if err != nil {
			continue
		}

		// 判断是否在同一天
		if orderTime.Year() == targetDate.Year() &&
			orderTime.Month() == targetDate.Month() &&
			orderTime.Day() == targetDate.Day() {
			orders = append(orders, order)
		}
	}

	// 按时间倒序
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].TimestampUnix > orders[j].TimestampUnix
	})

	return orders
}

// getRecentExecutions 获取截至指定日期的最近N笔成交
func getRecentExecutions(targetDate time.Time, limit int) []ExecutionData {
	var executions []ExecutionData

	for _, exec := range executionsCache {
		execTime, err := time.Parse(time.RFC3339, exec.Timestamp)
		if err != nil {
			continue
		}

		if execTime.Before(targetDate) || execTime.Equal(targetDate) {
			executions = append(executions, exec)
		}
	}

	// 按时间倒序
	sort.Slice(executions, func(i, j int) bool {
		return executions[i].TimestampUnix > executions[j].TimestampUnix
	})

	// 限制返回数量
	if len(executions) > limit {
		executions = executions[:limit]
	}

	return executions
}

// getKlinesBeforeDate 获取指定日期前N天的K线数据
func getKlinesBeforeDate(targetDate time.Time, days int) []KlineData {
	key := "XBTUSD_1d"
	klines, exists := klinesCache[key]
	if !exists {
		return []KlineData{}
	}

	targetTimestamp := targetDate.Unix()
	startTimestamp := targetDate.AddDate(0, 0, -days).Unix()

	var result []KlineData
	for _, kline := range klines {
		if kline.Time >= startTimestamp && kline.Time <= targetTimestamp {
			result = append(result, kline)
		}
	}

	return result
}
