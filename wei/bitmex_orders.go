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

// Order 结构表示单个订单记录
type Order struct {
    OrderID            string    `json:"orderID"`
    ClOrdID            string    `json:"clOrdID"`
    ClOrdLinkID        string    `json:"clOrdLinkID"`
    Account            int       `json:"account"`
    Symbol             string    `json:"symbol"`
    Side               string    `json:"side"`
    SimpleOrderQty     float64   `json:"simpleOrderQty"`
    OrderQty           int       `json:"orderQty"`
    Price              float64   `json:"price"`
    DisplayQty         int       `json:"displayQty"`
    StopPx             float64   `json:"stopPx"`
    PegOffsetValue     float64   `json:"pegOffsetValue"`
    PegPriceType       string    `json:"pegPriceType"`
    Currency           string    `json:"currency"`
    SettlCurrency      string    `json:"settlCurrency"`
    OrdType            string    `json:"ordType"`
    TimeInForce        string    `json:"timeInForce"`
    ExecInst           string    `json:"execInst"`
    ContingencyType    string    `json:"contingencyType"`
    ExDestination      string    `json:"exDestination"`
    OrdStatus          string    `json:"ordStatus"`
    Triggered          string    `json:"triggered"`
    WorkingIndicator   bool      `json:"workingIndicator"`
    OrdRejReason       string    `json:"ordRejReason"`
    SimpleLeavesQty    float64   `json:"simpleLeavesQty"`
    LeavesQty          int       `json:"leavesQty"`
    SimpleCumQty       float64   `json:"simpleCumQty"`
    CumQty             int       `json:"cumQty"`
    AvgPx              float64   `json:"avgPx"`
    MultiLegReportingType string `json:"multiLegReportingType"`
    Text               string    `json:"text"`
    TransactTime       string    `json:"transactTime"`
    Timestamp          string    `json:"timestamp"`
}

func generateSignature(method, path, expires string, apiSecret string) string {
    message := method + path + expires
    mac := hmac.New(sha256.New, []byte(apiSecret))
    mac.Write([]byte(message))
    return hex.EncodeToString(mac.Sum(nil))
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

        // Timestamp 在最后一列
        if len(record) > 0 {
            lastTimestamp = record[len(record)-1]
        }
    }

    return lastTimestamp, nil
}

// fetchAllOrders 获取所有历史订单记录（分页）
func fetchAllOrders(startTime string) ([]Order, error) {
    var allOrders []Order
    count := 500 // BitMEX API每次最多返回500条
    start := 0

    if startTime != "" {
        fmt.Printf("开始下载增量订单记录（从 %s 之后）...\n", startTime)
    } else {
        fmt.Println("开始下载所有历史订单记录...")
    }

    for {
        // 构建分页请求
        endpoint := fmt.Sprintf("/order?count=%d&start=%d&reverse=false", count, start)

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
        var orders []Order
        if err := json.Unmarshal(data, &orders); err != nil {
            return nil, fmt.Errorf("解析数据失败: %v", err)
        }

        // 没有更多数据了
        if len(orders) == 0 {
            fmt.Println("✓ 所有数据下载完成!")
            break
        }

        allOrders = append(allOrders, orders...)
        fmt.Printf("  已获取 %d 条记录，总计: %d 条\n", len(orders), len(allOrders))

        // 如果返回的记录数少于请求数，说明已经是最后一页
        if len(orders) < count {
            fmt.Println("✓ 已到达最后一页")
            break
        }

        start += count

        // 添加延迟避免API限流
        time.Sleep(500 * time.Millisecond)
    }

    return allOrders, nil
}

