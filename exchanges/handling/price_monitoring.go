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
	BinanceExchange     ExchangeType = "binance"
	GMGNExchange        ExchangeType = "gmgn"
	UniswapExchange     ExchangeType = "uniswap"
	PancakeswapExchange ExchangeType = "pancakeswap"
	CoinGeckoExchange   ExchangeType = "coingecko"
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

// TokenPriceInfo 表示不同交易所的代币价格信息
type TokenPriceInfo struct {
	Symbol         string                 `json:"symbol"`
	Timestamp      time.Time              `json:"timestamp"`
	ExchangePrices map[string]interface{} `json:"exchange_prices"`
	USDPrices      map[string]float64     `json:"usd_prices"`
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
	case UniswapExchange:
		return GetUniswapTokenPrice(symbol, symbol)
	case PancakeswapExchange:
		// 对于PancakeSwap，我们需要代币地址，这里简化处理，使用symbol作为地址
		return GetPancakeTokenPrice(symbol, symbol)
	case CoinGeckoExchange:
		// 对于CoinGecko，我们需要CoinGecko ID，这里简化处理，使用symbol作为ID
		return GetCoinGeckoTokenPrice(symbol, symbol, "")
	default:
		return nil, fmt.Errorf("不支持的交易所类型: %s", exchangeType)
	}
}

// GetMultiExchangeTokenPrices 获取所有交易所的代币价格信息
// symbol: 代币符号，例如 "ETH"
// tokenAddress: 代币地址，用于Uniswap和PancakeSwap
// coinGeckoID: CoinGecko ID，用于CoinGecko
func GetMultiExchangeTokenPrices(symbol, tokenAddress, coinGeckoID string) (*TokenPriceInfo, error) {
	if symbol == "" {
		return nil, fmt.Errorf("代币符号不能为空")
	}

	// 初始化结果
	result := &TokenPriceInfo{
		Symbol:         symbol,
		Timestamp:      time.Now(),
		ExchangePrices: make(map[string]interface{}),
		USDPrices:      make(map[string]float64),
	}

	// 定义要查询的交易所
	exchanges := []ExchangeType{
		BinanceExchange,
		GMGNExchange,
		UniswapExchange,
		PancakeswapExchange,
		CoinGeckoExchange,
	}

	// 存储成功获取价格的交易所数量
	var successCount int

	// 查询每个交易所的价格
	for _, exchange := range exchanges {
		var price interface{}
		var err error
		var usdPrice float64

		// 根据交易所类型获取价格
		switch exchange {
		case BinanceExchange:
			price, err = GetBinanceTokenPrice(symbol)
			if err == nil {
				if p, ok := price.(*BinanceTokenPrice); ok {
					usdPrice = p.LastPrice
				}
			}
		case GMGNExchange:
			price, err = GetGMGNTokenPrice(tokenAddress, symbol)
			if err == nil {
				if p, ok := price.(*GMGNTokenPrice); ok {
					usdPrice = p.USDPrice
				}
			}
		case UniswapExchange:
			price, err = GetUniswapTokenPrice(tokenAddress, symbol)
			if err == nil {
				if p, ok := price.(*UniswapTokenPrice); ok {
					usdPrice = p.USDPrice
				}
			}
		case PancakeswapExchange:
			price, err = GetPancakeTokenPrice(tokenAddress, symbol)
			if err == nil {
				if p, ok := price.(*PancakeTokenPrice); ok {
					usdPrice = p.USDPrice
				}
			}
		case CoinGeckoExchange:
			// 如果没有提供CoinGecko ID，则使用symbol
			id := coinGeckoID
			if id == "" {
				id = symbol
			}
			price, err = GetCoinGeckoTokenPrice(id, symbol, "")
			if err == nil {
				if p, ok := price.(*CoinGeckoTokenPrice); ok {
					usdPrice = p.USDPrice
				}
			}
		}

		// 记录价格和错误
		if err != nil {
			log.Printf("获取 %s 在 %s 交易所的价格失败: %v", symbol, exchange, err)
			continue
		}

		// 保存价格数据
		result.ExchangePrices[string(exchange)] = price
		result.USDPrices[string(exchange)] = usdPrice
		successCount++
	}

	// 如果没有获取到任何价格，返回错误
	if successCount == 0 {
		return nil, fmt.Errorf("未能从任何交易所获取到 %s 的价格", symbol)
	}

	log.Printf("成功从 %d 个交易所获取到 %s 的价格信息", successCount, symbol)
	return result, nil
}

// CompareTokenPrices 为了保持向后兼容性而保留的函数，内部调用GetMultiExchangeTokenPrices
// 已弃用：请使用 GetMultiExchangeTokenPrices 代替
func CompareTokenPrices(symbol, tokenAddress, coinGeckoID string) (*TokenPriceInfo, error) {
	return GetMultiExchangeTokenPrices(symbol, tokenAddress, coinGeckoID)
}
