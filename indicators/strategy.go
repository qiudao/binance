package indicators

import (
	"fmt"
	"time"
)

// 北京时间时区
var BeijingLocation = time.FixedZone("CST", 8*3600)

// SignalType 信号类型
type SignalType string

const (
	SignalLong  SignalType = "LONG"  // 做多信号
	SignalShort SignalType = "SHORT" // 做空信号
)

// TradingSignal 交易信号
type TradingSignal struct {
	Type          SignalType // 信号类型
	Time          time.Time  // 信号时间
	Price         float64    // 入场价格
	StopLoss      float64    // 止损价格
	RSI14         float64    // 当前RSI值
	MACD          float64    // 当前MACD值
	MACDSignal    float64    // 当前信号线值
	RiskAmount    float64    // 风险金额（入场价 - 止损价）
	RiskPercent   float64    // 风险百分比
}

// String 格式化输出信号
func (s *TradingSignal) String() string {
	return fmt.Sprintf("[%s] %s | 价格: %.2f | 止损: %.2f | 风险: %.2f (%.2f%%) | RSI: %.2f | MACD: %.4f",
		s.Type,
		s.Time.In(BeijingLocation).Format("2006-01-02 15:04:05"),
		s.Price,
		s.StopLoss,
		s.RiskAmount,
		s.RiskPercent,
		s.RSI14,
		s.MACD,
	)
}

// CheckLongSignal 检测做多信号
// 条件：
// 1. MACD金叉
// 2. 前10根K线中RSI < 30
// 3. 前10根K线最低价 < 当前价（作为止损价）
func CheckLongSignal(klines []KlineWithIndicators, index int) *TradingSignal {
	if index < 10 {
		// 数据不足
		return nil
	}

	current := klines[index]

	// 条件1: MACD金叉
	if !current.MacdCrossUp {
		return nil
	}

	// 条件2: 前10根K线中RSI < 30
	if !IsPrevRSILessThan(klines, index, 30, 10) {
		return nil
	}

	// 条件3: 前10根K线最低价作为止损
	stopLoss := GetPrevLowestPrice(klines, index, 10)
	if stopLoss >= current.Close {
		// 止损价不能高于或等于入场价
		return nil
	}

	// 计算风险
	riskAmount := current.Close - stopLoss
	riskPercent := (riskAmount / current.Close) * 100

	return &TradingSignal{
		Type:         SignalLong,
		Time:         time.UnixMilli(current.CloseTime),
		Price:        current.Close,
		StopLoss:     stopLoss,
		RSI14:        current.RSI14,
		MACD:         current.MACD,
		MACDSignal:   current.MACDSignal,
		RiskAmount:   riskAmount,
		RiskPercent:  riskPercent,
	}
}

// CheckShortSignal 检测做空信号
// 条件：
// 1. MACD死叉
// 2. 前10根K线中RSI > 70
// 3. 前10根K线最高价 > 当前价（作为止损价）
func CheckShortSignal(klines []KlineWithIndicators, index int) *TradingSignal {
	if index < 10 {
		// 数据不足
		return nil
	}

	current := klines[index]

	// 条件1: MACD死叉
	if !current.MacdCrossDown {
		return nil
	}

	// 条件2: 前10根K线中RSI > 70
	if !IsPrevRSIGreaterThan(klines, index, 70, 10) {
		return nil
	}

	// 条件3: 前10根K线最高价作为止损
	stopLoss := GetPrevHighestPrice(klines, index, 10)
	if stopLoss <= current.Close {
		// 止损价不能低于或等于入场价
		return nil
	}

	// 计算风险
	riskAmount := stopLoss - current.Close
	riskPercent := (riskAmount / current.Close) * 100

	return &TradingSignal{
		Type:         SignalShort,
		Time:         time.UnixMilli(current.CloseTime),
		Price:        current.Close,
		StopLoss:     stopLoss,
		RSI14:        current.RSI14,
		MACD:         current.MACD,
		MACDSignal:   current.MACDSignal,
		RiskAmount:   riskAmount,
		RiskPercent:  riskPercent,
	}
}

// ScanSignals 扫描所有K线，检测交易信号
func ScanSignals(klines []KlineWithIndicators) []*TradingSignal {
	var signals []*TradingSignal

	// 从第50根开始扫描（确保有足够的历史数据计算指标）
	startIndex := 50
	if len(klines) < startIndex {
		return signals
	}

	for i := startIndex; i < len(klines); i++ {
		// 检查做多信号
		if signal := CheckLongSignal(klines, i); signal != nil {
			signals = append(signals, signal)
		}

		// 检查做空信号
		if signal := CheckShortSignal(klines, i); signal != nil {
			signals = append(signals, signal)
		}
	}

	return signals
}
