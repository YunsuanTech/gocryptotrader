package token

import (
	"fmt"
)

// TradingSignal 定义交易信号类型
type TradingSignal string

const (
	// SignalBuy 买入信号
	SignalBuy TradingSignal = "BUY"
	// SignalSell 卖出信号
	SignalSell TradingSignal = "SELL"
	// SignalHold 持有/无信号
	SignalHold TradingSignal = "HOLD"
)

// AnalyzeKlineSignal 分析K线数据并返回交易信号
// 它获取指定代币的K线数据，并分析最新的三组数据以确定趋势。
// 如果中间一组数据的最高价和最低价都是三组中最高的，则返回卖出信号。
// 如果中间一组数据的最高价和最低价都是三组中最低的，则返回买入信号。
// 否则返回持有信号。
func AnalyzeKlineSignal(tokenAddress string) (TradingSignal, error) {
	fetcher, err := NewGMGNFetcher()
	if err != nil {
		return SignalHold, fmt.Errorf("创建 GMGNFetcher 失败: %w", err)
	}

	klineData, err := fetcher.GetTokenKlineData(tokenAddress)
	if err != nil {
		return SignalHold, fmt.Errorf("获取 K 线数据失败 for token %s: %w", tokenAddress, err)
	}

	if len(klineData) < 3 {
		return SignalHold, fmt.Errorf("K 线数据不足三组 for token %s, 无法分析", tokenAddress)
	}

	// 只取最新的完整三组数据进行分析
	k1 := klineData[len(klineData)-4]
	k2 := klineData[len(klineData)-3]
	k3 := klineData[len(klineData)-2]

	fmt.Printf("[AnalyzeKlineSignal] token: %s\nK1: high=%v, low=%v, open=%v, close=%v, time=%v\nK2: high=%v, low=%v, open=%v, close=%v, time=%v\nK3: high=%v, low=%v, open=%v, close=%v, time=%v\n", tokenAddress, k1.High, k1.Low, k1.Open, k1.Close, k1.Time, k2.High, k2.Low, k2.Open, k2.Close, k2.Time, k3.High, k3.Low, k3.Open, k3.Close, k3.Time)

	isStrictlyHighestHigh := k2.High >= k1.High && k2.High >= k3.High
	isStrictlyHighestLow := k2.Low >= k1.Low && k2.Low >= k3.Low

	if isStrictlyHighestHigh && isStrictlyHighestLow && (k2.High > k1.High || k2.High > k3.High || k2.Low > k1.Low || k2.Low > k3.Low) {
		fmt.Printf("[AnalyzeKlineSignal] 判断结果: 卖出 (SELL)\n")
		return SignalSell, nil
	}

	isStrictlyLowestHigh := k2.High <= k1.High && k2.High <= k3.High
	isStrictlyLowestLow := k2.Low <= k1.Low && k2.Low <= k3.Low

	if isStrictlyLowestHigh && isStrictlyLowestLow && (k2.High < k1.High || k2.High < k3.High || k2.Low < k1.Low || k2.Low < k3.Low) {
		fmt.Printf("[AnalyzeKlineSignal] 判断结果: 买入 (BUY)\n")
		return SignalBuy, nil
	}

	fmt.Printf("[AnalyzeKlineSignal] 判断结果: 持有 (HOLD)\n")
	return SignalHold, nil
}
