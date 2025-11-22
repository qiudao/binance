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
	"strconv"
	"time"
)

const (
	BaseURL = "https://api.binance.com"
)

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

func SaveToCSV(klines []Kline, filename string, symbol string, interval string) error {
	// 确保目录存在
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}

	// 检查文件是否存在
	fileExists := false
	if _, err := os.Stat(filename); err == nil {
		fileExists = true
	}

	// 以追加模式打开文件
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 如果文件是新创建的，写入表头
	if !fileExists {
		header := []string{
			"交易对", "时间间隔", "开盘时间", "开盘价", "最高价", "最低价", "收盘价",
			"成交量", "收盘时间", "成交额", "成交笔数", "主动买入量", "主动买入额",
		}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("写入表头失败: %w", err)
		}
	}

	// 写入数据
	for _, kline := range klines {
		record := []string{
			symbol,
			interval,
			time.UnixMilli(kline.OpenTime).Format("2006-01-02 15:04:05"),
			kline.Open,
			kline.High,
			kline.Low,
			kline.Close,
			kline.Volume,
			time.UnixMilli(kline.CloseTime).Format("2006-01-02 15:04:05"),
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

	klines, err := GetKlines(*symbol, *interval, 0, 0, *limit)
	if err != nil {
		fmt.Printf("获取K线数据失败: %v\n", err)
		return
	}

	fmt.Printf("成功获取 %d 条 %s %s K线数据\n", len(klines), *symbol, *interval)

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
		openTime := time.UnixMilli(kline.OpenTime).Format("2006-01-02 15:04:05")
		closeTime := time.UnixMilli(kline.CloseTime).Format("2006-01-02 15:04:05")
		fmt.Printf("时间: %s - %s\n", openTime, closeTime)
		fmt.Printf("  开: %s, 高: %s, 低: %s, 收: %s, 量: %s\n",
			kline.Open, kline.High, kline.Low, kline.Close, kline.Volume)
		fmt.Printf("  成交笔数: %d\n\n", kline.NumberOfTrades)
	}

	if len(klines) > 5 {
		fmt.Printf("... 还有 %d 条数据\n", len(klines)-5)
	}
}
