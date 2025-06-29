package handling

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gocryptotrader/database/models/sqlite3"
	currency "gocryptotrader/database/repository/currency"
	marketdata "gocryptotrader/database/repository/market_data"
)

// GMGN specific constants (using common constants from constants.go)
// Note: Common constants like GMGNBaseURL, SwapRouteEndpoint, ExactInEndpoint, ExactOutEndpoint,
// SOLAddress, USDAddress, ETHAddress, WETHAddress, USDCETHAddress, DefaultSlippage, DefaultFee,
// DefaultSOLAmount, DefaultTokenAmount, DefaultETHAmount, DefaultETHFromAddress, DefaultSOLFromAddress
// are now defined in constants.go

// SwapRouteParams represents the parameters for the swap route API
type SwapRouteParams struct {
	TokenInAddress  string  `json:"token_in_address"`
	TokenOutAddress string  `json:"token_out_address"`
	InAmount        string  `json:"in_amount"`
	FromAddress     string  `json:"from_address"`
	Slippage        float64 `json:"slippage"`
	SwapMode        string  `json:"swap_mode,omitempty"`
	Fee             float64 `json:"fee,omitempty"`
	IsAntiMEV       bool    `json:"is_anti_mev,omitempty"`
	Partner         string  `json:"partner,omitempty"`
}

// SwapInfo represents the swap information in the route plan
type SwapInfo struct {
	Label      string `json:"label"`
	InputMint  string `json:"inputMint"`
	OutputMint string `json:"outputMint"`
	InAmount   string `json:"inAmount"`
	OutAmount  string `json:"outAmount"`
	FeeAmount  string `json:"feeAmount"`
	FeeMint    string `json:"feeMint"`
}

// RoutePlan represents a single route in the route plan
type RoutePlan struct {
	SwapInfo SwapInfo `json:"swapInfo"`
	Percent  int      `json:"percent"`
}

// Quote represents the quote information in the response
type Quote struct {
	InputMint            string      `json:"inputMint"`
	InAmount             string      `json:"inAmount"`
	OutputMint           string      `json:"outputMint"`
	OutAmount            string      `json:"outAmount"`
	OtherAmountThreshold string      `json:"otherAmountThreshold"`
	InDecimals           int         `json:"inDecimals"`
	OutDecimals          int         `json:"outDecimals"`
	SwapMode             string      `json:"swapMode"`
	SlippageBps          string      `json:"slippageBps"`
	PlatformFee          string      `json:"platformFee"`
	PriceImpactPct       string      `json:"priceImpactPct"`
	RoutePlan            []RoutePlan `json:"routePlan"`
	TimeTaken            float64     `json:"timeTaken"`
}

// RawTransaction represents the raw transaction information
type RawTransaction struct {
	SwapTransaction           string `json:"swapTransaction"`
	LastValidBlockHeight      int64  `json:"lastValidBlockHeight"`
	PrioritizationFeeLamports int    `json:"prioritizationFeeLamports"`
	RecentBlockhash           string `json:"recentBlockhash"`
	Version                   string `json:"version"`
}

// SwapRouteData represents the data in the response
type SwapRouteData struct {
	Quote        Quote          `json:"quote"`
	RawTx        RawTransaction `json:"raw_tx"`
	AmountInUSD  string         `json:"amount_in_usd"`
	AmountOutUSD string         `json:"amount_out_usd"`
	JitoOrderID  interface{}    `json:"jito_order_id"`
}

// SwapRouteResponse represents the response from the swap route API
type SwapRouteResponse struct {
	Code int           `json:"code"`
	Msg  string        `json:"msg"`
	Tid  string        `json:"tid"`
	Data SwapRouteData `json:"data"`
}

// GMGNTokenPrice represents the calculated price information for a token
type GMGNTokenPrice struct {
	Address    string    `json:"address"`
	Symbol     string    `json:"symbol"`
	USDPrice   float64   `json:"usd_price"`
	SOLPrice   float64   `json:"sol_price"`
	Chain      string    `json:"chain"`
	LastUpdate time.Time `json:"last_update"`
}

// MultiChainRouteParams represents parameters for multi-chain route API
type MultiChainRouteParams struct {
	TokenInChain    string `json:"token_in_chain"`
	TokenOutChain   string `json:"token_out_chain"`
	TokenInAddress  string `json:"token_in_address"`
	TokenOutAddress string `json:"token_out_address"`
	InAmount        string `json:"in_amount,omitempty"`
	OutAmount       string `json:"out_amount,omitempty"`
	Src             string `json:"src,omitempty"`
}

