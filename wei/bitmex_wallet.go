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
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	apiID     = "ADU5yR0QPA6622twEtmPxVsW"
	apiKey    = "arodVpSl9kBFWSLkxK5gCZIFI6BV-GZ8PQWUbGS4ncI5GBhQ"
	baseURL   = "https://www.bitmex.com/api/v1"
)

// WalletHistory 结构表示钱包历史记录
type WalletHistory struct {
	TransactID     string  `json:"transactID"`
	TransactType   string  `json:"transactType"`
	TransactStatus string  `json:"transactStatus"`
	Account        int     `json:"account"`
	Currency       string  `json:"currency"`
	Amount         int64   `json:"amount"`          // Satoshi
	Fee            int64   `json:"fee"`             // Satoshi
	WalletBalance  int64   `json:"walletBalance"`   // Satoshi - 关键字段：钱包余额
	MarginBalance  int64   `json:"marginBalance"`   // Satoshi
	Timestamp      string  `json:"timestamp"`
	Address        string  `json:"address"`
	Tx             string  `json:"tx"`
	Text           string  `json:"text"`
}

func generateSignature(method, path, expires string, apiSecret string) string {
	message := method + path + expires
	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// fetchPrivateData 获取私有数据，需要认证
func fetchPrivateData(endpoint string) ([]byte, error) {
	url := baseURL + endpoint

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 生成签名所需的 expires 时间戳（当前时间 + 60 秒）
	expires := fmt.Sprintf("%d", time.Now().Unix()+60)

	verb := "GET"
	fullPath := "/api/v1" + endpoint

	req.Header.Set("api-key", apiID)
	signature := generateSignature(verb, fullPath, expires, apiKey)
	req.Header.Set("api-signature", signature)
	req.Header.Set("api-expires", expires)

	client := &http.Client{Timeout: 10 * time.Second}
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

		// Timestamp 在第10列（索引9）
		if len(record) > 9 {
			lastTimestamp = record[9]
		}
	}

	return lastTimestamp, nil
}

// fetchAllWalletHistory 获取所有钱包历史记录（分页）
func fetchAllWalletHistory(startTime string) ([]WalletHistory, error) {
	var allHistory []WalletHistory
	count := 10000 // BitMEX API默认值
	start := 0

	if startTime != "" {
		fmt.Printf("开始下载增量钱包历史记录（从 %s 之后）...\n", startTime)
	} else {
		fmt.Println("开始下载所有钱包历史记录...")
	}

	for {
		// 构建分页请求
		endpoint := fmt.Sprintf("/user/walletHistory?count=%d&start=%d&reverse=false", count, start)

		// 如果指定了开始时间，添加到查询参数
		if startTime != "" {
			endpoint += "&startTime=" + url.QueryEscape(startTime)
		}

		fmt.Printf("正在获取记录 %d-%d...\n", start, start+count)

		data, err := fetchPrivateData(endpoint)
		if err != nil {
			return nil, fmt.Errorf("获取数据失败: %v", err)
		}

		// 解析数据
		var history []WalletHistory
		if err := json.Unmarshal(data, &history); err != nil {
			return nil, fmt.Errorf("解析数据失败: %v", err)
		}

		// 没有更多数据了
		if len(history) == 0 {
			fmt.Println("✓ 所有数据下载完成!")
			break
		}

		allHistory = append(allHistory, history...)
		fmt.Printf("  已获取 %d 条记录，总计: %d 条\n", len(history), len(allHistory))

		// 如果返回的记录数少于请求数，说明已经是最后一页
		if len(history) < count {
			fmt.Println("✓ 已到达最后一页")
			break
		}

		start += count

		// 添加延迟避免API限流
		time.Sleep(500 * time.Millisecond)
	}

	return allHistory, nil
}

// satoshiToBTC 将 Satoshi 转换为 BTC（保留8位小数）
func satoshiToBTC(satoshi int64) string {
	btc := float64(satoshi) / 100000000.0
	return fmt.Sprintf("%.8f", btc)
}

