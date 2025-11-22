package indicators

import (
	"github.com/markcheno/go-talib"
)

// KlineData 代表单根K线的OHLCV数据
type KlineData struct {
	OpenTime  int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	CloseTime int64
}

// Indicators 包含计算后的技术指标
type Indicators struct {
	RSI14          float64 // RSI(14)
	MACD           float64 // MACD线
	MACDSignal     float64 // 信号线
	MACDHistogram  float64 // 柱状图
	MacdCrossUp    bool    // MACD金叉
	MacdCrossDown  bool    // MACD死叉
}

// KlineWithIndicators 带指标的K线数据
type KlineWithIndicators struct {
	KlineData
	Indicators
}

// CalculateIndicators 为K线数据计算技术指标
func CalculateIndicators(klines []KlineData) []KlineWithIndicators {
	n := len(klines)
	if n < 30 {
		// 数据不足，无法计算指标
		return nil
	}

	result := make([]KlineWithIndicators, n)

	// 提取收盘价
	closes := make([]float64, n)
	for i, k := range klines {
		result[i].KlineData = k
		closes[i] = k.Close
	}

	// 计算 RSI(14)
	if n >= 15 {
		rsi := talib.Rsi(closes, 14)
		for i := 0; i < n; i++ {
			result[i].RSI14 = rsi[i]
		}
	}

	// 计算 MACD(12, 26, 9)
	if n >= 27 {
		macd, signal, histogram := talib.Macd(closes, 12, 26, 9)
		for i := 0; i < n; i++ {
			result[i].MACD = macd[i]
			result[i].MACDSignal = signal[i]
			result[i].MACDHistogram = histogram[i]
		}

		// 检测MACD金叉和死叉
		for i := 1; i < n; i++ {
			prev := result[i-1]
			curr := result[i]

			// 金叉：MACD从下方穿过信号线
			if prev.MACD < prev.MACDSignal && curr.MACD > curr.MACDSignal {
				result[i].MacdCrossUp = true
			}

			// 死叉：MACD从上方穿过信号线
			if prev.MACD > prev.MACDSignal && curr.MACD < curr.MACDSignal {
				result[i].MacdCrossDown = true
			}
		}
	}

	return result
}

// GetPrevLowestPrice 获取前n根K线的最低价
func GetPrevLowestPrice(klines []KlineWithIndicators, currentIndex int, n int) float64 {
	if currentIndex < n {
		n = currentIndex
	}

	if n <= 0 {
		return 0
	}

	lowest := klines[currentIndex-n].Low
	for i := currentIndex - n + 1; i < currentIndex; i++ {
		if klines[i].Low < lowest {
			lowest = klines[i].Low
		}
	}
	return lowest
}

// GetPrevHighestPrice 获取前n根K线的最高价
func GetPrevHighestPrice(klines []KlineWithIndicators, currentIndex int, n int) float64 {
	if currentIndex < n {
		n = currentIndex
	}

	if n <= 0 {
		return 0
	}

	highest := klines[currentIndex-n].High
	for i := currentIndex - n + 1; i < currentIndex; i++ {
		if klines[i].High > highest {
			highest = klines[i].High
		}
	}
	return highest
}

// IsPrevRSILessThan 检查前n根K线中是否有RSI小于阈值
func IsPrevRSILessThan(klines []KlineWithIndicators, currentIndex int, threshold float64, n int) bool {
	if currentIndex < n {
		n = currentIndex
	}

	for i := currentIndex - n; i < currentIndex; i++ {
		if i >= 0 && klines[i].RSI14 < threshold {
			return true
		}
	}
	return false
}

// IsPrevRSIGreaterThan 检查前n根K线中是否有RSI大于阈值
func IsPrevRSIGreaterThan(klines []KlineWithIndicators, currentIndex int, threshold float64, n int) bool {
	if currentIndex < n {
		n = currentIndex
	}

	for i := currentIndex - n; i < currentIndex; i++ {
		if i >= 0 && klines[i].RSI14 > threshold {
			return true
		}
	}
	return false
}