// RouteStep represents a step in the route
type RouteStep struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
	Tool string `json:"tool"`
}

// Route represents a single route in the multi-chain response
type Route struct {
	ChainID            int         `json:"chain_id"`
	To                 string      `json:"to"`
	AmountIn           string      `json:"amount_in"`
	AmountOut          string      `json:"amount_out"`
	InputTokenAddress  string      `json:"input_token_address"`
	OutputTokenAddress string      `json:"output_token_address"`
	Type               string      `json:"type"`
	Path               []string    `json:"path"`
	PoolAddress        interface{} `json:"pool_address"`    // Can be string or []string
	FactoryAddress     interface{} `json:"factory_address"` // Can be string or []string
	Fee                interface{} `json:"fee"`             // Can be int or []int
	Steps              []RouteStep `json:"steps"`
	TokenInUSDPrice    string      `json:"token_in_usd_price"`
	AmountInUSD        string      `json:"amount_in_usd"`
	TokenOutUSDPrice   string      `json:"token_out_usd_price"`
	AmountOutUSD       string      `json:"amount_out_usd"`
	Value              string      `json:"value"`
	PriceImpact        string      `json:"price_impact"`
	GasLimit           string      `json:"gas_limit"`
}

// Volatilities represents price volatility information
type Volatilities struct {
	TokenIn  int  `json:"token_in"`
	TokenOut int  `json:"token_out"`
	IsFomo   bool `json:"is_fomo"`
}

// MultiChainRouteData represents the data in multi-chain route response
type MultiChainRouteData struct {
	Routes       []Route      `json:"routes"`
	Volatilities Volatilities `json:"volatilities"`
}

// MultiChainRouteResponse represents the response from multi-chain route API
type MultiChainRouteResponse struct {
	Code int                 `json:"code"`
	Msg  string              `json:"msg"`
	Data MultiChainRouteData `json:"data"`
}

// parseAmounts parses the amount values from the API response
func parseAmounts(response *SwapRouteResponse) (float64, float64, float64, float64, error) {
	// Parse USD amounts
	amountInUSD, err := strconv.ParseFloat(response.Data.AmountInUSD, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to parse amount_in_usd: %w", err)
	}

	amountOutUSD, err := strconv.ParseFloat(response.Data.AmountOutUSD, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to parse amount_out_usd: %w", err)
	}

	// Parse token amounts
	inAmount, err := strconv.ParseFloat(response.Data.Quote.InAmount, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to parse inAmount: %w", err)
	}

	outAmount, err := strconv.ParseFloat(response.Data.Quote.OutAmount, 64)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to parse outAmount: %w", err)
	}

	return amountInUSD, amountOutUSD, inAmount, outAmount, nil
}

// createDefaultParams creates default parameters for the swap route API
func createDefaultParams(tokenInAddress, tokenOutAddress, inAmount string) SwapRouteParams {
	return SwapRouteParams{
		TokenInAddress:  tokenInAddress,
		TokenOutAddress: tokenOutAddress,
		InAmount:        inAmount,
		FromAddress:     DefaultSOLFromAddress,
		Slippage:        DefaultSlippage,
		Fee:             DefaultFee,
		IsAntiMEV:       false,
	}
}

// GetGMGNTokenPrice fetches the price of a token across multiple chains (SOL/ETH/Base/BSC)
// It tries each chain in order until a successful response is found
func GetGMGNTokenPrice(tokenAddress string, symbol string) (*GMGNTokenPrice, error) {
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address cannot be empty")
	}

	// Define the chains to try in order
	chains := []string{"sol", "eth", "base", "bsc"}

	// Try each chain until we get a successful response
	for _, chain := range chains {
		price, err := getTokenPriceOnChain(tokenAddress, symbol, chain)
		if err == nil {
			return price, nil
		}
		// Log the error but continue to next chain
		log.Printf("Failed to get price for %s on %s chain: %v", symbol, chain, err)
	}

	return nil, fmt.Errorf("failed to get price for token %s on all supported chains", symbol)
}

// getGMGNSOLPrice fetches the price of SOL in USD
func getGMGNSOLPrice(symbol string) (*GMGNTokenPrice, error) {
	// Default parameters for the swap route API
	params := createDefaultParams(SOLAddress, USDAddress, DefaultSOLAmount)

	// Get the price using the swap route API
	response, err := fetchSwapRoute(params)
	if err != nil {
		return nil, err
	}

	// Parse amounts from response
	amountInUSD, _, _, _, err := parseAmounts(response)
	if err != nil {
		return nil, err
	}

	// 保存到数据库
	SaveGMGNPriceToDb(SOLAddress, symbol, amountInUSD, 1.0, "sol")

	// SOL price in USD is directly available
	return &GMGNTokenPrice{
		Address:    SOLAddress,
		Symbol:     symbol,
		USDPrice:   amountInUSD,
		SOLPrice:   1.0, // 1 SOL = 1 SOL
		Chain:      "sol",
		LastUpdate: time.Now(),
	}, nil
}

