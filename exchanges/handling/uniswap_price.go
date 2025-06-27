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
	currency "gocryptotrader/database/repository/currency"
	marketdata "gocryptotrader/database/repository/market_data"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Uniswap API and contract constants
const (
	// Ethereum RPC URL
	EthereumRPCURL = "https://mainnet.infura.io/v3/154e98baa16348e5ac3ad1fc05f9e257"

	// Token address constants
	WETHAddress = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2" // WETH (以太坊主网)
	UniswapUSDCAddress = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48" // USDC (以太坊主网)

	// Uniswap V3 Quoter contract address
	UniswapV3QuoterAddress = "0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6"

	// Default values
	UniswapDefaultFee = 500 // 0.05% fee tier
)

// UniswapTokenPrice represents the calculated price information for a token
type UniswapTokenPrice struct {
	Address    string    `json:"address"`
	Symbol     string    `json:"symbol"`
	USDPrice   float64   `json:"usd_price"`
	ETHPrice   float64   `json:"eth_price"`
	LastUpdate time.Time `json:"last_update"`
}

// UniswapQuoterABI is the ABI for the Uniswap V3 Quoter contract
const UniswapQuoterABI = `[{"inputs":[{"internalType":"address","name":"tokenIn","type":"address"},{"internalType":"address","name":"tokenOut","type":"address"},{"internalType":"uint24","name":"fee","type":"uint24"},{"internalType":"uint256","name":"amountIn","type":"uint256"},{"internalType":"uint160","name":"sqrtPriceLimitX96","type":"uint160"}],"name":"quoteExactInputSingle","outputs":[{"internalType":"uint256","name":"amountOut","type":"uint256"}],"stateMutability":"nonpayable","type":"function"}]`

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

	// Create a new instance of the Quoter contract
	quoterAddress := common.HexToAddress(UniswapV3QuoterAddress)
	abi, err := parseQuoterABI()
	if err != nil {
		return nil, fmt.Errorf("failed to parse Quoter ABI: %w", err)
	}

	// Get token/ETH price
	ethPrice, err := getTokenETHPrice(client, abi, tokenAddress, quoterAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get token/ETH price: %w", err)
	}

	// Get ETH/USD price
	ethUSDPrice, err := getETHUSDPriceValue(client, abi, quoterAddress)
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

	// Create a new instance of the Quoter contract
	quoterAddress := common.HexToAddress(UniswapV3QuoterAddress)
	abi, err := parseQuoterABI()
	if err != nil {
		return nil, fmt.Errorf("failed to parse Quoter ABI: %w", err)
	}

	// Get ETH/USD price
	ethUSDPrice, err := getETHUSDPriceValue(client, abi, quoterAddress)
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

// parseQuoterABI parses the Uniswap V3 Quoter ABI
func parseQuoterABI() (interface{}, error) {
	// In a real implementation, you would use abigen to generate Go bindings
	// For simplicity, we'll return the ABI string
	return UniswapQuoterABI, nil
}

// getTokenETHPrice gets the price of a token in ETH
func getTokenETHPrice(client *ethclient.Client, abi interface{}, tokenAddress string, quoterAddress common.Address) (float64, error) {
	// In a real implementation, you would use the contract bindings to call the contract
	// For this example, we'll simulate a response
	
	// Create a token amount of 1 token with 18 decimals
	// This would be used in the actual contract call
	_ = new(big.Int).Mul(big.NewInt(1), big.NewInt(int64(math.Pow10(18))))
	
	// Call the Uniswap V3 Quoter contract
	// This is a simplified version; in reality, you would use the contract bindings
	// to call the quoteExactInputSingle function
	
	// Simulate a response for demonstration purposes
	// In a real implementation, you would get this from the contract call
	amountOut := new(big.Int).Mul(big.NewInt(100), big.NewInt(int64(math.Pow10(6)))) // Example: 100 USDC
	
	// Convert to float64
	amountOutFloat := new(big.Float).SetInt(amountOut)
	result, _ := new(big.Float).Quo(amountOutFloat, big.NewFloat(math.Pow10(6))).Float64()
	
	return result, nil
}

// getETHUSDPriceValue gets the price of ETH in USD
func getETHUSDPriceValue(client *ethclient.Client, abi interface{}, quoterAddress common.Address) (float64, error) {
	// In a real implementation, you would use the contract bindings to call the contract
	// For this example, we'll simulate a response
	
	// Create an ETH amount of 1 ETH with 18 decimals
	// This would be used in the actual contract call
	_ = new(big.Int).Mul(big.NewInt(1), big.NewInt(int64(math.Pow10(18))))
	
	// Call the Uniswap V3 Quoter contract
	// This is a simplified version; in reality, you would use the contract bindings
	// to call the quoteExactInputSingle function
	
	// Simulate a response for demonstration purposes
	// In a real implementation, you would get this from the contract call
	amountOut := new(big.Int).Mul(big.NewInt(1800), big.NewInt(int64(math.Pow10(6)))) // Example: 1800 USDC
	
	// Convert to float64
	amountOutFloat := new(big.Float).SetInt(amountOut)
	result, _ := new(big.Float).Quo(amountOutFloat, big.NewFloat(math.Pow10(6))).Float64()
	
	return result, nil
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

// UniswapExample demonstrates how to use the Uniswap price module
func UniswapExample() {
	// Get ETH price in USD
	ethPrice, err := GetUniswapTokenPrice(WETHAddress, "ETH")
	if err != nil {
		log.Printf("Failed to get ETH price: %v", err)
		return
	}
	log.Printf("ETH Price: $%.2f", ethPrice.USDPrice)

	// Get a token price (e.g., UNI token)
	uniTokenAddress := "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984" // UNI token address
	uniPrice, err := GetUniswapTokenPrice(uniTokenAddress, "UNI")
	if err != nil {
		log.Printf("Failed to get UNI price: %v", err)
		return
	}
	log.Printf("UNI Price: $%.2f, %.8f ETH", uniPrice.USDPrice, uniPrice.ETHPrice)
}