package token

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// GMGN API URL constants
const (
	GMGNBaseURL       = "https://gmgn.ai"
	SwapRouteEndpoint = "/defi/router/v1/sol/tx/get_swap_route"

	// Token address constants
	SolAddress  = "So11111111111111111111111111111111111111112"
	USDCAddress = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"

	// Default values
	DefaultFromAddress = "5Jn2fbBaf9QQG4NsNeXEnM26Yar33atPuhjUBG8zUi1H"
	DefaultSlippage    = 10.0
	DefaultFee         = 0.006
	DefaultSolAmount   = "1000000000" // 1 SOL in lamports
	DefaultTokenAmount = "1000000"    // A small amount of token
)

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

// TokenPrice represents the calculated price information for a token
type TokenPrice struct {
	Address    string    `json:"address"`
	USDPrice   float64   `json:"usd_price"`
	SOLPrice   float64   `json:"sol_price"`
	LastUpdate time.Time `json:"last_update"`
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
		FromAddress:     DefaultFromAddress,
		Slippage:        DefaultSlippage,
		Fee:             DefaultFee,
		IsAntiMEV:       false,
	}
}

// GetTokenPrice fetches the price of a token in USD and SOL
func GetTokenPrice(tokenAddress string) (*TokenPrice, error) {
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address cannot be empty")
	}

	// If the token is SOL itself, return a simple response
	if tokenAddress == SolAddress {
		return getSOLPrice()
	}

	// Default parameters for the swap route API
	params := createDefaultParams(SolAddress, tokenAddress, DefaultSolAmount)

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

	return &TokenPrice{
		Address:    tokenAddress,
		USDPrice:   usdPrice,
		SOLPrice:   solPrice,
		LastUpdate: time.Now(),
	}, nil
}

// getSOLPrice fetches the price of SOL in USD
func getSOLPrice() (*TokenPrice, error) {
	// Default parameters for the swap route API
	params := createDefaultParams(SolAddress, USDCAddress, DefaultSolAmount)

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

	// SOL price in USD is directly available
	return &TokenPrice{
		Address:    SolAddress,
		USDPrice:   amountInUSD,
		SOLPrice:   1.0, // 1 SOL = 1 SOL
		LastUpdate: time.Now(),
	}, nil
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

	// Create a new HTTP client and request
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, baseURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if the response status code is not 200 OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status code: %d, body: %s", resp.StatusCode, string(body))
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
