package handling

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"time"

	"gocryptotrader/database/models/sqlite3"
	"gocryptotrader/database/repository/currency"
	marketdata "gocryptotrader/database/repository/market_data"

	"github.com/daoleno/uniswapv3-sdk/examples/contract"
	"github.com/daoleno/uniswapv3-sdk/examples/helper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Uniswap specific constants (using common constants from constants.go)
// Note: Common constants like EthereumRPCURL, ETHAddress, USDCAddress, UniswapDefaultFee are now defined in constants.go

// UniswapTokenPrice represents the calculated price information for a token
type UniswapTokenPrice struct {
	Address    string    `json:"address"`
	Symbol     string    `json:"symbol"`
	USDPrice   float64   `json:"usd_price"`
	ETHPrice   float64   `json:"eth_price"`
	LastUpdate time.Time `json:"last_update"`
}

// GetUniswapTokenPrice fetches the price of a token in USD and ETH
func GetUniswapTokenPrice(tokenAddress string, symbol string) (*UniswapTokenPrice, error) {
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address cannot be empty")
	}

	// If the token is ETH/WETH itself, return a simple response with ETH/USD price
	if tokenAddress == WETHAddress {
		return getETHUSDPrice(symbol)
	}

	// Connect to Ethereum network
	client, err := ethclient.Dial(EthereumRPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum network: %w", err)
	}

	// Create quoter contract instance
	quoterContract, err := contract.NewUniswapv3Quoter(common.HexToAddress(helper.ContractV3Quoter), client)
	if err != nil {
		return nil, fmt.Errorf("failed to create quoter contract: %w", err)
	}

	// Get token/ETH price
	ethPrice, err := getTokenETHPrice(quoterContract, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get token/ETH price: %w", err)
	}

	// Get ETH/USD price
	ethUSDPrice, err := getETHUSDPriceValue(quoterContract)
	if err != nil {
		return nil, fmt.Errorf("failed to get ETH/USD price: %w", err)
	}

	// Calculate USD price
	usdPrice := ethPrice * ethUSDPrice

	// Save to database
	SaveUniswapPriceToDb(tokenAddress, symbol, usdPrice, ethPrice)

	return &UniswapTokenPrice{
		Address:    tokenAddress,
		Symbol:     symbol,
		USDPrice:   usdPrice,
		ETHPrice:   ethPrice,
		LastUpdate: time.Now(),
	}, nil
}

// getETHUSDPrice fetches the price of ETH in USD
func getETHUSDPrice(symbol string) (*UniswapTokenPrice, error) {
	// Connect to Ethereum network
	client, err := ethclient.Dial(EthereumRPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum network: %w", err)
	}

	// Create quoter contract instance
	quoterContract, err := contract.NewUniswapv3Quoter(common.HexToAddress(helper.ContractV3Quoter), client)
	if err != nil {
		return nil, fmt.Errorf("failed to create quoter contract: %w", err)
	}

	// Get ETH/USD price
	ethUSDPrice, err := getETHUSDPriceValue(quoterContract)
	if err != nil {
		return nil, fmt.Errorf("failed to get ETH/USD price: %w", err)
	}

	// Save to database
	SaveUniswapPriceToDb(WETHAddress, symbol, ethUSDPrice, 1.0)

	return &UniswapTokenPrice{
		Address:    WETHAddress,
		Symbol:     symbol,
		USDPrice:   ethUSDPrice,
		ETHPrice:   1.0, // 1 ETH = 1 ETH
		LastUpdate: time.Now(),
	}, nil
}

// getTokenETHPrice gets the price of a token in ETH
func getTokenETHPrice(quoterContract *contract.Uniswapv3Quoter, tokenAddress string) (float64, error) {
	// 设置代币对：Token 和 WETH
	token0 := common.HexToAddress(tokenAddress)
	token1 := common.HexToAddress(WETHAddress)

	fee := big.NewInt(int64(CommonUniswapDefaultFee))  // 0.05%费率池
	amountIn := helper.FloatStringToBigInt("1.00", 18) // 查询1个代币的价格
	sqrtPriceLimitX96 := big.NewInt(0)                 // 无价格限制

	var out []interface{}
	rawCaller := &contract.Uniswapv3QuoterRaw{Contract: quoterContract}

	// 调用合约获取兑换率
	err := rawCaller.Call(nil, &out, "quoteExactInputSingle", token0, token1,
		fee, amountIn, sqrtPriceLimitX96)
	if err != nil {
		return 0, fmt.Errorf("failed to call contract: %w", err)
	}

	// 检查并处理结果
	if len(out) == 0 {
		return 0, fmt.Errorf("no valid result returned")
	}

	amountOut, ok := out[0].(*big.Int)
	if !ok {
		return 0, fmt.Errorf("invalid result type")
	}

	// 将结果转换为可读格式（WETH有18位小数）
	amountOutFloat := new(big.Float).SetInt(amountOut)
	ethPrice, _ := new(big.Float).Quo(amountOutFloat, big.NewFloat(math.Pow10(18))).Float64()

	return ethPrice, nil
}