// saveToCSV 将订单记录保存为CSV文件
func saveToCSV(orders []Order, filename string, appendMode bool) error {
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
            "OrderID", "ClOrdID", "ClOrdLinkID", "Account", "Symbol", "Side",
            "SimpleOrderQty", "OrderQty", "Price", "DisplayQty", "StopPx", "PegOffsetValue",
            "PegPriceType", "Currency", "SettlCurrency", "OrdType", "TimeInForce", "ExecInst",
            "ContingencyType", "ExDestination", "OrdStatus", "Triggered", "WorkingIndicator",
            "OrdRejReason", "SimpleLeavesQty", "LeavesQty", "SimpleCumQty", "CumQty",
            "AvgPx", "MultiLegReportingType", "Text", "TransactTime", "Timestamp",
        }
        if err := writer.Write(header); err != nil {
            return fmt.Errorf("写入表头失败: %v", err)
        }
    }

    // 写入每条记录
    for _, order := range orders {
        record := []string{
            order.OrderID,
            order.ClOrdID,
            order.ClOrdLinkID,
            strconv.Itoa(order.Account),
            order.Symbol,
            order.Side,
            fmt.Sprintf("%.8f", order.SimpleOrderQty),
            strconv.Itoa(order.OrderQty),
            fmt.Sprintf("%.2f", order.Price),
            strconv.Itoa(order.DisplayQty),
            fmt.Sprintf("%.2f", order.StopPx),
            fmt.Sprintf("%.2f", order.PegOffsetValue),
            order.PegPriceType,
            order.Currency,
            order.SettlCurrency,
            order.OrdType,
            order.TimeInForce,
            order.ExecInst,
            order.ContingencyType,
            order.ExDestination,
            order.OrdStatus,
            order.Triggered,
            strconv.FormatBool(order.WorkingIndicator),
            order.OrdRejReason,
            fmt.Sprintf("%.8f", order.SimpleLeavesQty),
            strconv.Itoa(order.LeavesQty),
            fmt.Sprintf("%.8f", order.SimpleCumQty),
            strconv.Itoa(order.CumQty),
            fmt.Sprintf("%.2f", order.AvgPx),
            order.MultiLegReportingType,
            order.Text,
            order.TransactTime,
            order.Timestamp,
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

    fmt.Println("=== BitMEX 历史订单记录下载工具 ===\n")

    // 测试API连接
    fmt.Println("测试API连接...")
    testEndpoint := "/order?count=1"
    _, err := fetchPrivateData(testEndpoint)
    if err != nil {
        fmt.Printf("❌ API连接失败: %v\n", err)
        fmt.Println("\n提示: 请检查API凭证是否正确")
        return
    }
    fmt.Println("✓ API连接成功!\n")

    var startTime string
    var filename = "orders.csv"
    var appendMode bool

    if *updateMode {
        // 增量更新模式
        fmt.Println("运行模式: 增量更新")

        // 检查 orders.csv 是否存在
        if _, err := os.Stat(filename); os.IsNotExist(err) {
            fmt.Println("⚠ 未找到 orders.csv 文件，将进行全量下载")
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

    // 下载订单记录
    allOrders, err := fetchAllOrders(startTime)
    if err != nil {
        fmt.Printf("❌ 下载失败: %v\n", err)
        return
    }

    // 如果是增量更新，过滤掉与最后一条记录时间相同的记录（可能重复）
    if appendMode && startTime != "" && len(allOrders) > 0 {
        var filteredOrders []Order
        for _, order := range allOrders {
            if order.Timestamp != startTime {
                filteredOrders = append(filteredOrders, order)
            }
        }
        allOrders = filteredOrders
        fmt.Printf("过滤重复记录后剩余: %d 条\n", len(allOrders))
    }

    if len(allOrders) == 0 {
        fmt.Println("\n✓ 没有新的订单记录")
        return
    }

    fmt.Printf("\n总共下载了 %d 条订单记录\n", len(allOrders))

    // 显示统计信息
    if len(allOrders) > 0 {
        fmt.Println("\n第一条记录:")
        fmt.Printf("  时间: %s\n", allOrders[0].TransactTime)
        fmt.Printf("  交易对: %s\n", allOrders[0].Symbol)
        fmt.Printf("  状态: %s\n", allOrders[0].OrdStatus)

        fmt.Println("\n最后一条记录:")
        last := allOrders[len(allOrders)-1]
        fmt.Printf("  时间: %s\n", last.TransactTime)
        fmt.Printf("  交易对: %s\n", last.Symbol)
        fmt.Printf("  状态: %s\n", last.OrdStatus)
    }

    // 保存为CSV文件
    fmt.Printf("\n正在保存到文件: %s\n", filename)

    if err := saveToCSV(allOrders, filename, appendMode); err != nil {
        fmt.Printf("❌ 保存失败: %v\n", err)
        return
    }

    if appendMode {
        fmt.Printf("✓ 成功追加 %d 条新记录到 %s\n", len(allOrders), filename)
    } else {
        fmt.Printf("✓ 成功保存 %d 条记录到 %s\n", len(allOrders), filename)
    }
}
