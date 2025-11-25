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
    "strings"
    "time"
)

const (
    apiID     = "ADU5yR0QPA6622twEtmPxVsW"
    apiKey    = "arodVpSl9kBFWSLkxK5gCZIFI6BV-GZ8PQWUbGS4ncI5GBhQ"
    baseURL   = "https://www.bitmex.com/api/v1"
)

// Execution 结构表示单笔交易记录（详细信息）
type Execution struct {
    ExecID          string    `json:"execID"`
    OrderID         string    `json:"orderID"`
    ClOrdID         string    `json:"clOrdID"`
    Account         int       `json:"account"`
    Symbol          string    `json:"symbol"`
    Side            string    `json:"side"`
    LastQty         int       `json:"lastQty"`
    LastPx          float64   `json:"lastPx"`
    OrderQty        int       `json:"orderQty"`
    Price           float64   `json:"price"`
    LeavesQty       int       `json:"leavesQty"`
    CumQty          int       `json:"cumQty"`
    AvgPx           float64   `json:"avgPx"`
    Commission      float64   `json:"commission"`
    TransactTime    string    `json:"transactTime"`
    Timestamp       string    `json:"timestamp"`
    OrdType         string    `json:"ordType"`
    ExecType        string    `json:"execType"`
    OrdStatus       string    `json:"ordStatus"`
    Currency        string    `json:"currency"`
    Text            string    `json:"text"`
}

// TradeBucket 结构表示 K 线数据
type TradeBucket struct {
    Timestamp time.Time `json:"timestamp"`
    Open      float64   `json:"open"`
    High      float64   `json:"high"`
    Low       float64   `json:"low"`
    Close     float64   `json:"close"`
    Volume    int       `json:"volume"`
}

func generateSignature(method, path, expires string, apiSecret string) string {
    // 注意：@coolish 未提供 apiSecret，如果需要签名，需联系他获取
    message := method + path + expires
    mac := hmac.New(sha256.New, []byte(apiSecret))
    mac.Write([]byte(message))
    return hex.EncodeToString(mac.Sum(nil))
}

// fetchPublicData 获取公开数据，不需要认证
func fetchPublicData(endpoint string) ([]byte, error) {
    // 请求 URL
    url := baseURL + endpoint

    // 请求头
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    // 发送请求
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // 读取响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API request failed with status: %s, body: %s", resp.Status, string(body))
    }

    return body, nil
}

// fetchPrivateData 获取私有数据，需要认证
func fetchPrivateData(endpoint string) ([]byte, error) {
    // 请求 URL
    url := baseURL + endpoint

    // 请求头
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    // 生成签名所需的 expires 时间戳（当前时间 + 60 秒）
    expires := fmt.Sprintf("%d", time.Now().Unix()+60)

    // 签名需要使用完整路径（包括 /api/v1）
    verb := "GET"
    fullPath := "/api/v1" + endpoint

    req.Header.Set("api-key", apiID)
    signature := generateSignature(verb, fullPath, expires, apiKey)
    req.Header.Set("api-signature", signature)
    req.Header.Set("api-expires", expires)

    // 发送请求
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // 读取响应
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
            return "", nil // 文件不存在，返回空字符串
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

        // TransactTime 在第14列（索引14）
        if len(record) > 14 {
            lastTimestamp = record[14]
        }
    }

    return lastTimestamp, nil
}

// findLatestCSV 查找最新的CSV文件
func findLatestCSV() string {
    files, err := os.ReadDir(".")
    if err != nil {
        return ""
    }

    var latestFile string
    var latestTime time.Time

    for _, file := range files {
        if !file.IsDir() && strings.HasPrefix(file.Name(), "bitmex_executions_") && strings.HasSuffix(file.Name(), ".csv") {
            info, err := file.Info()
            if err != nil {
                continue
            }
            if info.ModTime().After(latestTime) {
                latestTime = info.ModTime()
                latestFile = file.Name()
            }
        }
    }

    return latestFile
}

// fetchAllExecutions 获取所有历史交易记录（分页）
func fetchAllExecutions(startTime string) ([]Execution, error) {
    var allExecutions []Execution
    count := 500 // BitMEX API每次最多返回500条
    start := 0

    if startTime != "" {
        fmt.Printf("开始下载增量交易记录（从 %s 之后）...\n", startTime)
    } else {
        fmt.Println("开始下载所有历史交易记录...")
    }

    for {
        // 构建分页请求
        endpoint := fmt.Sprintf("/execution?count=%d&start=%d&reverse=false", count, start)

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
        var executions []Execution
        if err := json.Unmarshal(data, &executions); err != nil {
            return nil, fmt.Errorf("解析数据失败: %v", err)
        }

        // 没有更多数据了
        if len(executions) == 0 {
            fmt.Println("✓ 所有数据下载完成!")
            break
        }

        allExecutions = append(allExecutions, executions...)
        fmt.Printf("  已获取 %d 条记录，总计: %d 条\n", len(executions), len(allExecutions))

        // 如果返回的记录数少于请求数，说明已经是最后一页
        if len(executions) < count {
            fmt.Println("✓ 已到达最后一页")
            break
        }

        start += count

        // 添加延迟避免API限流
        time.Sleep(500 * time.Millisecond)
    }

    return allExecutions, nil
}

