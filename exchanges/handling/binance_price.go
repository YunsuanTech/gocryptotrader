package handling

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"gocryptotrader/database/models/sqlite3"
	currency "gocryptotrader/database/repository/currency"
	marketdata "gocryptotrader/database/repository/market_data"

	"github.com/gorilla/websocket"
)

// Binance API URL constants
const (
	BinanceWSBaseURL  = "wss://stream.binance.com"
	BinanceTickerPath = "/ws/%s@ticker"
)

// 自定义类型处理数字/字符串混合值
type FlexString string

func (f *FlexString) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		// 如果是字符串类型
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*f = FlexString(s)
	} else {
		// 如果是数字类型
		var n float64
		if err := json.Unmarshal(b, &n); err != nil {
			return err
		}
		*f = FlexString(strconv.FormatFloat(n, 'f', -1, 64))
	}
	return nil
}

// BinanceTicker 定义结构体，匹配API返回的所有字段
type BinanceTicker struct {
	EventType          string     `json:"e"` // 事件类型
	EventTime          int64      `json:"E"` // 事件时间(毫秒)
	Symbol             string     `json:"s"` // 交易对
	PriceChange        FlexString `json:"p"` // 价格变化
	PriceChangePercent FlexString `json:"P"` // 价格变化百分比
	WeightedAvgPrice   FlexString `json:"w"` // 加权平均价
	PrevClosePrice     FlexString `json:"x"` // 前收盘价
	LastPrice          FlexString `json:"c"` // 最新价
	LastQuantity       FlexString `json:"Q"` // 最新成交量
	BidPrice           FlexString `json:"b"` // 买一价
	BidQuantity        FlexString `json:"B"` // 买一量
	AskPrice           FlexString `json:"a"` // 卖一价
	AskQuantity        FlexString `json:"A"` // 卖一量
	OpenPrice          FlexString `json:"o"` // 开盘价
	HighPrice          FlexString `json:"h"` // 最高价
	LowPrice           FlexString `json:"l"` // 最低价
	Volume             FlexString `json:"v"` // 成交量
	QuoteVolume        FlexString `json:"q"` // 成交额
	OpenTime           int64      `json:"O"` // 统计开始时间(毫秒)
	CloseTime          int64      `json:"C"` // 统计结束时间(毫秒)
	FirstTradeID       int64      `json:"F"` // 首笔成交ID
	LastTradeID        int64      `json:"L"` // 末笔成交ID
	TradeCount         int64      `json:"n"` // 成交笔数
}

// BinanceTickerHandler 定义处理行情数据的回调函数类型
type BinanceTickerHandler func(ticker BinanceTicker)

// BinanceTokenPrice 表示币安代币价格信息
type BinanceTokenPrice struct {
	Symbol             string    `json:"symbol"`
	LastPrice          float64   `json:"last_price"`
	BidPrice           float64   `json:"bid_price"`
	AskPrice           float64   `json:"ask_price"`
	PriceChange        float64   `json:"price_change"`
	PriceChangePercent float64   `json:"price_change_percent"`
	HighPrice          float64   `json:"high_price"`
	LowPrice           float64   `json:"low_price"`
	Volume             float64   `json:"volume"`
	LastUpdate         time.Time `json:"last_update"`
}

// DefaultBinanceTickerHandler 默认的行情数据处理函数
func DefaultBinanceTickerHandler(ticker BinanceTicker) {
	// 只转换和显示价格相关字段
	lastPrice, _ := strconv.ParseFloat(string(ticker.LastPrice), 64)
	openPrice, _ := strconv.ParseFloat(string(ticker.OpenPrice), 64)
	highPrice, _ := strconv.ParseFloat(string(ticker.HighPrice), 64)
	lowPrice, _ := strconv.ParseFloat(string(ticker.LowPrice), 64)
	priceChange, _ := strconv.ParseFloat(string(ticker.PriceChange), 64)
	priceChangePercent, _ := strconv.ParseFloat(string(ticker.PriceChangePercent), 64)

	log.Printf("【%s】最新价: %.2f | 24h变化: %.2f (%.2f%%) | 开盘价: %.2f | 最高价: %.2f | 最低价: %.2f",
		ticker.Symbol,
		lastPrice,
		priceChange,
		priceChangePercent,
		openPrice,
		highPrice,
		lowPrice)
}

