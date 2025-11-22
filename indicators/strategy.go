package indicators

import (
	"fmt"
	"math"
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

// ========== 背离检测 ==========

// DivergenceType 背离类型
type DivergenceType string

const (
	DivergenceBullish DivergenceType = "看涨" // 看涨背离（价格↓ 指标↑）
	DivergenceBearish DivergenceType = "看跌" // 看跌背离（价格↑ 指标↓）
)

// DivergenceSignal 背离信号
type DivergenceSignal struct {
	Type               DivergenceType  // 背离类型
	FirstSignal        *TradingSignal  // 第一个信号
	SecondSignal       *TradingSignal  // 第二个信号（触发背离的信号）
	PriceChange        float64         // 价格变化
	PriceChangePercent float64         // 价格变化百分比
	RSIChange          float64         // RSI变化
	MACDChange         float64         // MACD变化
	TimeGapMinutes     int             // 时间间隔（分钟）
}

// String 格式化输出背离信号
func (d *DivergenceSignal) String() string {
	return fmt.Sprintf("[%s背离] %s | 间隔: %d分钟 | 价格: %.2f→%.2f (%+.2f%%) | RSI: %.2f→%.2f (%+.2f) | MACD: %.4f→%.4f (%+.4f)",
		d.Type,
		d.SecondSignal.Time.In(BeijingLocation).Format("2006-01-02 15:04:05"),
		d.TimeGapMinutes,
		d.FirstSignal.Price,
		d.SecondSignal.Price,
		d.PriceChangePercent,
		d.FirstSignal.RSI14,
		d.SecondSignal.RSI14,
		d.RSIChange,
		d.FirstSignal.MACD,
		d.SecondSignal.MACD,
		d.MACDChange,
	)
}

// DetectDivergence 检测背离信号
// 背离是指价格走势与技术指标走势相反的现象，是强烈的反转信号
//
// 看涨背离（Bullish Divergence）：
//   - 价格创新低（下跌）
//   - 但RSI和MACD未创新低（反而上涨）
//   - 说明：虽然价格在跌，但动能在增强 → 强烈买入信号
//
// 看跌背离（Bearish Divergence）：
//   - 价格创新高（上涨）
//   - 但RSI和MACD未创新高（反而下跌）
//   - 说明：虽然价格在涨，但动能在减弱 → 强烈卖出信号
func DetectDivergence(signals []*TradingSignal) []*DivergenceSignal {
	// 配置参数
	const (
		maxTimeGapMinutes     = 30   // 最大时间间隔（分钟）
		minPriceChangePercent = 0.1  // 最小价格变化百分比
	)

	var divergences []*DivergenceSignal

	// 遍历所有信号，寻找同类型的连续信号
	for i := 1; i < len(signals); i++ {
		curr := signals[i]

		// 向前查找相同类型的信号
		for j := i - 1; j >= 0; j-- {
			prev := signals[j]

			// 只比较相同类型的信号
			if curr.Type != prev.Type {
				continue
			}

			// 检查时间间隔
			timeGap := curr.Time.Sub(prev.Time).Minutes()
			if timeGap > maxTimeGapMinutes {
				break // 时间间隔太大，停止向前查找
			}

			// 计算价格变化
			priceChange := curr.Price - prev.Price
			priceChangePercent := (priceChange / prev.Price) * 100

			// 检查价格变化幅度
			if math.Abs(priceChangePercent) < minPriceChangePercent {
				continue
			}

			// 计算指标变化
			rsiChange := curr.RSI14 - prev.RSI14
			macdChange := curr.MACD - prev.MACD

			// 检测看涨背离（LONG信号）
			// 条件：价格下跌 且 RSI上涨 且 MACD上涨
			if curr.Type == SignalLong {
				if priceChange < 0 && rsiChange > 0 && macdChange > 0 {
					divergences = append(divergences, &DivergenceSignal{
						Type:               DivergenceBullish,
						FirstSignal:        prev,
						SecondSignal:       curr,
						PriceChange:        priceChange,
						PriceChangePercent: priceChangePercent,
						RSIChange:          rsiChange,
						MACDChange:         macdChange,
						TimeGapMinutes:     int(timeGap),
					})
					break // 找到背离后，不再向前查找
				}
			}

			// 检测看跌背离（SHORT信号）
			// 条件：价格上涨 且 RSI下跌 且 MACD下跌
			if curr.Type == SignalShort {
				if priceChange > 0 && rsiChange < 0 && macdChange < 0 {
					divergences = append(divergences, &DivergenceSignal{
						Type:               DivergenceBearish,
						FirstSignal:        prev,
						SecondSignal:       curr,
						PriceChange:        priceChange,
						PriceChangePercent: priceChangePercent,
						RSIChange:          rsiChange,
						MACDChange:         macdChange,
						TimeGapMinutes:     int(timeGap),
					})
					break // 找到背离后，不再向前查找
				}
			}
		}
	}

	return divergences
}
