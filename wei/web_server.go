package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

// KlineData K线数据
type KlineData struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

// OrderData 订单数据
type OrderData struct {
	OrderID       string  `json:"orderId"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Price         float64 `json:"price"`
	Qty           int     `json:"qty"`
	OrderType     string  `json:"orderType"`
	Status        string  `json:"status"`
	Timestamp     string  `json:"timestamp"`
	TimestampUnix int64   `json:"timestampUnix"`
}

// ExecutionData 成交数据
type ExecutionData struct {
	ExecID        string  `json:"execId"`
	OrderID       string  `json:"orderId"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Price         float64 `json:"price"`
	Qty           int     `json:"qty"`
	Commission    float64 `json:"commission"`
	Timestamp     string  `json:"timestamp"`
	TimestampUnix int64   `json:"timestampUnix"`
}

// Position 仓位数据
type Position struct {
	Symbol         string  `json:"symbol"`
	Side           string  `json:"side"`
	Qty            int     `json:"qty"`
	EntryPrice     float64 `json:"entryPrice"`
	CurrentPrice   float64 `json:"currentPrice"`
	UnrealizedPNL  float64 `json:"unrealizedPnl"`
	UnrealizedPNLPercent float64 `json:"unrealizedPnlPercent"`
}

// AccountInfo 账户信息
type AccountInfo struct {
	Balance        float64 `json:"balance"`
	TodayPNL       float64 `json:"todayPnl"`
	TodayPNLPercent float64 `json:"todayPnlPercent"`
	TotalPNL       float64 `json:"totalPnl"`
	WinRate        float64 `json:"winRate"`
	TotalTrades    int     `json:"totalTrades"`
}

// 全局数据缓存
var (
	klinesCache     map[string][]KlineData
	ordersCache     []OrderData
	executionsCache []ExecutionData
)