// saveToCSV 将交易记录保存为CSV文件
func saveToCSV(executions []Execution, filename string, appendMode bool) error {
    var file *os.File
    var err error

    if appendMode {
        // 追加模式：打开文件，指针移到文件末尾
        file, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
        if err != nil {
            return fmt.Errorf("打开文件失败: %v", err)
        }
    } else {
        // 新建模式：创建新文件
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
            "ExecID", "OrderID", "ClOrdID", "Account", "Symbol", "Side",
            "LastQty", "LastPx", "OrderQty", "Price", "LeavesQty", "CumQty",
            "AvgPx", "Commission", "TransactTime", "Timestamp", "OrdType",
            "ExecType", "OrdStatus", "Currency", "Text",
        }
        if err := writer.Write(header); err != nil {
            return fmt.Errorf("写入表头失败: %v", err)
        }
    }

    // 写入每条记录
    for _, exec := range executions {
        record := []string{
            exec.ExecID,
            exec.OrderID,
            exec.ClOrdID,
            strconv.Itoa(exec.Account),
            exec.Symbol,
            exec.Side,
            strconv.Itoa(exec.LastQty),
            fmt.Sprintf("%.2f", exec.LastPx),
            strconv.Itoa(exec.OrderQty),
            fmt.Sprintf("%.2f", exec.Price),
            strconv.Itoa(exec.LeavesQty),
            strconv.Itoa(exec.CumQty),
            fmt.Sprintf("%.2f", exec.AvgPx),
            fmt.Sprintf("%.8f", exec.Commission),
            exec.TransactTime,
            exec.Timestamp,
            exec.OrdType,
            exec.ExecType,
            exec.OrdStatus,
            exec.Currency,
            exec.Text,
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

    fmt.Println("=== BitMEX 历史交易记录下载工具 ===\n")

    // 测试API连接
    fmt.Println("测试API连接...")
    testEndpoint := "/execution?count=1"
    _, err := fetchPrivateData(testEndpoint)
    if err != nil {
        fmt.Printf("❌ API连接失败: %v\n", err)
        fmt.Println("\n提示: 请检查API凭证是否正确")
        return
    }
    fmt.Println("✓ API连接成功!\n")

    var startTime string
    var filename string
    var appendMode bool

    if *updateMode {
        // 增量更新模式
        fmt.Println("运行模式: 增量更新")

        // 查找最新的CSV文件
        filename = findLatestCSV()
        if filename == "" {
            fmt.Println("⚠ 未找到现有CSV文件，将进行全量下载")
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
        fmt.Println("运行模式: 全量下载\n")
        filename = fmt.Sprintf("bitmex_executions_%s.csv", time.Now().Format("20060102_150405"))
    }

    // 下载交易记录
    allExecutions, err := fetchAllExecutions(startTime)
    if err != nil {
        fmt.Printf("❌ 下载失败: %v\n", err)
        return
    }

    // 如果是增量更新，过滤掉与最后一条记录时间相同的记录（可能重复）
    if appendMode && startTime != "" && len(allExecutions) > 0 {
        var filteredExecutions []Execution
        for _, exec := range allExecutions {
            if exec.TransactTime != startTime {
                filteredExecutions = append(filteredExecutions, exec)
            }
        }
        allExecutions = filteredExecutions
        fmt.Printf("过滤重复记录后剩余: %d 条\n", len(allExecutions))
    }

    if len(allExecutions) == 0 {
        fmt.Println("\n✓ 没有新的交易记录")
        return
    }

    fmt.Printf("\n总共下载了 %d 条交易记录\n", len(allExecutions))

    // 显示统计信息
    fmt.Println("\n第一条记录:")
    fmt.Printf("  时间: %s\n", allExecutions[0].TransactTime)
    fmt.Printf("  交易对: %s\n", allExecutions[0].Symbol)

    fmt.Println("\n最后一条记录:")
    last := allExecutions[len(allExecutions)-1]
    fmt.Printf("  时间: %s\n", last.TransactTime)
    fmt.Printf("  交易对: %s\n", last.Symbol)

    // 保存为CSV文件
    fmt.Printf("\n正在保存到文件: %s\n", filename)

    if err := saveToCSV(allExecutions, filename, appendMode); err != nil {
        fmt.Printf("❌ 保存失败: %v\n", err)
        return
    }

    if appendMode {
        fmt.Printf("✓ 成功追加 %d 条新记录到 %s\n", len(allExecutions), filename)
    } else {
        fmt.Printf("✓ 成功保存 %d 条记录到 %s\n", len(allExecutions), filename)
    }
}
