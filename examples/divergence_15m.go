package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"binance-kline/indicators"
)

// 北京时间时区
var BeijingLocation = time.FixedZone("CST", 8*3600)

func main() {
	// 读取15分钟CSV文件
	csvFile := "data/klines_15m.csv"
	klines, err := loadKlinesFromCSV(csvFile)
	if err != nil {
		fmt.Printf("读取CSV文件失败: %v\n", err)
		fmt.Println("请先运行: make save-15m")
		return
	}

	fmt.Printf("\n============ 15分钟K线背离信号检测 ============\n")
	fmt.Printf("成功加载 %d 条15分钟K线数据\n\n", len(klines))

	// 计算技术指标
	fmt.Println("正在计算技术指标 (RSI14, MACD)...")
	klinesWithIndicators := indicators.CalculateIndicators(klines)
	if klinesWithIndicators == nil {
		fmt.Println("数据不足，无法计算指标（至少需要50根K线）")
		fmt.Println("请运行: make save-15m 获取更多数据")
		return
	}

	fmt.Printf("指标计算完成！\n\n")

	// 显示最后5根K线的指标
	fmt.Println("=== 最近5根K线指标 ===")
	printLastNIndicators(klinesWithIndicators, 5)

	// 扫描交易信号
	fmt.Println("\n=== 扫描15分钟交易信号 ===")
	signals := indicators.ScanSignals(klinesWithIndicators)

	if len(signals) == 0 {
		fmt.Println("未发现符合条件的交易信号")
		fmt.Println("\n信号条件：")
		fmt.Println("  做多: MACD金叉 + 前10根K线RSI<30 + 前10根最低价作止损")
		fmt.Println("  做空: MACD死叉 + 前10根K线RSI>70 + 前10根最高价作止损")
		return
	}

	fmt.Printf("发现 %d 个交易信号：\n\n", len(signals))

	// 统计信号
	longCount := 0
	shortCount := 0
	for _, signal := range signals {
		fmt.Println(signal.String())
		if signal.Type == indicators.SignalLong {
			longCount++
		} else {
			shortCount++
		}
	}

	fmt.Printf("\n=== 信号统计 ===\n")
	fmt.Printf("总信号数: %d\n", len(signals))
	fmt.Printf("做多信号: %d\n", longCount)
	fmt.Printf("做空信号: %d\n", shortCount)

	// 检测背离信号（核心功能）
	fmt.Println("\n========================================")
	fmt.Println("=== 15分钟背离信号检测（强烈反转信号！）===")
	fmt.Println("========================================")

	divergences := indicators.DetectDivergence(signals)

	if len(divergences) == 0 {
		fmt.Println("\n未发现背离信号")
		fmt.Println("\n💡 什么是背离？")
		fmt.Println("  背离是指价格走势与技术指标走势相反的现象，是强烈的趋势反转信号！")
		fmt.Println()
		fmt.Println("  🔺 看涨背离（底背离）：")
		fmt.Println("     价格创新低↓ 但 RSI/MACD未创新低（反而上涨↑）")
		fmt.Println("     → 说明：虽然价格在跌，但动能在增强 → 强烈买入信号！")
		fmt.Println()
		fmt.Println("  🔻 看跌背离（顶背离）：")
		fmt.Println("     价格创新高↑ 但 RSI/MACD未创新高（反而下跌↓）")
		fmt.Println("     → 说明：虽然价格在涨，但动能在减弱 → 强烈卖出信号！")
	} else {
		fmt.Printf("\n🎯 发现 %d 个背离信号（强烈反转信号！）\n\n", len(divergences))

		bullishCount := 0
		bearishCount := 0

		for i, div := range divergences {
			fmt.Printf("【背离信号 #%d】\n", i+1)
			fmt.Println(div.String())

			// 根据背离类型显示建议
			if div.Type == indicators.DivergenceBullish {
				fmt.Println("  💡 建议: 强烈买入信号！考虑在 " + fmt.Sprintf("%.2f", div.SecondSignal.Price) + " 附近做多")
				fmt.Println("  📍 止损: " + fmt.Sprintf("%.2f", div.SecondSignal.StopLoss))
				bullishCount++
			} else {
				fmt.Println("  💡 建议: 强烈卖出信号！考虑在 " + fmt.Sprintf("%.2f", div.SecondSignal.Price) + " 附近做空")
				fmt.Println("  📍 止损: " + fmt.Sprintf("%.2f", div.SecondSignal.StopLoss))
				bearishCount++
			}
			fmt.Println()
		}

		fmt.Printf("=== 背离统计 ===\n")
		fmt.Printf("🔺 看涨背离: %d 个 (价格↓ 指标↑ → 强烈买入信号)\n", bullishCount)
		fmt.Printf("🔻 看跌背离: %d 个 (价格↑ 指标↓ → 强烈卖出信号)\n", bearishCount)
		fmt.Printf("\n⚡ 注意：背离信号是最强烈的反转信号之一，建议重点关注！\n")
		fmt.Printf("📊 在15分钟级别，背离信号可用于日内交易和短线波段\n")
		fmt.Printf("📈 15分钟级别相比5分钟，信号更稳定，假信号更少\n")
	}

	fmt.Println("\n============================================")
}

