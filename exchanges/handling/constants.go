package handling

import "time"

// Common token addresses used across different exchanges
const (
	// Ethereum addresses
	ETHAddress     = "0x0000000000000000000000000000000000000000" // Native ETH
	WETHAddress    = "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2" // Wrapped ETH
	USDCAddress    = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48" // USDC (Ethereum mainnet)
	USDCETHAddress = "0xA0b86a33E6441b8C4505B6B8C0E4b4e8C5b8b8b8" // USDC on ETH (example)

	// BSC addresses
	WBNBAddress    = "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c" // Wrapped BNB
	BUSDAddress    = "0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56" // BUSD
	USDTAddress    = "0x55d398326f99059fF775485246999027B3197955" // USDT (BSC)
	USDCBSCAddress = "0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d" // USDC (BSC)
	CAKEAddress    = "0x0E09FaBB73Bd3Ade0a17ECC321fD13a19e81cE82" // CAKE

	// Solana addresses
	SOLAddress = "So11111111111111111111111111111111111111112"  // SOL
	USDAddress = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" // USD (Solana)
)

// Common RPC URLs
const (
	BSCRPCURL      = "https://bsc-dataseed.binance.org/"
	EthereumRPCURL = "https://mainnet.infura.io/v3/154e98baa16348e5ac3ad1fc05f9e257"
)

// Common contract addresses
const (
	PancakeRouterAddress = "0x10ED43C718714eb63d5aA57B78B54704E256024E" // PancakeSwap V2 Router
)

// Common default values
const (
	CommonDefaultTimeout    = 30 * time.Second
	DefaultAmountIn         = 10.0 // Default input amount for 10 tokens (increased for better precision)
	DefaultSlippage         = 10.0
	DefaultFee              = 0.006
	CommonUniswapDefaultFee = 500 // 0.05% fee tier
)

// Common default amounts in different units
const (
	DefaultSOLAmount      = "1000000000"          // 1 SOL in lamports
	DefaultTokenAmount    = "1000000"             // A small amount of token
	DefaultETHAmount      = "1000000000000000000" // 1 ETH in wei
	DefaultETHFromAddress = "0xb2bA9EA690428E01d82572438De4514F16446251"
	DefaultSOLFromAddress = "vp4ppQ97v9aAeAbBBEgnxdyxjjzWaKhorjCRHewZ7rR"
)

// API Base URLs
const (
	GMGNBaseURL      = "https://gmgn.ai"
	CoinGeckoBaseURL = "https://api.coingecko.com/api/v3"
	BinanceWSBaseURL = "wss://stream.binance.com"
)

// API Endpoints
const (
	// GMGN endpoints
	SwapRouteEndpoint = "/defi/router/v1/sol/tx/get_swap_route"
	ExactInEndpoint   = "/defi/router/v1/tx/available_routes_exact_in"
	ExactOutEndpoint  = "/defi/router/v1/tx/available_routes_exact_out"

	// CoinGecko endpoints
	CommonPriceEndpoint = "/simple/price"

	// Binance endpoints
	BinanceTickerPath = "/ws/%s@ticker"
)