// GetBinanceTokenPrice 获取指定交易对的最新价格
func GetBinanceTokenPrice(symbol string) (*BinanceTokenPrice, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	// 创建一个通道用于接收价格数据
	priceChan := make(chan *BinanceTokenPrice, 1)

	// 创建一个自定义的回调函数，将价格数据发送到通道
	callback := func(ticker BinanceTicker) {
		lastPrice, _ := strconv.ParseFloat(string(ticker.LastPrice), 64)
		bidPrice, _ := strconv.ParseFloat(string(ticker.BidPrice), 64)
		askPrice, _ := strconv.ParseFloat(string(ticker.AskPrice), 64)
		highPrice, _ := strconv.ParseFloat(string(ticker.HighPrice), 64)
		lowPrice, _ := strconv.ParseFloat(string(ticker.LowPrice), 64)
		volume, _ := strconv.ParseFloat(string(ticker.Volume), 64)
		priceChange, _ := strconv.ParseFloat(string(ticker.PriceChange), 64)
		priceChangePercent, _ := strconv.ParseFloat(string(ticker.PriceChangePercent), 64)

		priceChan <- &BinanceTokenPrice{
			Symbol:             ticker.Symbol,
			LastPrice:          lastPrice,
			BidPrice:           bidPrice,
			AskPrice:           askPrice,
			PriceChange:        priceChange,
			PriceChangePercent: priceChangePercent,
			HighPrice:          highPrice,
			LowPrice:           lowPrice,
			Volume:             volume,
			LastUpdate:         time.Now(),
		}
	}

	// 启动一个goroutine来监控价格
	errChan := make(chan error, 1)
	go func() {
		errChan <- MonitorBinancePrice(symbol, callback, 5*time.Second)
	}()

	// 等待价格数据或错误
	select {
	case price := <-priceChan:
		return price, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for price data")
	}
}

// MonitorBinancePrice 监控指定交易对的价格
// symbol: 交易对，例如 "btcusdt"
// callback: 回调函数，用于处理接收到的行情数据，如果为 nil 则使用默认处理函数
// timeout: 监控持续时间，如果为 0 则一直监控直到出错
func MonitorBinancePrice(symbol string, callback BinanceTickerHandler, timeout time.Duration) error {
	// 设置默认值
	if callback == nil {
		callback = DefaultBinanceTickerHandler
	}

	// 构建WebSocket URL
	u := url.URL{
		Scheme: "wss",
		Host:   "stream.binance.com",
		Path:   fmt.Sprintf("/ws/%s@ticker", symbol),
	}
	log.Printf("连接到: %s", u.String())

	// 建立WebSocket连接
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	defer c.Close()

	// 设置超时
	var timer *time.Timer
	var done chan struct{}
	if timeout > 0 {
		done = make(chan struct{})
		timer = time.AfterFunc(timeout, func() {
			log.Printf("监控时间到达 %s，关闭连接", timeout)
			c.Close()
			close(done)
		})
		defer timer.Stop()
	}

	// 处理接收到的消息
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("读取错误:", err)
			break
		}

		var ticker BinanceTicker
		if err := json.Unmarshal(message, &ticker); err != nil {
			log.Printf("JSON 解析错误: %v", err)
			log.Printf("原始数据: %s", message)
			continue
		}

		// 调用回调函数处理行情数据
		callback(ticker)

		// 如果配置了保存到数据库，则将数据保存到数据库
		SaveBinanceTickerToDb(ticker, "binance")

		// 检查是否已超时
		if timeout > 0 {
			select {
			case <-done:
				return nil
			default:
			}
		}
	}

	return nil
}