// loadKlinesFromCSV 从CSV文件加载K线数据
func loadKlinesFromCSV(filename string) ([]indicators.KlineData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// 跳过表头
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	var klines []indicators.KlineData

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// CSV格式：交易对,时间间隔,开盘时间,开盘价,最高价,最低价,收盘价,成交量,收盘时间,成交额,成交笔数,主动买入量,主动买入额
		if len(record) < 9 {
			continue
		}

		// 解析时间（CSV中存储的是北京时间）
		openTime, err := time.ParseInLocation("2006-01-02 15:04:05", record[2], BeijingLocation)
		if err != nil {
			continue
		}
		closeTime, err := time.ParseInLocation("2006-01-02 15:04:05", record[8], BeijingLocation)
		if err != nil {
			continue
		}

		// 解析价格
		open, _ := strconv.ParseFloat(record[3], 64)
		high, _ := strconv.ParseFloat(record[4], 64)
		low, _ := strconv.ParseFloat(record[5], 64)
		close, _ := strconv.ParseFloat(record[6], 64)
		volume, _ := strconv.ParseFloat(record[7], 64)

		kline := indicators.KlineData{
			OpenTime:  openTime.UnixMilli(),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
			CloseTime: closeTime.UnixMilli(),
		}

		klines = append(klines, kline)
	}

	// 反转数据，使其按时间从旧到新排列
	for i, j := 0, len(klines)-1; i < j; i, j = i+1, j-1 {
		klines[i], klines[j] = klines[j], klines[i]
	}

	return klines, nil
}

// printLastNIndicators 打印最后N根K线的指标
func printLastNIndicators(klines []indicators.KlineWithIndicators, n int) {
	start := len(klines) - n
	if start < 0 {
		start = 0
	}

	fmt.Printf("%-20s %10s %10s %10s %10s %8s %10s %10s\n",
		"时间(北京)", "开盘", "最高", "最低", "收盘", "RSI14", "MACD", "信号线")
	fmt.Println("---------------------------------------------------------------------------------------------------")

	for i := start; i < len(klines); i++ {
		k := klines[i]
		timeStr := time.UnixMilli(k.CloseTime).In(BeijingLocation).Format("2006-01-02 15:04")

		crossInfo := ""
		if k.MacdCrossUp {
			crossInfo = " [金叉↑]"
		} else if k.MacdCrossDown {
			crossInfo = " [死叉↓]"
		}

		fmt.Printf("%-20s %10.2f %10.2f %10.2f %10.2f %8.2f %10.4f %10.4f%s\n",
			timeStr, k.Open, k.High, k.Low, k.Close, k.RSI14, k.MACD, k.MACDSignal, crossInfo)
	}
}
