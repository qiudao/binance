package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	apiID   = "ADU5yR0QPA6622twEtmPxVsW"
	apiKey  = "arodVpSl9kBFWSLkxK5gCZIFI6BV-GZ8PQWUbGS4ncI5GBhQ"
	baseURL = "https://www.bitmex.com/api/v1"
)

// Kline K线数据结构
type Kline struct {
	Timestamp time.Time `json:"timestamp"`
	Symbol    string    `json:"symbol"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
	Trades    int       `json:"trades"`
}

func generateSignature(method, path, expires string, apiSecret string) string {
	message := method + path + expires
	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// fetchPublicData 获取公开数据（K线数据是公开的）
func fetchPublicData(endpoint string) ([]byte, error) {
	url := baseURL + endpoint

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %s, body: %s", resp.Status, string(body))
	}

	return body, nil
}

// fetchAllKlines 获取所有K线数据
func fetchAllKlines(symbol string, binSize string, startTime string) ([]Kline, error) {
	var allKlines []Kline
	count := 1000 // BitMEX API每次最多返回1000条
	start := 0

	if startTime != "" {
		fmt.Printf("开始下载K线数据（从 %s 之后）...\n", startTime)
	} else {
		fmt.Println("开始下载所有历史K线数据...")
	}

	for {
		endpoint := fmt.Sprintf("/trade/bucketed?binSize=%s&partial=false&symbol=%s&count=%d&start=%d&reverse=false",
			binSize, symbol, count, start)

		if startTime != "" {
			endpoint += "&startTime=" + startTime
		}

		fmt.Printf("正在获取记录 %d-%d...\n", start, start+count)

		data, err := fetchPublicData(endpoint)
		if err != nil {
			return nil, fmt.Errorf("获取数据失败: %v", err)
		}

		var klines []Kline
		if err := json.Unmarshal(data, &klines); err != nil {
			return nil, fmt.Errorf("解析数据失败: %v", err)
		}

		if len(klines) == 0 {
			fmt.Println("✓ 所有数据下载完成!")
			break
		}

		allKlines = append(allKlines, klines...)
		fmt.Printf("  已获取 %d 条记录，总计: %d 条\n", len(klines), len(allKlines))

		if len(klines) < count {
			fmt.Println("✓ 已到达最后一页")
			break
		}

		start += count
		time.Sleep(500 * time.Millisecond)
	}

	return allKlines, nil
}

// getLastTimestampFromCSV 从CSV文件读取最后一条记录的时间戳
func getLastTimestampFromCSV(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// 跳过表头
	_, err = reader.Read()
	if err != nil {
		return "", err
	}

	var lastTimestamp string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if len(record) > 0 {
			lastTimestamp = record[0] // Timestamp在第一列
		}
	}

	return lastTimestamp, nil
}

// saveToCSV 保存K线数据为CSV
func saveToCSV(klines []Kline, filename string, appendMode bool) error {
	var file *os.File
	var err error

	if appendMode {
		file, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("打开文件失败: %v", err)
		}
	} else {
		file, err = os.Create(filename)
		if err != nil {
			return fmt.Errorf("创建文件失败: %v", err)
		}
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if !appendMode {
		header := []string{"Timestamp", "Symbol", "Open", "High", "Low", "Close", "Volume", "Trades"}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("写入表头失败: %v", err)
		}
	}

	for _, kline := range klines {
		record := []string{
			kline.Timestamp.Format(time.RFC3339),
			kline.Symbol,
			fmt.Sprintf("%.2f", kline.Open),
			fmt.Sprintf("%.2f", kline.High),
			fmt.Sprintf("%.2f", kline.Low),
			fmt.Sprintf("%.2f", kline.Close),
			fmt.Sprintf("%d", kline.Volume),
			fmt.Sprintf("%d", kline.Trades),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("写入记录失败: %v", err)
		}
	}

	return nil
}

func main() {
	symbol := flag.String("symbol", "XBTUSD", "交易对符号 (XBTUSD, ETHUSD, etc.)")
	binSize := flag.String("timeframe", "1d", "时间周期 (1m, 5m, 1h, 1d)")
	updateMode := flag.Bool("update", false, "增量更新模式")
	flag.Parse()

	fmt.Println("=== BitMEX K线数据下载工具 ===\n")

	filename := fmt.Sprintf("klines_%s_%s.csv", *symbol, *binSize)
	var startTime string
	var appendMode bool

	if *updateMode {
		fmt.Println("运行模式: 增量更新")

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			fmt.Println("⚠ 未找到现有文件，将进行全量下载")
			*updateMode = false
		} else {
			fmt.Printf("找到现有文件: %s\n", filename)

			lastTime, err := getLastTimestampFromCSV(filename)
			if err != nil {
				fmt.Printf("❌ 读取CSV文件失败: %v\n", err)
				return
			}

			if lastTime == "" {
				fmt.Println("⚠ CSV文件为空，将进行全量下载")
				*updateMode = false
			} else {
				startTime = lastTime
				appendMode = true
				fmt.Printf("最后记录时间: %s\n\n", lastTime)
			}
		}
	}

	if !*updateMode {
		fmt.Println("运行模式: 全量下载")

		if _, err := os.Stat(filename); err == nil {
			backupName := filename + ".bak"
			if err := os.Rename(filename, backupName); err != nil {
				fmt.Printf("⚠ 备份文件失败: %v\n", err)
			} else {
				fmt.Printf("✓ 已备份现有文件: %s\n", backupName)
			}
		}
		fmt.Println()
	}

	// 下载K线数据
	allKlines, err := fetchAllKlines(*symbol, *binSize, startTime)
	if err != nil {
		fmt.Printf("❌ 下载失败: %v\n", err)
		return
	}

	// 过滤重复记录
	if appendMode && startTime != "" && len(allKlines) > 0 {
		var filteredKlines []Kline
		for _, kline := range allKlines {
			if kline.Timestamp.Format(time.RFC3339) != startTime {
				filteredKlines = append(filteredKlines, kline)
			}
		}
		allKlines = filteredKlines
		fmt.Printf("过滤重复记录后剩余: %d 条\n", len(allKlines))
	}

	if len(allKlines) == 0 {
		fmt.Println("\n✓ 没有新的K线数据")
		return
	}

	fmt.Printf("\n总共下载了 %d 条K线记录\n", len(allKlines))

	if len(allKlines) > 0 {
		fmt.Println("\n第一条记录:")
		fmt.Printf("  时间: %s\n", allKlines[0].Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("  开: %.2f  高: %.2f  低: %.2f  收: %.2f\n",
			allKlines[0].Open, allKlines[0].High, allKlines[0].Low, allKlines[0].Close)

		fmt.Println("\n最后一条记录:")
		last := allKlines[len(allKlines)-1]
		fmt.Printf("  时间: %s\n", last.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("  开: %.2f  高: %.2f  低: %.2f  收: %.2f\n",
			last.Open, last.High, last.Low, last.Close)
	}

	fmt.Printf("\n正在保存到文件: %s\n", filename)

	if err := saveToCSV(allKlines, filename, appendMode); err != nil {
		fmt.Printf("❌ 保存失败: %v\n", err)
		return
	}

	if appendMode {
		fmt.Printf("✓ 成功追加 %d 条新记录到 %s\n", len(allKlines), filename)
	} else {
		fmt.Printf("✓ 成功保存 %d 条记录到 %s\n", len(allKlines), filename)
	}
}
