package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

const (
	BaseURL = "https://api.binance.com"
)

// 北京时间时区
var BeijingLocation = time.FixedZone("CST", 8*3600)

type Kline struct {
	OpenTime                 int64
	Open                     string
	High                     string
	Low                      string
	Close                    string
	Volume                   string
	CloseTime                int64
	QuoteAssetVolume         string
	NumberOfTrades           int
	TakerBuyBaseAssetVolume  string
	TakerBuyQuoteAssetVolume string
	Ignore                   string
}

func GetKlines(symbol string, interval string, startTime, endTime int64, limit int) ([]Kline, error) {
	url := fmt.Sprintf("%s/api/v3/klines?symbol=%s&interval=%s", BaseURL, symbol, interval)

	if startTime > 0 {
		url += fmt.Sprintf("&startTime=%d", startTime)
	}
	if endTime > 0 {
		url += fmt.Sprintf("&endTime=%d", endTime)
	}
	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回错误: %s, 状态码: %d", string(body), resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var rawKlines [][]interface{}
	if err := json.Unmarshal(body, &rawKlines); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	klines := make([]Kline, len(rawKlines))
	for i, raw := range rawKlines {
		klines[i] = Kline{
			OpenTime:                 int64(raw[0].(float64)),
			Open:                     raw[1].(string),
			High:                     raw[2].(string),
			Low:                      raw[3].(string),
			Close:                    raw[4].(string),
			Volume:                   raw[5].(string),
			CloseTime:                int64(raw[6].(float64)),
			QuoteAssetVolume:         raw[7].(string),
			NumberOfTrades:           int(raw[8].(float64)),
			TakerBuyBaseAssetVolume:  raw[9].(string),
			TakerBuyQuoteAssetVolume: raw[10].(string),
			Ignore:                   raw[11].(string),
		}
	}

	return klines, nil
}

// GetKlinesBatch 批量获取K线数据，支持超过1000条的请求
func GetKlinesBatch(symbol string, interval string, totalLimit int) ([]Kline, error) {
	if totalLimit <= 1000 {
		return GetKlines(symbol, interval, 0, 0, totalLimit)
	}

	var allKlines []Kline
	seen := make(map[int64]bool) // 用于去重
	batchSize := 1000
	batches := (totalLimit + batchSize - 1) / batchSize // 向上取整
	var endTime int64 = 0 // 0表示当前时间

	for i := 0; i < batches; i++ {
		currentBatch := i + 1
		fmt.Printf("正在获取第 %d/%d 批...\n", currentBatch, batches)

		// 带重试的获取逻辑
		var klines []Kline
		var err error
		maxRetries := 3

		for retry := 0; retry < maxRetries; retry++ {
			klines, err = GetKlines(symbol, interval, 0, endTime, batchSize)
			if err == nil {
				break
			}

			if retry < maxRetries-1 {
				waitTime := time.Duration(retry+1) * time.Second
				fmt.Printf("  请求失败，%v 后重试 (%d/%d)...\n", waitTime, retry+1, maxRetries)
				time.Sleep(waitTime)
			}
		}

		if err != nil {
			return allKlines, fmt.Errorf("批次 %d 获取失败: %w", currentBatch, err)
		}

		if len(klines) == 0 {
			fmt.Printf("  第 %d 批未获取到数据，停止\n", currentBatch)
			break
		}

		// 去重并添加到结果集
		addedCount := 0
		for _, kline := range klines {
			if !seen[kline.OpenTime] {
				seen[kline.OpenTime] = true
				allKlines = append(allKlines, kline)
				addedCount++
			}
		}

		fmt.Printf("  第 %d 批获取 %d 条，去重后添加 %d 条，累计 %d/%d 条\n",
			currentBatch, len(klines), addedCount, len(allKlines), totalLimit)

		// 如果已经获取足够数据，停止
		if len(allKlines) >= totalLimit {
			break
		}

		// 更新 endTime 为当前批次最早的时间 - 1ms（第一条是最早的）
		if len(klines) > 0 {
			earliestTime := klines[0].OpenTime
			endTime = earliestTime - 1
		}

		// 添加延迟避免触发API限流（除了最后一批）
		if i < batches-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	// 按时间从新到旧排序（最新的在前面）
	sort.Slice(allKlines, func(i, j int) bool {
		return allKlines[i].OpenTime > allKlines[j].OpenTime
	})

	// 截取到指定数量
	if len(allKlines) > totalLimit {
		allKlines = allKlines[:totalLimit]
	}

	return allKlines, nil
}

func SaveToCSV(klines []Kline, filename string, symbol string, interval string) error {
	// 确保目录存在
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}

	// 以覆盖模式创建文件（而不是追加模式）
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	header := []string{
		"交易对", "时间间隔", "开盘时间", "开盘价", "最高价", "最低价", "收盘价",
		"成交量", "收盘时间", "成交额", "成交笔数", "主动买入量", "主动买入额",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("写入表头失败: %w", err)
	}

	// 写入数据
	for _, kline := range klines {
		record := []string{
			symbol,
			interval,
			time.UnixMilli(kline.OpenTime).In(BeijingLocation).Format("2006-01-02 15:04:05"),
			kline.Open,
			kline.High,
			kline.Low,
			kline.Close,
			kline.Volume,
			time.UnixMilli(kline.CloseTime).In(BeijingLocation).Format("2006-01-02 15:04:05"),
			kline.QuoteAssetVolume,
			strconv.Itoa(kline.NumberOfTrades),
			kline.TakerBuyBaseAssetVolume,
			kline.TakerBuyQuoteAssetVolume,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("写入数据失败: %w", err)
		}
	}

	return nil
}

func main() {
	// 命令行参数
	symbol := flag.String("symbol", "BTCUSDT", "交易对")
	interval := flag.String("interval", "1m", "K线间隔 (1m, 5m, 15m, 1h, 4h, 1d)")
	limit := flag.Int("limit", 100, "获取K线数量")
	output := flag.String("output", "", "输出CSV文件路径（不指定则打印到屏幕）")
	flag.Parse()

	// 使用批量获取函数，自动处理超过1000条的情况
	klines, err := GetKlinesBatch(*symbol, *interval, *limit)
	if err != nil {
		fmt.Printf("获取K线数据失败: %v\n", err)
		return
	}

	fmt.Printf("\n成功获取 %d 条 %s %s K线数据\n", len(klines), *symbol, *interval)

	// 如果指定了输出文件，保存到CSV
	if *output != "" {
		if err := SaveToCSV(klines, *output, *symbol, *interval); err != nil {
			fmt.Printf("保存到CSV失败: %v\n", err)
			return
		}
		fmt.Printf("数据已保存到: %s\n", *output)
		return
	}

	// 否则打印到屏幕
	fmt.Println()
	for i, kline := range klines {
		if i >= 5 {
			break
		}
		openTime := time.UnixMilli(kline.OpenTime).In(BeijingLocation).Format("2006-01-02 15:04:05")
		closeTime := time.UnixMilli(kline.CloseTime).In(BeijingLocation).Format("2006-01-02 15:04:05")
		fmt.Printf("时间: %s - %s (北京时间)\n", openTime, closeTime)
		fmt.Printf("  开: %s, 高: %s, 低: %s, 收: %s, 量: %s\n",
			kline.Open, kline.High, kline.Low, kline.Close, kline.Volume)
		fmt.Printf("  成交笔数: %d\n\n", kline.NumberOfTrades)
	}

	if len(klines) > 5 {
		fmt.Printf("... 还有 %d 条数据\n", len(klines)-5)
	}
}
