package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// Execution 成交记录
type Execution struct {
	Timestamp time.Time
	Symbol    string
	Side      string
	Qty       int
	Price     float64
}

// Kline K线数据
type Kline struct {
	Timestamp time.Time
	Close     float64
}

// WalletRecord 钱包记录
type WalletRecord struct {
	Timestamp time.Time
	Balance   float64
}

// DailyPosition 每日仓位数据
type DailyPosition struct {
	Date          string
	PositionQty   int
	Price         float64
	Balance       float64
	PositionValue float64
	PositionRatio float64
	Side          string
}

func main() {
	log.Println("📊 开始计算每日 BTC 仓位比例...")

	// 1. 加载数据
	executions := loadExecutions("executions.csv")
	klines := loadKlines("klines_XBTUSD_1d.csv")
	walletRecords := loadWalletRecords("wallet.csv")

	log.Printf("✓ 加载 %d 条成交记录", len(executions))
	log.Printf("✓ 加载 %d 条 K线数据", len(klines))
	log.Printf("✓ 加载 %d 条钱包记录", len(walletRecords))

	// 2. 计算每日仓位
	dailyPositions := calculateDailyPositions(executions, klines, walletRecords)

	// 3. 保存为 CSV
	outputFile := "daily_position.csv"
	if err := saveDailyPositions(dailyPositions, outputFile); err != nil {
		log.Fatalf("❌ 保存失败: %v", err)
	}

	log.Printf("✅ 成功保存 %d 天的仓位数据到 %s", len(dailyPositions), outputFile)
	log.Println("📈 统计信息:")
	printSummary(dailyPositions)
}

// loadExecutions 加载成交记录
func loadExecutions(filename string) []Execution {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("无法打开 %s: %v", filename, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("读取 CSV 失败: %v", err)
	}

	var executions []Execution
	for i, record := range records {
		if i == 0 {
			continue // 跳过表头
		}

		timestamp, _ := time.Parse(time.RFC3339, record[14])
		symbol := record[4]
		side := record[5]
		qty, _ := strconv.Atoi(record[6])
		price, _ := strconv.ParseFloat(record[7], 64)
		execType := record[17] // ExecType 字段

		// 只保留 XBTUSD
		if !strings.Contains(symbol, "XBT") && symbol != "XBTUSD" {
			continue
		}

		// 跳过 Funding (资金费率结算)，只保留真实交易
		if execType == "Funding" {
			continue
		}

		executions = append(executions, Execution{
			Timestamp: timestamp,
			Symbol:    symbol,
			Side:      side,
			Qty:       qty,
			Price:     price,
		})
	}

	return executions
}

// loadKlines 加载 K线数据
func loadKlines(filename string) map[string]Kline {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("无法打开 %s: %v", filename, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("读取 CSV 失败: %v", err)
	}

	klines := make(map[string]Kline)
	for i, record := range records {
		if i == 0 {
			continue
		}

		timestamp, _ := time.Parse(time.RFC3339, record[0])
		close, _ := strconv.ParseFloat(record[5], 64)

		// 使用日期作为 key
		dateKey := timestamp.Format("2006-01-02")
		klines[dateKey] = Kline{
			Timestamp: timestamp,
			Close:     close,
		}
	}

	return klines
}

// loadWalletRecords 加载钱包记录
func loadWalletRecords(filename string) []WalletRecord {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("无法打开 %s: %v", filename, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("读取 CSV 失败: %v", err)
	}

	var walletRecords []WalletRecord
	for i, record := range records {
		if i == 0 {
			continue
		}

		timestamp, _ := time.Parse(time.RFC3339, record[9])
		balance, _ := strconv.ParseFloat(record[11], 64)

		walletRecords = append(walletRecords, WalletRecord{
			Timestamp: timestamp,
			Balance:   balance,
		})
	}

	return walletRecords
}

