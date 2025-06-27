package handling

import (
	"fmt"
	"log"
	"time"
)

// TokenPriceProvider 定义获取代币价格的接口
type TokenPriceProvider interface {
	GetTokenPrice(symbol string) (interface{}, error)
}

// ExchangeType 表示交易所类型
type ExchangeType string

// 支持的交易所类型
const (
	BinanceExchange ExchangeType = "binance"
	GMGNExchange    ExchangeType = "gmgn"
)

// MonitorPrice 监控指定交易对的价格
// symbol: 交易对，例如 "btcusdt"
// exchangeType: 交易所类型，例如 "binance" 或 "gmgn"
// timeout: 监控持续时间，如果为 0 则一直监控直到出错
func MonitorPrice(symbol string, exchangeType ExchangeType, timeout time.Duration) error {
	log.Printf("开始监控 %s 交易所的 %s 价格", exchangeType, symbol)

	switch exchangeType {
	case BinanceExchange:
		// 使用币安价格监控
		return MonitorBinancePrice(symbol, nil, timeout)
	case GMGNExchange:
		// GMGN 不支持实时监控，只能获取当前价格
		price, err := GetGMGNTokenPrice(symbol, symbol)
		if err != nil {
			return err
		}
		log.Printf("GMGN 交易所 %s 价格: USD %.6f, SOL %.6f",
			symbol, price.USDPrice, price.SOLPrice)
		return nil
	default:
		return fmt.Errorf("不支持的交易所类型: %s", exchangeType)
	}
}

// GetTokenPrice 获取指定交易对的价格
// symbol: 交易对，例如 "btcusdt"
// exchangeType: 交易所类型，例如 "binance" 或 "gmgn"
func GetTokenPrice(symbol string, exchangeType ExchangeType) (interface{}, error) {
	switch exchangeType {
	case BinanceExchange:
		return GetBinanceTokenPrice(symbol)
	case GMGNExchange:
		return GetGMGNTokenPrice(symbol, symbol)
	default:
		return nil, fmt.Errorf("不支持的交易所类型: %s", exchangeType)
	}
}