// SaveBinanceTickerToDb 将行情数据保存到数据库
func SaveBinanceTickerToDb(ticker BinanceTicker, exchangeName string) {
	// 转换价格数据
	lastPrice, _ := strconv.ParseFloat(string(ticker.LastPrice), 64)
	bidPrice, _ := strconv.ParseFloat(string(ticker.BidPrice), 64)
	askPrice, _ := strconv.ParseFloat(string(ticker.AskPrice), 64)
	volume24h, _ := strconv.ParseFloat(string(ticker.Volume), 64)
	highPrice, _ := strconv.ParseFloat(string(ticker.HighPrice), 64)
	lowPrice, _ := strconv.ParseFloat(string(ticker.LowPrice), 64)
	openPrice, _ := strconv.ParseFloat(string(ticker.OpenPrice), 64)

	// 将原始数据转换为JSON字符串
	rawData, _ := json.Marshal(ticker)

	// 创建上下文
	ctx := context.Background()

	// 先保存或更新货币信息
	// 尝试从数据库中获取货币信息
	currencyRepo := "gocryptotrader/database/repository/currency"
	_ = currencyRepo // 避免未使用的导入警告

	currenc, err := currency.GetCurrencyBySymbol(ctx, ticker.Symbol)
	if err != nil {
		// 货币不存在，创建新的货币记录
		now := time.Now()
		newCurrency := &sqlite3.Currency{
			Symbol:          ticker.Symbol,
			Name:            ticker.Symbol, // 使用Symbol作为Name，后续可以更新
			Decimals:        8,             // 默认小数位数，可以根据实际情况调整
			ContractAddress: "",            // Binance API没有提供合约地址
			Chain:           "",            // Binance API没有提供链信息
			IsActive:        true,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		// 插入新的货币记录
		err = currency.InsertCurrency(ctx, newCurrency)
		if err != nil {
			log.Printf("保存货币信息到数据库失败: %v", err)
		} else {
			log.Printf("已创建新货币 %s 到数据库，ID: %d", newCurrency.Symbol, newCurrency.ID)
			currenc = newCurrency
		}
	} else {
		// 货币存在，更新货币信息
		currenc.UpdatedAt = time.Now()
		// 如果有需要更新的字段，可以在这里设置

		err = currency.UpdateCurrency(ctx, currenc)
		if err != nil {
			log.Printf("更新货币信息失败: %v", err)
		} else {
			log.Printf("已更新货币 %s 信息", currenc.Symbol)
		}
	}

	// 获取交易对ID
	tradingPairID := 1 // 默认值
	if currenc != nil && currenc.ID > 0 {
		tradingPairID = currenc.ID
	}

	// 创建 MarketData 对象
	marketData := &sqlite3.MarketData{
		ExchangeID:    1,                    // 假设 Binance 的 exchange_id 为 1，实际应该从配置或数据库中获取
		TradingPairID: int64(tradingPairID), // 使用货币ID作为交易对ID，转换为int64类型
		Timestamp:     time.Now(),
		LastPrice:     lastPrice,
		BidPrice:      bidPrice,
		AskPrice:      askPrice,
		Volume24h:     volume24h,
		High24h:       highPrice,
		Low24h:        lowPrice,
		OpenPrice24h:  openPrice,
		ClosePrice24h: lastPrice, // 使用最新价格作为收盘价
		LiquidityUSD:  0,         // 这个字段在 Binance Ticker 中没有对应值
		SlippageBPS:   0,         // 这个字段在 Binance Ticker 中没有对应值
		SourceDataRaw: string(rawData),
	}

	// 保存市场数据到数据库
	err = marketdata.InsertMarketData(ctx, marketData)
	if err != nil {
		log.Printf("保存市场数据到数据库失败: %v", err)
	} else {
		log.Printf("已保存 %s 市场数据到数据库，ID: %d", ticker.Symbol, marketData.ID)
	}
}