// saveToCSV 将钱包历史记录保存为CSV文件
func saveToCSV(history []WalletHistory, filename string, appendMode bool) error {
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

	// 只在新建模式下写入表头
	if !appendMode {
		header := []string{
			"TransactID", "TransactType", "TransactStatus", "Account", "Currency",
			"Amount_Satoshi", "Amount_BTC", "Fee_Satoshi", "Fee_BTC",
			"Timestamp", "WalletBalance_Satoshi", "WalletBalance_BTC",
			"MarginBalance_Satoshi", "MarginBalance_BTC",
			"Address", "Tx", "Text",
		}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("写入表头失败: %v", err)
		}
	}

	// 写入每条记录
	for _, h := range history {
		record := []string{
			h.TransactID,
			h.TransactType,
			h.TransactStatus,
			strconv.Itoa(h.Account),
			h.Currency,
			strconv.FormatInt(h.Amount, 10),
			satoshiToBTC(h.Amount),
			strconv.FormatInt(h.Fee, 10),
			satoshiToBTC(h.Fee),
			h.Timestamp,
			strconv.FormatInt(h.WalletBalance, 10),
			satoshiToBTC(h.WalletBalance),
			strconv.FormatInt(h.MarginBalance, 10),
			satoshiToBTC(h.MarginBalance),
			h.Address,
			h.Tx,
			h.Text,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("写入记录失败: %v", err)
		}
	}

	return nil
}

func main() {
	// 解析命令行参数
	updateMode := flag.Bool("update", false, "增量更新模式（只下载新记录）")
	flag.Parse()

	fmt.Println("=== BitMEX 钱包历史记录下载工具 ===\n")

	// 测试API连接
	fmt.Println("测试API连接...")
	testEndpoint := "/user/walletHistory?count=1"
	_, err := fetchPrivateData(testEndpoint)
	if err != nil {
		fmt.Printf("❌ API连接失败: %v\n", err)
		fmt.Println("\n提示: 请检查API凭证是否正确")
		return
	}
	fmt.Println("✓ API连接成功!\n")

	var startTime string
	var filename = "wallet.csv"
	var appendMode bool

	if *updateMode {
		// 增量更新模式
		fmt.Println("运行模式: 增量更新")

		// 检查 wallet.csv 是否存在
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			fmt.Println("⚠ 未找到 wallet.csv 文件，将进行全量下载")
			*updateMode = false
		} else {
			fmt.Printf("找到现有文件: %s\n", filename)

			// 读取最后一条记录的时间
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
		// 全量下载模式
		fmt.Println("运行模式: 全量下载")

		// 如果文件已存在，备份为 .bak
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

	// 下载钱包历史记录
	allHistory, err := fetchAllWalletHistory(startTime)
	if err != nil {
		fmt.Printf("❌ 下载失败: %v\n", err)
		return
	}

	// 如果是增量更新，过滤掉与最后一条记录时间相同的记录（可能重复）
	if appendMode && startTime != "" && len(allHistory) > 0 {
		var filteredHistory []WalletHistory
		for _, h := range allHistory {
			if h.Timestamp != startTime {
				filteredHistory = append(filteredHistory, h)
			}
		}
		allHistory = filteredHistory
		fmt.Printf("过滤重复记录后剩余: %d 条\n", len(allHistory))
	}

	if len(allHistory) == 0 {
		fmt.Println("\n✓ 没有新的钱包历史记录")
		return
	}

	fmt.Printf("\n总共下载了 %d 条钱包历史记录\n", len(allHistory))

	// 显示统计信息
	if len(allHistory) > 0 {
		fmt.Println("\n第一条记录:")
		fmt.Printf("  时间: %s\n", allHistory[0].Timestamp)
		fmt.Printf("  类型: %s\n", allHistory[0].TransactType)
		fmt.Printf("  金额: %s BTC\n", satoshiToBTC(allHistory[0].Amount))
		fmt.Printf("  钱包余额: %s BTC\n", satoshiToBTC(allHistory[0].WalletBalance))

		fmt.Println("\n最后一条记录:")
		last := allHistory[len(allHistory)-1]
		fmt.Printf("  时间: %s\n", last.Timestamp)
		fmt.Printf("  类型: %s\n", last.TransactType)
		fmt.Printf("  金额: %s BTC\n", satoshiToBTC(last.Amount))
		fmt.Printf("  钱包余额: %s BTC\n", satoshiToBTC(last.WalletBalance))
	}

	// 保存为CSV文件
	fmt.Printf("\n正在保存到文件: %s\n", filename)

	if err := saveToCSV(allHistory, filename, appendMode); err != nil {
		fmt.Printf("❌ 保存失败: %v\n", err)
		return
	}

	if appendMode {
		fmt.Printf("✓ 成功追加 %d 条新记录到 %s\n", len(allHistory), filename)
	} else {
		fmt.Printf("✓ 成功保存 %d 条记录到 %s\n", len(allHistory), filename)
	}
}
