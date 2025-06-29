package handling

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"gocryptotrader/database/models/sqlite3"
	currency "gocryptotrader/database/repository/currency"
	marketdata "gocryptotrader/database/repository/market_data"
)

// CoinGecko API URL constants
// CoinGecko specific constants (using common constants from constants.go)
// Note: Common constants like CoinGeckoBaseURL, PriceEndpoint, DefaultTimeout are now defined in constants.go

// CoinGeckoToken 表示CoinGecko平台上的代币
type CoinGeckoToken struct {
	Symbol      string
	CoinGeckoID string // CoinGecko平台上的代币ID
	Chain       string // 代币所在的区块链
}

// CoinGeckoPriceResponse 表示CoinGecko价格API的响应
type CoinGeckoPriceResponse map[string]map[string]float64

// CoinGeckoTokenPrice 表示代币价格信息
type CoinGeckoTokenPrice struct {
	Symbol      string    `json:"symbol"`
	CoinGeckoID string    `json:"coingecko_id"`
	USDPrice    float64   `json:"usd_price"`
	LastUpdate  time.Time `json:"last_update"`
}

// CoinGeckoTokenConfig 表示要监控的CoinGecko代币配置
type CoinGeckoTokenConfig struct {
	Symbol      string
	CoinGeckoID string
	Chain       string
}

// GetCoinGeckoTokenPrice 获取单个代币的价格
func GetCoinGeckoTokenPrice(coinGeckoID string, symbol string, chain string) (*CoinGeckoTokenPrice, error) {
	if coinGeckoID == "" {
		return nil, fmt.Errorf("CoinGecko ID不能为空")
	}

	// 创建代币列表
	tokens := []CoinGeckoToken{
		{Symbol: symbol, CoinGeckoID: coinGeckoID, Chain: chain},
	}

	// 获取价格
	priceData, err := getTokenPrices(tokens)
	if err != nil {
		return nil, err
	}

	// 检查是否获取到价格
	price, exists := priceData[coinGeckoID]
	if !exists {
		return nil, fmt.Errorf("无法获取代币 %s 的价格", symbol)
	}

	usdPrice, exists := price["usd"]
	if !exists {
		return nil, fmt.Errorf("无法获取代币 %s 的USD价格", symbol)
	}

	// 保存到数据库
	SaveCoinGeckoPriceToDb(coinGeckoID, symbol, chain, usdPrice)

	return &CoinGeckoTokenPrice{
		Symbol:      symbol,
		CoinGeckoID: coinGeckoID,
		USDPrice:    usdPrice,
		LastUpdate:  time.Now(),
	}, nil
}

// GetMultipleTokenPrices 获取多个代币的价格
func GetMultipleTokenPrices(tokens []CoinGeckoToken) (map[string]*CoinGeckoTokenPrice, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("代币列表不能为空")
	}

	// 获取价格
	priceData, err := getTokenPrices(tokens)
	if err != nil {
		return nil, err
	}

	// 处理结果
	result := make(map[string]*CoinGeckoTokenPrice)
	for _, token := range tokens {
		if price, exists := priceData[token.CoinGeckoID]; exists {
			if usdPrice, exists := price["usd"]; exists {
				// 保存到数据库
				SaveCoinGeckoPriceToDb(token.CoinGeckoID, token.Symbol, token.Chain, usdPrice)

				// 添加到结果
				result[token.Symbol] = &CoinGeckoTokenPrice{
					Symbol:      token.Symbol,
					CoinGeckoID: token.CoinGeckoID,
					USDPrice:    usdPrice,
					LastUpdate:  time.Now(),
				}
			}
		}
	}

	return result, nil
}

// getTokenPrices 获取多个代币价格
func getTokenPrices(tokens []CoinGeckoToken) (CoinGeckoPriceResponse, error) {
	// 构建代币ID列表
	var ids []string
	for _, token := range tokens {
		ids = append(ids, token.CoinGeckoID)
	}

	// 创建API URL
	url := fmt.Sprintf("%s%s?ids=%s&vs_currencies=usd",
		CoinGeckoBaseURL, CommonPriceEndpoint, joinIDs(ids))

	// 创建HTTP客户端
	client := &http.Client{Timeout: CommonDefaultTimeout}

	// 发送请求
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误 %d: %s", resp.StatusCode, resp.Status)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析JSON
	var priceData CoinGeckoPriceResponse
	if err := json.Unmarshal(body, &priceData); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	return priceData, nil
}

// joinIDs 辅助函数：连接ID列表
func joinIDs(ids []string) string {
	if len(ids) == 0 {
		return ""
	}

	return strings.Join(ids, ",")
}

