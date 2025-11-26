#!/bin/bash

# ============================================================
# BitMEX Web Server 测试和诊断脚本
# ============================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 服务器配置
SERVER_URL="http://localhost:8080"
REPORT_DIR="report"
REPORT_FILE="$REPORT_DIR/web_server_test_$(date +%Y%m%d_%H%M%S).txt"

# 计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# ============================================================
# 辅助函数
# ============================================================

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

log_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_header() {
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  $1"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
}

# 测试 HTTP 端点
test_endpoint() {
    local endpoint=$1
    local description=$2
    local expect_json=${3:-true}

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    log_info "测试: $description"
    log_info "URL: $SERVER_URL$endpoint"

    # 发送请求
    response=$(curl -s -w "\n%{http_code}" "$SERVER_URL$endpoint" 2>&1)
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    # 检查 HTTP 状态码
    if [ "$http_code" = "200" ]; then
        log_success "HTTP 200 OK"

        # 检查是否为 JSON
        if [ "$expect_json" = "true" ]; then
            if echo "$body" | jq . >/dev/null 2>&1; then
                log_success "响应格式: 有效的 JSON"

                # 显示数据统计
                if echo "$body" | jq -e 'type == "array"' >/dev/null 2>&1; then
                    count=$(echo "$body" | jq 'length')
                    log_info "  └─ 数组长度: $count 条记录"

                    # 显示第一条记录（如果存在）
                    if [ "$count" -gt 0 ]; then
                        first=$(echo "$body" | jq '.[0]')
                        log_info "  └─ 示例数据:"
                        echo "$first" | jq . | sed 's/^/       /'
                    fi
                elif echo "$body" | jq -e 'type == "object"' >/dev/null 2>&1; then
                    log_info "  └─ 对象数据:"
                    echo "$body" | jq . | sed 's/^/       /'
                fi
            else
                log_error "响应格式: 无效的 JSON"
                log_info "  └─ 响应内容: ${body:0:200}..."
            fi
        fi
    elif [ "$http_code" = "404" ]; then
        log_error "HTTP 404 Not Found - 端点不存在或数据未找到"
    else
        log_error "HTTP $http_code - 请求失败"
        log_info "  └─ 响应: ${body:0:200}..."
    fi

    echo ""
}

# 测试服务器是否运行
test_server_running() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    log_info "检查服务器是否运行..."

    if curl -s -f "$SERVER_URL" >/dev/null 2>&1; then
        log_success "服务器正在运行"
        return 0
    else
        log_error "服务器未运行或无法连接"
        return 1
    fi
}

# 测试静态文件
test_static_files() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    log_info "测试静态文件: index.html"

    response=$(curl -s -w "\n%{http_code}" "$SERVER_URL/" 2>&1)
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ]; then
        if echo "$body" | grep -q "BitMEX Trading Dashboard"; then
            log_success "index.html 加载成功"

            # 检查关键元素
            if echo "$body" | grep -q "lightweight-charts"; then
                log_success "  └─ 包含 TradingView 图表库"
            fi

            if echo "$body" | grep -q "chart"; then
                log_success "  └─ 包含图表元素"
            fi
        else
            log_error "index.html 内容异常"
        fi
    else
        log_error "无法加载 index.html (HTTP $http_code)"
    fi

    echo ""
}

# 性能测试
test_performance() {
    local endpoint=$1
    local description=$2

    log_info "性能测试: $description"

    # 进行 5 次请求，计算平均响应时间
    total_time=0
    for i in {1..5}; do
        start=$(date +%s%N)
        curl -s "$SERVER_URL$endpoint" >/dev/null 2>&1
        end=$(date +%s%N)
        duration=$((($end - $start) / 1000000)) # 转换为毫秒
        total_time=$(($total_time + $duration))
    done

    avg_time=$(($total_time / 5))

    if [ $avg_time -lt 1000 ]; then
        log_success "平均响应时间: ${avg_time}ms (优秀)"
    elif [ $avg_time -lt 3000 ]; then
        log_success "平均响应时间: ${avg_time}ms (良好)"
    else
        log_warning "平均响应时间: ${avg_time}ms (较慢)"
    fi

    echo ""
}

# 数据一致性测试
test_data_consistency() {
    log_info "测试数据一致性..."

    # 检查 K线数据时间戳是否递增
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    klines=$(curl -s "$SERVER_URL/api/klines?symbol=XBTUSD&timeframe=1d")

    if echo "$klines" | jq -e 'type == "array" and length > 0' >/dev/null 2>&1; then
        # 检查时间戳是否递增
        is_sorted=$(echo "$klines" | jq '[.[].time] | . == (. | sort)')

        if [ "$is_sorted" = "true" ]; then
            log_success "K线数据时间戳正确排序"
        else
            log_error "K线数据时间戳未正确排序"
        fi
    fi

    # 检查账户数据的逻辑性
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    account=$(curl -s "$SERVER_URL/api/account")

    if echo "$account" | jq -e '.balance' >/dev/null 2>&1; then
        balance=$(echo "$account" | jq '.balance')
        total_pnl=$(echo "$account" | jq '.totalPnl')
        win_rate=$(echo "$account" | jq '.winRate')

        # 检查胜率是否在 0-100 之间
        if (( $(echo "$win_rate >= 0 && $win_rate <= 100" | bc -l) )); then
            log_success "胜率数据合理: ${win_rate}%"
        else
            log_error "胜率数据异常: ${win_rate}%"
        fi
    fi

    echo ""
}