// calculateDailyPositions 计算每日仓位
func calculateDailyPositions(executions []Execution, klines map[string]Kline, walletRecords []WalletRecord) []DailyPosition {
	// 确定日期范围
	startDate := time.Date(2020, 5, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Now()

	var dailyPositions []DailyPosition

	// 按日期遍历
	for date := startDate; date.Before(endDate) || date.Equal(endDate); date = date.AddDate(0, 0, 1) {
		dateStr := date.Format("2006-01-02")

		// 1. 计算截至该日期的累计持仓
		positionQty := 0
		for _, exec := range executions {
			if exec.Timestamp.After(date) {
				break
			}

			if exec.Side == "Buy" {
				positionQty += exec.Qty
			} else {
				positionQty -= exec.Qty
			}
		}

		// 2. 获取该日收盘价
		kline, exists := klines[dateStr]
		if !exists {
			continue // 该日无 K线数据，跳过
		}
		price := kline.Close

		// 3. 获取该日余额
		balance := getBalanceAtDate(walletRecords, date)
		if balance == 0 {
			continue
		}

		// 4. 计算仓位价值和比例
		var positionValue float64
		var positionRatio float64
		var side string

		if positionQty == 0 {
			side = "Flat"
			positionValue = 0
			positionRatio = 0
		} else {
			// 反向合约: 仓位价值 = abs(数量) / 价格
			positionValue = math.Abs(float64(positionQty)) / price

			// 仓位比例 = 仓位价值 / 余额
			positionRatio = positionValue / balance

			// 确定方向
			if positionQty > 0 {
				side = "Long"
			} else {
				side = "Short"
				positionRatio = -positionRatio // Short 为负值
			}
		}

		dailyPositions = append(dailyPositions, DailyPosition{
			Date:          dateStr,
			PositionQty:   positionQty,
			Price:         price,
			Balance:       balance,
			PositionValue: positionValue,
			PositionRatio: positionRatio,
			Side:          side,
		})
	}

	return dailyPositions
}

// getBalanceAtDate 获取指定日期的余额
func getBalanceAtDate(walletRecords []WalletRecord, targetDate time.Time) float64 {
	var balance float64

	for _, record := range walletRecords {
		if record.Timestamp.After(targetDate) {
			break
		}
		balance = record.Balance
	}

	return balance
}

// saveDailyPositions 保存到 CSV
func saveDailyPositions(positions []DailyPosition, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	header := []string{"Date", "PositionQty", "Price", "Balance", "PositionValue", "PositionRatio", "Side"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// 写入数据
	for _, pos := range positions {
		record := []string{
			pos.Date,
			fmt.Sprintf("%d", pos.PositionQty),
			fmt.Sprintf("%.2f", pos.Price),
			fmt.Sprintf("%.8f", pos.Balance),
			fmt.Sprintf("%.8f", pos.PositionValue),
			fmt.Sprintf("%.4f", pos.PositionRatio),
			pos.Side,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// printSummary 打印统计信息
func printSummary(positions []DailyPosition) {
	if len(positions) == 0 {
		return
	}

	longCount := 0
	shortCount := 0
	flatCount := 0
	maxLongRatio := 0.0
	maxShortRatio := 0.0

	for _, pos := range positions {
		switch pos.Side {
		case "Long":
			longCount++
			if pos.PositionRatio > maxLongRatio {
				maxLongRatio = pos.PositionRatio
			}
		case "Short":
			shortCount++
			if math.Abs(pos.PositionRatio) > maxShortRatio {
				maxShortRatio = math.Abs(pos.PositionRatio)
			}
		case "Flat":
			flatCount++
		}
	}

	total := len(positions)
	fmt.Printf("  总天数: %d\n", total)
	fmt.Printf("  Long 天数: %d (%.1f%%)\n", longCount, float64(longCount)/float64(total)*100)
	fmt.Printf("  Short 天数: %d (%.1f%%)\n", shortCount, float64(shortCount)/float64(total)*100)
	fmt.Printf("  空仓天数: %d (%.1f%%)\n", flatCount, float64(flatCount)/float64(total)*100)
	fmt.Printf("  最大 Long 倍数: %.2fx\n", maxLongRatio)
	fmt.Printf("  最大 Short 倍数: %.2fx\n", maxShortRatio)

	// 显示最近5天
	fmt.Println("\n  最近5天:")
	start := len(positions) - 5
	if start < 0 {
		start = 0
	}
	for i := start; i < len(positions); i++ {
		pos := positions[i]
		fmt.Printf("    %s: %s %.2fx\n", pos.Date, pos.Side, math.Abs(pos.PositionRatio))
	}
}