// MonitorCoinGeckoPrice 监控CoinGecko上指定代币的价格
func MonitorCoinGeckoPrice(coinGeckoID string, symbol string, chain string, interval time.Duration, stopCh <-chan struct{}) {
	log.Printf("开始监控 CoinGecko 上 %s 的价格，间隔: %v", symbol, interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 立即获取一次价格
	go func() {
		price, err := GetCoinGeckoTokenPrice(coinGeckoID, symbol, chain)
		if err != nil {
			log.Printf("获取 %s 初始价格失败: %v", symbol, err)
		} else {
			log.Printf("%s 初始价格: $%.4f", symbol, price.USDPrice)
		}
	}()

	for {
		select {
		case <-ticker.C:
			go func() {
				price, err := GetCoinGeckoTokenPrice(coinGeckoID, symbol, chain)
				if err != nil {
					log.Printf("获取 %s 价格失败: %v", symbol, err)
				} else {
					log.Printf("%s 当前价格: $%.4f", symbol, price.USDPrice)
				}
			}()
		case <-stopCh:
			log.Printf("停止监控 CoinGecko 上 %s 的价格", symbol)
			return
		}
	}
}

// MonitorMultipleCoinGeckoTokens 同时监控多个CoinGecko代币的价格
func MonitorMultipleCoinGeckoTokens(tokens []CoinGeckoTokenConfig, interval time.Duration) chan struct{} {
	log.Printf("开始监控 %d 个代币在 CoinGecko 上的价格，间隔: %v", len(tokens), interval)

	// 创建一个全局停止信号通道
	stopCh := make(chan struct{})

	// 为每个代币启动一个监控协程
	for _, token := range tokens {
		go MonitorCoinGeckoPrice(token.CoinGeckoID, token.Symbol, token.Chain, interval, stopCh)
	}

	return stopCh
}

// SaveCoinGeckoPriceToDb 将CoinGecko价格数据保存到数据库
func SaveCoinGeckoPriceToDb(coinGeckoID string, symbol string, chain string, usdPrice float64) {
	// 创建价格数据对象
	priceData := map[string]interface{}{
		"coingecko_id": coinGeckoID,
		"symbol":       symbol,
		"usd_price":    usdPrice,
		"chain":        chain,
		"last_update":  time.Now(),
	}

	// 将原始数据转换为JSON字符串
	rawData, _ := json.Marshal(priceData)

	// 创建上下文
	ctx := context.Background()

	// 先保存或更新货币信息
	// 尝试从数据库中获取货币信息
	currenc, err := currency.GetCurrencyBySymbol(ctx, symbol)
	if err != nil {
		// 货币不存在，创建新的货币记录
		now := time.Now()
		newCurrency := &sqlite3.Currency{
			Symbol:          symbol,
			Name:            symbol, // 使用Symbol作为Name，后续可以更新
			Decimals:        18,     // 默认小数位数，可以根据实际情况调整
			ContractAddress: "",     // CoinGecko不提供合约地址
			Chain:           chain,  // 使用提供的链信息
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
		if currenc.Chain == "" && chain != "" {
			currenc.Chain = chain
		}

		err = currency.UpdateCurrency(ctx, currenc)
		if err != nil {
			log.Printf("更新货币信息失败: %v", err)
		} else {
			log.Printf("已更新货币 %s 信息", currenc.Symbol)
		}
	}

	// 获取交易对ID
	tradingPairID := 4 // 默认值，假设CoinGecko的exchange_id为4
	if currenc != nil && currenc.ID > 0 {
		tradingPairID = currenc.ID
	}

	// 创建 MarketData 对象
	marketData := &sqlite3.MarketData{
		ExchangeID:    4,                    // 假设 CoinGecko 的 exchange_id 为 4，实际应该从配置或数据库中获取
		TradingPairID: int64(tradingPairID), // 使用货币ID作为交易对ID
		Timestamp:     time.Now(),
		LastPrice:     usdPrice,
		BidPrice:      0, // CoinGecko API 不提供这些数据
		AskPrice:      0,
		Volume24h:     0,
		High24h:       0,
		Low24h:        0,
		OpenPrice24h:  0,
		ClosePrice24h: usdPrice, // 使用最新价格作为收盘价
		LiquidityUSD:  0,
		SlippageBPS:   0,
		SourceDataRaw: string(rawData),
	}

	// 保存市场数据到数据库
	err = marketdata.InsertMarketData(ctx, marketData)
	if err != nil {
		log.Printf("保存CoinGecko市场数据到数据库失败: %v", err)
	} else {
		log.Printf("已保存 %s CoinGecko市场数据到数据库，ID: %d", symbol, marketData.ID)
	}
}

// CoinGeckoExample 展示如何使用CoinGecko价格模块的示例
func CoinGeckoExample() {
	// 示例1: 获取单个代币价格
	log.Println("示例1: 获取ETH代币价格")
	ethPrice, err := GetCoinGeckoTokenPrice("ethereum", "ETH", "ethereum")
	if err != nil {
		log.Printf("获取ETH价格失败: %v\n", err)
	} else {
		log.Printf("ETH价格: $%.4f\n", ethPrice.USDPrice)
	}

	// 示例2: 获取多个代币价格
	log.Println("示例2: 获取多个代币价格")
	tokens := []CoinGeckoToken{
		{Symbol: "BTC", CoinGeckoID: "bitcoin", Chain: "bitcoin"},
		{Symbol: "ETH", CoinGeckoID: "ethereum", Chain: "ethereum"},
		{Symbol: "USDT", CoinGeckoID: "tether", Chain: "ethereum"},
	}

	prices, err := GetMultipleTokenPrices(tokens)
	if err != nil {
		log.Printf("获取多个代币价格失败: %v\n", err)
	} else {
		for symbol, price := range prices {
			log.Printf("%s价格: $%.4f\n", symbol, price.USDPrice)
		}
	}

	// 示例3: 监控多个代币价格
	log.Println("示例3: 监控多个代币价格")
	tokenConfigs := []CoinGeckoTokenConfig{
		{Symbol: "BTC", CoinGeckoID: "bitcoin", Chain: "bitcoin"},
		{Symbol: "ETH", CoinGeckoID: "ethereum", Chain: "ethereum"},
		{Symbol: "USDT", CoinGeckoID: "tether", Chain: "ethereum"},
	}

	// 每30秒更新一次价格
	stopCh := MonitorMultipleCoinGeckoTokens(tokenConfigs, 30*time.Second)

	// 运行5分钟后停止
	log.Println("将在5分钟后停止监控")
	time.Sleep(5 * time.Minute)
	close(stopCh)
	log.Println("已停止所有监控")
}