# ============================================================
# 主测试流程
# ============================================================

main() {
    # 创建报告目录
    mkdir -p "$REPORT_DIR"

    # 开始测试
    print_header "BitMEX Web Server 诊断测试"

    log_info "开始时间: $(date '+%Y-%m-%d %H:%M:%S')"
    log_info "服务器地址: $SERVER_URL"
    echo ""

    # 1. 服务器运行状态
    print_header "1. 服务器运行状态"
    if ! test_server_running; then
        log_error "服务器未运行，终止测试"
        exit 1
    fi
    echo ""

    # 2. 静态文件测试
    print_header "2. 静态文件测试"
    test_static_files

    # 3. API 端点测试
    print_header "3. API 端点测试"

    test_endpoint "/api/klines?symbol=XBTUSD&timeframe=1d" "K线数据 (XBTUSD 1d)"
    test_endpoint "/api/klines?symbol=ETHUSD&timeframe=1d" "K线数据 (ETHUSD 1d)"
    test_endpoint "/api/orders" "所有订单"
    test_endpoint "/api/orders/pending" "未成交订单"
    test_endpoint "/api/executions" "成交记录"
    test_endpoint "/api/positions" "当前仓位"
    test_endpoint "/api/account" "账户信息"

    # 4. 数据一致性测试
    print_header "4. 数据一致性测试"
    test_data_consistency

    # 5. 性能测试
    print_header "5. 性能测试"
    test_performance "/api/klines?symbol=XBTUSD&timeframe=1d" "K线数据查询"
    test_performance "/api/executions" "成交记录查询"

    # 6. CORS 测试
    print_header "6. CORS 配置测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    cors_header=$(curl -s -I "$SERVER_URL/api/account" | grep -i "access-control-allow-origin")
    if [ -n "$cors_header" ]; then
        log_success "CORS 已启用: $cors_header"
    else
        log_warning "CORS 未配置"
    fi
    echo ""

    # 7. 错误处理测试
    print_header "7. 错误处理测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    log_info "测试不存在的端点..."
    response=$(curl -s -w "\n%{http_code}" "$SERVER_URL/api/nonexistent" 2>&1)
    http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" = "404" ]; then
        log_success "正确返回 404 错误"
    else
        log_error "错误处理异常 (HTTP $http_code)"
    fi
    echo ""

    # ============================================================
    # 测试总结
    # ============================================================

    print_header "测试总结"

    echo "总测试数: $TOTAL_TESTS"
    echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
    echo -e "${RED}失败: $FAILED_TESTS${NC}"
    echo ""

    success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo "成功率: ${success_rate}%"
    echo ""

    if [ $FAILED_TESTS -eq 0 ]; then
        log_success "所有测试通过！"
    else
        log_warning "有 $FAILED_TESTS 个测试失败"
    fi

    echo ""
    log_info "结束时间: $(date '+%Y-%m-%d %H:%M:%S')"

    # ============================================================
    # 生成详细报告
    # ============================================================

    print_header "系统信息"

    log_info "数据文件状态:"

    if [ -f "wallet.csv" ]; then
        lines=$(wc -l < wallet.csv)
        size=$(ls -lh wallet.csv | awk '{print $5}')
        log_success "  wallet.csv: $lines 行, $size"
    else
        log_warning "  wallet.csv: 不存在"
    fi

    if [ -f "orders.csv" ]; then
        lines=$(wc -l < orders.csv)
        size=$(ls -lh orders.csv | awk '{print $5}')
        log_success "  orders.csv: $lines 行, $size"
    else
        log_warning "  orders.csv: 不存在"
    fi

    if [ -f "executions.csv" ]; then
        lines=$(wc -l < executions.csv)
        size=$(ls -lh executions.csv | awk '{print $5}')
        log_success "  executions.csv: $lines 行, $size"
    else
        log_warning "  executions.csv: 不存在"
    fi

    if [ -f "klines_XBTUSD_1d.csv" ]; then
        lines=$(wc -l < klines_XBTUSD_1d.csv)
        size=$(ls -lh klines_XBTUSD_1d.csv | awk '{print $5}')
        log_success "  klines_XBTUSD_1d.csv: $lines 行, $size"
    else
        log_warning "  klines_XBTUSD_1d.csv: 不存在"
    fi

    echo ""

    # ============================================================
    # 快速诊断建议
    # ============================================================

    if [ $FAILED_TESTS -gt 0 ]; then
        print_header "诊断建议"

        log_info "如果测试失败，请检查："
        echo "  1. 确认 web_server.go 正在运行"
        echo "  2. 检查数据文件是否存在且格式正确"
        echo "  3. 查看服务器日志输出"
        echo "  4. 确认端口 8080 未被占用"
        echo "  5. 检查防火墙设置"
        echo ""
    fi

    log_info "完整报告已保存至: $REPORT_FILE"
}

# 运行主函数并保存输出
main 2>&1 | tee "$REPORT_FILE"

# 返回状态码
if [ $FAILED_TESTS -eq 0 ]; then
    exit 0
else
    exit 1
fi