// getETHUSDPriceValue gets the price of ETH in USD
func getETHUSDPriceValue(quoterContract *contract.Uniswapv3Quoter) (float64, error) {
	// 设置代币对：WETH 和 USDC
	token0 := common.HexToAddress(WETHAddress)
	token1 := common.HexToAddress(USDCAddress)

	fee := big.NewInt(int64(CommonUniswapDefaultFee))  // 0.05%费率池（WETH/USDC常用费率）
	amountIn := helper.FloatStringToBigInt("1.00", 18) // 查询1个ETH的价格
	sqrtPriceLimitX96 := big.NewInt(0)                 // 无价格限制

	var out []interface{}
	rawCaller := &contract.Uniswapv3QuoterRaw{Contract: quoterContract}

	// 调用合约获取兑换率
	err := rawCaller.Call(nil, &out, "quoteExactInputSingle", token0, token1,
		fee, amountIn, sqrtPriceLimitX96)
	if err != nil {
		return 0, fmt.Errorf("failed to call contract: %w", err)
	}

	// 检查并处理结果
	if len(out) == 0 {
		return 0, fmt.Errorf("no valid result returned")
	}

	amountOut, ok := out[0].(*big.Int)
	if !ok {
		return 0, fmt.Errorf("invalid result type")
	}

	// 将结果转换为可读格式（USDC有6位小数）
	amountOutFloat := new(big.Float).SetInt(amountOut)
	usdPrice, _ := new(big.Float).Quo(amountOutFloat, big.NewFloat(math.Pow10(6))).Float64()

	return usdPrice, nil
}

// SaveUniswapPriceToDb saves Uniswap price data to the database
func SaveUniswapPriceToDb(tokenAddress string, symbol string, usdPrice float64, ethPrice float64) {
	// Create price data object
	priceData := map[string]interface{}{
		"address":     tokenAddress,
		"symbol":      symbol,
		"usd_price":   usdPrice,
		"eth_price":   ethPrice,
		"last_update": time.Now(),
	}

	// Convert raw data to JSON string
	rawData, _ := json.Marshal(priceData)

	// Create context
	ctx := context.Background()

	// First save or update currency information
	// Try to get currency information from database
	currenc, err := currency.GetCurrencyBySymbol(ctx, symbol)
	if err != nil {
		// Currency doesn't exist, create a new currency record
		now := time.Now()
		newCurrency := &sqlite3.Currency{
			Symbol:          symbol,
			Name:            symbol, // Use Symbol as Name, can be updated later
			Decimals:        18,     // Default decimal places for ERC20 tokens
			ContractAddress: tokenAddress,
			Chain:           "ethereum", // Uniswap is on Ethereum
			IsActive:        true,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		// Insert new currency record
		err = currency.InsertCurrency(ctx, newCurrency)
		if err != nil {
			log.Printf("Failed to save currency information to database: %v", err)
		} else {
			log.Printf("Created new currency %s in database, ID: %d", newCurrency.Symbol, newCurrency.ID)
			currenc = newCurrency
		}
	} else {
		// Currency exists, update currency information
		currenc.UpdatedAt = time.Now()
		// If there are fields to update, set them here
		// For Uniswap, we can update contract address and chain information
		if currenc.ContractAddress == "" && tokenAddress != "" {
			currenc.ContractAddress = tokenAddress
		}
		if currenc.Chain == "" {
			currenc.Chain = "ethereum"
		}

		err = currency.UpdateCurrency(ctx, currenc)
		if err != nil {
			log.Printf("Failed to update currency information: %v", err)
		} else {
			log.Printf("Updated currency %s information", currenc.Symbol)
		}
	}

	// Get trading pair ID
	tradingPairID := 3 // Default value
	if currenc != nil && currenc.ID > 0 {
		tradingPairID = currenc.ID
	}

	// Create MarketData object
	marketData := &sqlite3.MarketData{
		ExchangeID:    3,                    // Assume Uniswap's exchange_id is 3, should be retrieved from config or database
		TradingPairID: int64(tradingPairID), // Use currency ID as trading pair ID
		Timestamp:     time.Now(),
		LastPrice:     usdPrice,
		BidPrice:      0, // Uniswap doesn't provide these data
		AskPrice:      0,
		Volume24h:     0,
		High24h:       0,
		Low24h:        0,
		OpenPrice24h:  0,
		ClosePrice24h: usdPrice, // Use latest price as closing price
		LiquidityUSD:  0,
		SlippageBPS:   0,
		SourceDataRaw: string(rawData),
	}

	// Save market data to database
	err = marketdata.InsertMarketData(ctx, marketData)
	if err != nil {
		log.Printf("Failed to save Uniswap market data to database: %v", err)
	} else {
		log.Printf("Saved %s Uniswap market data to database, ID: %d", symbol, marketData.ID)
	}
}

// UniswapExample 演示如何使用Uniswap价格查询功能
func UniswapExample() {
	// 获取ETH价格
	ethPrice, err := GetUniswapTokenPrice(WETHAddress, "ETH")
	if err != nil {
		log.Printf("获取ETH价格失败: %v", err)
		return
	}
	log.Printf("ETH价格: $%.2f", ethPrice.USDPrice)

	// 获取UNI代币价格
	uniAddress := "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984" // UNI代币地址
	uniPrice, err := GetUniswapTokenPrice(uniAddress, "UNI")
	if err != nil {
		log.Printf("获取UNI价格失败: %v", err)
		return
	}
	log.Printf("UNI价格: $%.2f, ETH价格: %.6f ETH", uniPrice.USDPrice, uniPrice.ETHPrice)
}
