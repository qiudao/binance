package main

import (
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

const (
    apiID     = "ADU5yR0QPA6622twEtmPxVsW"
    apiKey    = "arodVpSl9kBFWSLkxK5gCZIFI6BV-GZ8PQWUbGS4ncI5GBhQ"
    baseURL   = "https://www.bitmex.com/api/v1"
)

// Execution 结构表示单笔交易记录
type Execution struct {
    ExecID   string  `json:"execID"`
    OrderID  string  `json:"orderID"`
    Symbol   string  `json:"symbol"`
    Price    float64 `json:"price"`
    OrderQty int     `json:"orderQty"`
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

func fetchData(endpoint string) ([]byte, error) {
    // 当前时间戳（Unix 时间，单位：秒）
    expires := fmt.Sprintf("%d", time.Now().Unix()+60) // 有效期 60 秒

    // 请求 URL
    url := baseURL + endpoint

    // 请求头
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("api-key", apiID)
    // 如果需要签名，取消注释以下代码（需提供 apiSecret）
    // signature := generateSignature("GET", endpoint, expires, apiSecret)
    // req.Header.Set("api-signature", signature)
    // req.Header.Set("api-expires", expires)

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

func main() {
    // 获取交易历史
    endpoint := "/execution"
    data, err := fetchData(endpoint)
    if err != nil {
        fmt.Printf("Error fetching executions: %v\n", err)
        return
    }

    // 解析交易数据
    var executions []Execution
    if err := json.Unmarshal(data, &executions); err != nil {
        fmt.Printf("Error unmarshaling executions: %v\n", err)
        return
    }
    fmt.Println("Recent Executions:")
    for i, exec := range executions[:5] {
        fmt.Printf("%d. ExecID: %s, OrderID: %s, Price: %.2f, Qty: %d\n", i+1, exec.ExecID, exec.OrderID, exec.Price, exec.OrderQty)
    }

    // 获取 K 线数据
    endpoint = "/trade/bucketed?symbol=XBTUSD&binSize=1d&count=100"
    data, err = fetchData(endpoint)
    if err != nil {
        fmt.Printf("Error fetching trade buckets: %v\n", err)
        return
    }

    // 解析 K 线数据
    var tradeBuckets []TradeBucket
    if err := json.Unmarshal(data, &tradeBuckets); err != nil {
        fmt.Printf("Error unmarshaling trade buckets: %v\n", err)
        return
    }
    fmt.Println("\nRecent Trade Buckets (Daily):")
    for i, bucket := range tradeBuckets[:5] {
        fmt.Printf("%d. Timestamp: %s, Open: %.2f, Close: %.2f, Volume: %d\n", i+1, bucket.Timestamp, bucket.Open, bucket.Close, bucket.Volume)
    }
}