func main() {
	// 加载数据
	log.Println("正在加载数据...")
	loadData()

	// 设置路由
	http.HandleFunc("/api/klines", handleKlines)
	http.HandleFunc("/api/orders", handleOrders)
	http.HandleFunc("/api/orders/pending", handlePendingOrders)
	http.HandleFunc("/api/executions", handleExecutions)
	http.HandleFunc("/api/positions", handlePositions)
	http.HandleFunc("/api/account", handleAccount)

	// 静态文件服务
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	// 启动服务器
	port := "8080"
	log.Printf("🌐 Web服务器启动成功!")
	log.Printf("   访问: http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, enableCORS(http.DefaultServeMux)))
}

// enableCORS 启用CORS
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loadData 加载所有CSV数据
func loadData() {
	klinesCache = make(map[string][]KlineData)

	// 加载K线数据
	symbols := []string{"XBTUSD", "ETHUSD"}
	timeframes := []string{"1d", "1h"}

	for _, symbol := range symbols {
		for _, tf := range timeframes {
			filename := fmt.Sprintf("klines_%s_%s.csv", symbol, tf)
			if klines, err := loadKlines(filename); err == nil {
				key := fmt.Sprintf("%s_%s", symbol, tf)
				klinesCache[key] = klines
				log.Printf("✓ 加载 %s: %d 条记录", filename, len(klines))
			}
		}
	}

	// 加载订单数据
	if orders, err := loadOrders("orders.csv"); err == nil {
		ordersCache = orders
		log.Printf("✓ 加载 orders.csv: %d 条记录", len(orders))
	}

	// 加载成交数据
	if execs, err := loadExecutions("executions.csv"); err == nil {
		executionsCache = execs
		log.Printf("✓ 加载 executions.csv: %d 条记录", len(execs))
	}
}

// loadKlines 加载K线CSV文件
func loadKlines(filename string) ([]KlineData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var klines []KlineData
	for i, record := range records {
		if i == 0 {
			continue // 跳过表头
		}

		timestamp, _ := time.Parse(time.RFC3339, record[0])
		open, _ := strconv.ParseFloat(record[2], 64)
		high, _ := strconv.ParseFloat(record[3], 64)
		low, _ := strconv.ParseFloat(record[4], 64)
		close, _ := strconv.ParseFloat(record[5], 64)
		volume, _ := strconv.ParseInt(record[6], 10, 64)

		klines = append(klines, KlineData{
			Time:   timestamp.Unix(),
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	return klines, nil
}

// loadOrders 加载订单CSV文件
func loadOrders(filename string) ([]OrderData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var orders []OrderData
	for i, record := range records {
		if i == 0 {
			continue
		}

		timestamp, _ := time.Parse(time.RFC3339, record[31])
		price, _ := strconv.ParseFloat(record[8], 64)
		qty, _ := strconv.Atoi(record[7])

		orders = append(orders, OrderData{
			OrderID:       record[0],
			Symbol:        record[4],
			Side:          record[5],
			Price:         price,
			Qty:           qty,
			OrderType:     record[15],
			Status:        record[20],
			Timestamp:     record[31],
			TimestampUnix: timestamp.Unix(),
		})
	}

	return orders, nil
}

// loadExecutions 加载成交CSV文件
func loadExecutions(filename string) ([]ExecutionData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var executions []ExecutionData
	for i, record := range records {
		if i == 0 {
			continue
		}

		timestamp, _ := time.Parse(time.RFC3339, record[14])
		price, _ := strconv.ParseFloat(record[7], 64)
		qty, _ := strconv.Atoi(record[6])
		commission, _ := strconv.ParseFloat(record[13], 64)

		executions = append(executions, ExecutionData{
			ExecID:        record[0],
			OrderID:       record[1],
			Symbol:        record[4],
			Side:          record[5],
			Price:         price,
			Qty:           qty,
			Commission:    commission,
			Timestamp:     record[14],
			TimestampUnix: timestamp.Unix(),
		})
	}

	return executions, nil
}

// API Handlers

func handleKlines(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	timeframe := r.URL.Query().Get("timeframe")

	if symbol == "" {
		symbol = "XBTUSD"
	}
	if timeframe == "" {
		timeframe = "1d"
	}

	key := fmt.Sprintf("%s_%s", symbol, timeframe)
	klines, exists := klinesCache[key]

	if !exists {
		http.Error(w, "数据不存在", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(klines)
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	orders := ordersCache
	if status != "" {
		var filtered []OrderData
		for _, order := range orders {
			if order.Status == status {
				filtered = append(filtered, order)
			}
		}
		orders = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func handlePendingOrders(w http.ResponseWriter, r *http.Request) {
	var pending []OrderData
	for _, order := range ordersCache {
		if order.Status == "New" || order.Status == "PartiallyFilled" {
			pending = append(pending, order)
		}
	}

	// 按时间倒序
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].TimestampUnix > pending[j].TimestampUnix
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pending)
}

func handleExecutions(w http.ResponseWriter, r *http.Request) {
	executions := executionsCache

	// 按时间倒序
	sort.Slice(executions, func(i, j int) bool {
		return executions[i].TimestampUnix > executions[j].TimestampUnix
	})

	// 限制返回数量
	if len(executions) > 100 {
		executions = executions[:100]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(executions)
}

func handlePositions(w http.ResponseWriter, r *http.Request) {
	positions := calculatePositions()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(positions)
}

func handleAccount(w http.ResponseWriter, r *http.Request) {
	account := calculateAccountInfo()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

// calculatePositions 计算当前仓位
func calculatePositions() []Position {
	positionsMap := make(map[string]*Position)

	for _, exec := range executionsCache {
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

		// 简化：使用最后成交价作为当前价
		pos.CurrentPrice = exec.Price
	}

	// 计算入场均价
	for symbol, pos := range positionsMap {
		if pos.Qty == 0 {
			delete(positionsMap, symbol)
			continue
		}

		var totalCost float64
		var totalQty int

		for _, exec := range executionsCache {
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

		if totalQty != 0 {
			pos.EntryPrice = float64(totalQty) / totalCost
		}

		// 设置方向
		if pos.Qty > 0 {
			pos.Side = "Long"
		} else {
			pos.Side = "Short"
		}

		// 计算未实现盈亏
		if pos.EntryPrice > 0 {
			pos.UnrealizedPNL = (1.0/pos.EntryPrice - 1.0/pos.CurrentPrice) * float64(pos.Qty)
			pos.UnrealizedPNLPercent = (pos.CurrentPrice/pos.EntryPrice - 1.0) * 100
			if pos.Side == "Short" {
				pos.UnrealizedPNL = -pos.UnrealizedPNL
				pos.UnrealizedPNLPercent = -pos.UnrealizedPNLPercent
			}
		}
	}

	var positions []Position
	for _, pos := range positionsMap {
		positions = append(positions, *pos)
	}

	return positions
}

// calculateAccountInfo 计算账户信息
func calculateAccountInfo() AccountInfo {
	// 读取wallet.csv获取余额信息
	file, err := os.Open("wallet.csv")
	if err != nil {
		return AccountInfo{}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return AccountInfo{}
	}

	if len(records) < 2 {
		return AccountInfo{}
	}

	// 获取最新余额
	lastRecord := records[len(records)-1]
	balance, _ := strconv.ParseFloat(lastRecord[11], 64)

	// 计算总盈亏
	var totalPNL float64
	var winCount, totalCount int

	for i := 1; i < len(records); i++ {
		if records[i][1] == "RealisedPNL" {
			pnl, _ := strconv.ParseFloat(records[i][6], 64)
			totalPNL += pnl
			totalCount++
			if pnl > 0 {
				winCount++
			}
		}
	}

	var winRate float64
	if totalCount > 0 {
		winRate = float64(winCount) / float64(totalCount) * 100
	}

	// 计算今日盈亏 (简化：使用最后10笔)
	var todayPNL float64
	startIdx := len(records) - 10
	if startIdx < 1 {
		startIdx = 1
	}

	for i := startIdx; i < len(records); i++ {
		if records[i][1] == "RealisedPNL" {
			pnl, _ := strconv.ParseFloat(records[i][6], 64)
			todayPNL += pnl
		}
	}

	todayPNLPercent := (todayPNL / balance) * 100

	return AccountInfo{
		Balance:         balance,
		TodayPNL:        todayPNL,
		TodayPNLPercent: todayPNLPercent,
		TotalPNL:        totalPNL,
		WinRate:         winRate,
		TotalTrades:     totalCount,
	}
}