// makeHTTPRequest makes a generic HTTP GET request and returns the response body
func makeHTTPRequest(url string, timeout time.Duration) ([]byte, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set common headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// fetchSwapRoute makes a request to the GMGN API to get the swap route
func fetchSwapRoute(params SwapRouteParams) (*SwapRouteResponse, error) {
	if params.TokenInAddress == "" || params.TokenOutAddress == "" {
		return nil, fmt.Errorf("token addresses cannot be empty")
	}

	// Build the URL with query parameters
	baseURL, err := url.Parse(GMGNBaseURL + SwapRouteEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	query := baseURL.Query()
	query.Set("token_in_address", params.TokenInAddress)
	query.Set("token_out_address", params.TokenOutAddress)
	query.Set("in_amount", params.InAmount)
	query.Set("from_address", params.FromAddress)
	query.Set("slippage", fmt.Sprintf("%v", params.Slippage))

	if params.SwapMode != "" {
		query.Set("swap_mode", params.SwapMode)
	}

	if params.Fee > 0 {
		query.Set("fee", fmt.Sprintf("%v", params.Fee))
	}

	if params.IsAntiMEV {
		query.Set("is_anti_mev", "true")
	}

	if params.Partner != "" {
		query.Set("partner", params.Partner)
	}

	baseURL.RawQuery = query.Encode()

	// Make HTTP request
	body, err := makeHTTPRequest(baseURL.String(), 10*time.Second)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var response SwapRouteResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if the API returned an error
	if response.Code != 0 {
		return nil, fmt.Errorf("API returned error: %s", response.Msg)
	}

	return &response, nil
}

// SaveGMGNPriceToDb 将GMGN价格数据保存到数据库
// fetchMultiChainRoute fetches route data from multi-chain API
func fetchMultiChainRoute(params MultiChainRouteParams, exactIn bool) (*MultiChainRouteResponse, error) {
	// Choose endpoint based on exactIn parameter
	endpoint := ExactInEndpoint
	if !exactIn {
		endpoint = ExactOutEndpoint
	}

	// Build URL with query parameters
	baseURL, err := url.Parse(GMGNBaseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Add query parameters
	q := baseURL.Query()
	q.Add("token_in_chain", params.TokenInChain)
	q.Add("token_out_chain", params.TokenOutChain)
	q.Add("token_in_address", params.TokenInAddress)
	q.Add("token_out_address", params.TokenOutAddress)
	if exactIn && params.InAmount != "" {
		q.Add("in_amount", params.InAmount)
	}
	if !exactIn && params.OutAmount != "" {
		q.Add("out_amount", params.OutAmount)
	}
	if params.Src != "" {
		q.Add("src", params.Src)
	}
	baseURL.RawQuery = q.Encode()

	// Make HTTP request
	body, err := makeHTTPRequest(baseURL.String(), 30*time.Second)
	if err != nil {
		return nil, err
	}

	// Parse response
	var response MultiChainRouteResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Check API response code
	if response.Code != 0 {
		return nil, fmt.Errorf("API error: %s", response.Msg)
	}

	return &response, nil
}

// getBaseCurrencyAddress returns the base currency address for a given chain
func getBaseCurrencyAddress(chain string) (string, error) {
	switch strings.ToLower(chain) {
	case "eth":
		return WETHAddress, nil
	case "base":
		return ETHAddress, nil // Base uses ETH as native currency
	case "bsc":
		return WBNBAddress, nil
	default:
		return "", fmt.Errorf("unsupported chain: %s. Supported chains: eth, base, bsc", chain)
	}
}

// GetMultiChainTokenPrice fetches token price on ETH/Base/BSC chains
func GetMultiChainTokenPrice(tokenAddress, symbol, chain string) (*GMGNTokenPrice, error) {
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address cannot be empty")
	}
	if chain == "" {
		return nil, fmt.Errorf("chain cannot be empty")
	}

	// Special handling for WETH - use ETH price since they are 1:1
	if strings.EqualFold(tokenAddress, WETHAddress) && strings.ToLower(chain) == "eth" {
		return getETHPrice(symbol)
	}

	// Get base currency address for the chain
	baseCurrencyAddress, err := getBaseCurrencyAddress(chain)
	if err != nil {
		return nil, err
	}

	// Create parameters for token to base currency (ETH/BNB)
	params := MultiChainRouteParams{
		TokenInChain:    strings.ToLower(chain),
		TokenOutChain:   strings.ToLower(chain),
		TokenInAddress:  tokenAddress,
		TokenOutAddress: baseCurrencyAddress,
		InAmount:        DefaultTokenAmount,
		Src:             "gmgn",
	}

	// Fetch route data
	response, err := fetchMultiChainRoute(params, true)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch route data: %w", err)
	}

	if len(response.Data.Routes) == 0 {
		return nil, fmt.Errorf("no routes found for token %s on chain %s", tokenAddress, chain)
	}

	// Use the first route
	route := response.Data.Routes[0]

	// Parse amounts
	amountIn, err := strconv.ParseFloat(route.AmountIn, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount_in: %w", err)
	}

	amountOut, err := strconv.ParseFloat(route.AmountOut, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount_out: %w", err)
	}

	// Parse USD prices
	tokenInUSDPrice, err := strconv.ParseFloat(route.TokenInUSDPrice, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token_in_usd_price: %w", err)
	}

	tokenOutUSDPrice, err := strconv.ParseFloat(route.TokenOutUSDPrice, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token_out_usd_price: %w", err)
	}

	// Calculate USD price
	var usdPrice float64

	// USD price calculation
	if tokenInUSDPrice > 0 {
		usdPrice = tokenInUSDPrice
	} else if tokenOutUSDPrice > 0 && amountOut > 0 {
		// Calculate based on output token USD price
		usdPrice = (tokenOutUSDPrice * amountOut) / amountIn
	}

	// 保存到数据库
	SaveGMGNPriceToDb(tokenAddress, symbol, usdPrice, 0, strings.ToLower(chain))

	return &GMGNTokenPrice{
		Address:    tokenAddress,
		Symbol:     symbol,
		USDPrice:   usdPrice,
		SOLPrice:   0, // Not applicable for ETH/Base/BSC chains
		Chain:      strings.ToLower(chain),
		LastUpdate: time.Now(),
	}, nil
}

// getTokenPriceOnChain fetches token price on a specific chain
func getTokenPriceOnChain(tokenAddress, symbol, chain string) (*GMGNTokenPrice, error) {
	switch strings.ToLower(chain) {
	case "sol":
		return getSOLTokenPrice(tokenAddress, symbol)
	case "eth", "base", "bsc":
		return GetMultiChainTokenPrice(tokenAddress, symbol, chain)
	default:
		return nil, fmt.Errorf("unsupported chain: %s", chain)
	}
}

// getSOLTokenPrice fetches token price on Solana chain
func getSOLTokenPrice(tokenAddress, symbol string) (*GMGNTokenPrice, error) {
	// If the token is SOL itself, return a simple response
	if tokenAddress == SOLAddress {
		return getGMGNSOLPrice(symbol)
	}

	// Default parameters for the swap route API
	params := createDefaultParams(SOLAddress, tokenAddress, DefaultSOLAmount)

	// Get the price using the swap route API
	response, err := fetchSwapRoute(params)
	if err != nil {
		return nil, err
	}

	// Parse amounts from response
	_, amountOutUSD, inAmount, outAmount, err := parseAmounts(response)
	if err != nil {
		return nil, err
	}

	// Calculate the SOL price (how many tokens per 1 SOL)
	// Adjust for decimals
	inAmountInSOL := inAmount / math.Pow10(response.Data.Quote.InDecimals)
	outAmountInTokens := outAmount / math.Pow10(response.Data.Quote.OutDecimals)

	// Avoid division by zero
	if inAmountInSOL == 0 {
		return nil, fmt.Errorf("invalid calculation: inAmountInSOL is zero")
	}

	solPrice := outAmountInTokens / inAmountInSOL

	// Calculate the USD price (USD per token)
	// Avoid division by zero
	if outAmountInTokens == 0 {
		return nil, fmt.Errorf("invalid calculation: outAmountInTokens is zero")
	}

	usdPrice := amountOutUSD / outAmountInTokens

	// 保存到数据库
	SaveGMGNPriceToDb(tokenAddress, symbol, usdPrice, solPrice, "sol")

	return &GMGNTokenPrice{
		Address:    tokenAddress,
		Symbol:     symbol,
		USDPrice:   usdPrice,
		SOLPrice:   solPrice,
		Chain:      "sol",
		LastUpdate: time.Now(),
	}, nil
}

// GetETHTokenPrice is a convenience function for ETH chain tokens
func GetETHTokenPrice(tokenAddress, symbol string) (*GMGNTokenPrice, error) {
	return getTokenPriceOnChain(tokenAddress, symbol, "eth")
}

// GetBaseTokenPrice is a convenience function for Base chain tokens
func GetBaseTokenPrice(tokenAddress, symbol string) (*GMGNTokenPrice, error) {
	return getTokenPriceOnChain(tokenAddress, symbol, "base")
}

// GetBSCTokenPrice is a convenience function for BSC chain tokens
func GetBSCTokenPrice(tokenAddress, symbol string) (*GMGNTokenPrice, error) {
	return getTokenPriceOnChain(tokenAddress, symbol, "bsc")
}

// GetSOLTokenPrice is a convenience function for SOL chain tokens
func GetSOLTokenPrice(tokenAddress, symbol string) (*GMGNTokenPrice, error) {
	return getTokenPriceOnChain(tokenAddress, symbol, "sol")
}

func SaveGMGNPriceToDb(tokenAddress string, symbol string, usdPrice float64, solPrice float64, chain string) {
	// 创建价格数据对象
	priceData := map[string]interface{}{
		"address":     tokenAddress,
		"symbol":      symbol,
		"usd_price":   usdPrice,
		"sol_price":   solPrice,
		"last_update": time.Now(),
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
			Decimals:        8,      // 默认小数位数，可以根据实际情况调整
			ContractAddress: tokenAddress,
			Chain:           chain, // 使用传入的链名称
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
		// 对于GMGN，我们可以更新合约地址和链信息
		if currenc.ContractAddress == "" && tokenAddress != "" {
			currenc.ContractAddress = tokenAddress
		}
		if currenc.Chain == "" {
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
	tradingPairID := 2 // 默认值
	if currenc != nil && currenc.ID > 0 {
		tradingPairID = currenc.ID
	}

	// 创建 MarketData 对象
	marketData := &sqlite3.MarketData{
		ExchangeID:    2,                    // 假设 GMGN 的 exchange_id 为 2，实际应该从配置或数据库中获取
		TradingPairID: int64(tradingPairID), // 使用货币ID作为交易对ID
		Timestamp:     time.Now(),
		LastPrice:     usdPrice,
		BidPrice:      0, // GMGN API 不提供这些数据
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
		log.Printf("保存GMGN市场数据到数据库失败: %v", err)
	} else {
		log.Printf("已保存 %s GMGN市场数据到数据库，ID: %d", symbol, marketData.ID)
	}
}

// getETHPrice fetches ETH price using ETH->USDC route
func getETHPrice(symbol string) (*GMGNTokenPrice, error) {
	// Use ETH native address to get price against USDC
	params := MultiChainRouteParams{
		TokenInChain:    "eth",
		TokenOutChain:   "eth",
		TokenInAddress:  ETHAddress, // Native ETH
		TokenOutAddress: USDCAddress,
		InAmount:        DefaultETHAmount,
		Src:             "gmgn",
	}

	// Fetch route data
	response, err := fetchMultiChainRoute(params, true)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ETH route data: %w", err)
	}

	if len(response.Data.Routes) == 0 {
		return nil, fmt.Errorf("no routes found for ETH")
	}

	// Use the first route
	route := response.Data.Routes[0]

	// Parse amounts
	amountIn, err := strconv.ParseFloat(route.AmountIn, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount_in: %w", err)
	}

	amountOut, err := strconv.ParseFloat(route.AmountOut, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount_out: %w", err)
	}

	// Parse USD prices
	tokenInUSDPrice, err := strconv.ParseFloat(route.TokenInUSDPrice, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token_in_usd_price: %w", err)
	}

	tokenOutUSDPrice, err := strconv.ParseFloat(route.TokenOutUSDPrice, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token_out_usd_price: %w", err)
	}

	// Calculate USD price
	var usdPrice float64

	// USD price calculation
	if tokenInUSDPrice > 0 {
		usdPrice = tokenInUSDPrice
	} else if tokenOutUSDPrice > 0 && amountOut > 0 {
		// Calculate based on output token USD price
		usdPrice = (tokenOutUSDPrice * amountOut) / amountIn
	}

	// 保存到数据库 (use WETH address for consistency)
	SaveGMGNPriceToDb(WETHAddress, symbol, usdPrice, 0, "eth")

	return &GMGNTokenPrice{
		Address:    WETHAddress, // Return WETH address for consistency
		Symbol:     symbol,
		USDPrice:   usdPrice,
		SOLPrice:   0,
		Chain:      "eth",
		LastUpdate: time.Now(),
	}, nil
}
