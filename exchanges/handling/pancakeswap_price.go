package handling

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"gocryptotrader/database/models/sqlite3"
	currency "gocryptotrader/database/repository/currency"
	marketdata "gocryptotrader/database/repository/market_data"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// PancakeSwap specific constants (using common constants from constants.go)
// Note: Common constants like BSCRPCURL, PancakeRouterAddress, WBNBAddress, BUSDAddress,
// USDTAddress, CAKEAddress, WETHAddress, DefaultAmountIn are now defined in constants.go

// PancakeRouter ABI 片段 (仅包含 getAmountsOut 函数)
const pancakeRouterABI = `[{
	"inputs": [
		{"internalType":"uint256","name":"amountIn","type":"uint256"},
		{"internalType":"address[]","name":"path","type":"address[]"}
	],
	"name":"getAmountsOut",
	"outputs": [{"internalType":"uint256[]","name":"amounts","type":"uint256[]"}],
	"stateMutability":"view",
	"type":"function"
}]`

// 代币精度映射 (已知代币)
var tokenDecimals = map[string]int{
	strings.ToLower(WBNBAddress):    18,
	strings.ToLower(BUSDAddress):    18,
	strings.ToLower(USDTAddress):    18,
	strings.ToLower(USDCBSCAddress): 18, // USDC在BSC上是18位小数
	strings.ToLower(CAKEAddress):    18,
	strings.ToLower(WETHAddress):    18,
}

// PancakeTokenPrice 表示PancakeSwap代币价格信息
type PancakeTokenPrice struct {
	Symbol     string    `json:"symbol"`
	Address    string    `json:"address"`
	USDPrice   float64   `json:"usd_price"`
	BNBPrice   float64   `json:"bnb_price"`
	LastUpdate time.Time `json:"last_update"`
}

// TokenConfig 表示要监控的代币配置
type TokenConfig struct {
	Symbol  string
	Address string
}

// GetPancakeTokenPrice 获取代币在PancakeSwap上的价格
func GetPancakeTokenPrice(tokenAddress string, symbol string) (*PancakeTokenPrice, error) {
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address cannot be empty")
	}

	// 初始化以太坊客户端
	client, err := ethclient.Dial(BSCRPCURL)
	if err != nil {
		return nil, fmt.Errorf("连接BSC节点失败: %v", err)
	}
	defer client.Close()

	// 如果代币是BNB本身，获取BNB对USDT的价格
	if strings.EqualFold(tokenAddress, WBNBAddress) {
		return getPancakeBNBPrice(client, symbol)
	}

	// 优先使用USDT获取代币价格（流动性最稳定）
	usdPrice, err := getTokenPrice(client, tokenAddress, USDTAddress, DefaultAmountIn)
	if err != nil {
		// 如果USDT路径失败，尝试USDC路径
		usdPrice, err = getTokenPrice(client, tokenAddress, USDCBSCAddress, DefaultAmountIn)
		if err != nil {
			// 如果USDC路径也失败，尝试BUSD路径（向后兼容）
			usdPrice, err = getTokenPrice(client, tokenAddress, BUSDAddress, DefaultAmountIn)
			if err != nil {
				return nil, fmt.Errorf("获取USD价格失败: %v", err)
			}
		}
	}

	// 保存到数据库（不再查询BNB价格）
	SavePancakePriceToDb(tokenAddress, symbol, usdPrice, 0)

	return &PancakeTokenPrice{
		Symbol:     symbol,
		Address:    tokenAddress,
		USDPrice:   usdPrice,
		BNBPrice:   0, // 不再提供BNB价格
		LastUpdate: time.Now(),
	}, nil
}

// getPancakeBNBPrice 获取BNB在PancakeSwap上的价格
func getPancakeBNBPrice(client *ethclient.Client, symbol string) (*PancakeTokenPrice, error) {
	// 优先使用USDT获取BNB价格（流动性最稳定）
	usdPrice, err := getTokenPrice(client, WBNBAddress, USDTAddress, DefaultAmountIn)
	if err != nil {
		// 如果USDT路径失败，尝试USDC路径
		usdPrice, err = getTokenPrice(client, WBNBAddress, USDCBSCAddress, DefaultAmountIn)
		if err != nil {
			// 如果USDC路径也失败，尝试BUSD路径（向后兼容）
			usdPrice, err = getTokenPrice(client, WBNBAddress, BUSDAddress, DefaultAmountIn)
			if err != nil {
				return nil, fmt.Errorf("获取BNB价格失败: %v", err)
			}
		}
	}

	// 保存到数据库
	SavePancakePriceToDb(WBNBAddress, symbol, usdPrice, 0)

	return &PancakeTokenPrice{
		Symbol:     symbol,
		Address:    WBNBAddress,
		USDPrice:   usdPrice,
		BNBPrice:   0, // 不再提供BNB价格
		LastUpdate: time.Now(),
	}, nil
}

// getTokenPrice 获取代币价格 (返回单个代币的价格)
func getTokenPrice(client *ethclient.Client, tokenIn, tokenOut string, amount float64) (float64, error) {
	// 1. 获取代币精度
	inDecimals, err := getTokenDecimals(client, tokenIn)
	if err != nil {
		return 0, fmt.Errorf("获取输入代币精度失败: %w", err)
	}

	outDecimals, err := getTokenDecimals(client, tokenOut)
	if err != nil {
		return 0, fmt.Errorf("获取输出代币精度失败: %w", err)
	}

	// 2. 转换输入金额为wei单位
	amountInWei := toWei(amount, inDecimals)

	// 3. 首先尝试直接路由
	path := []common.Address{
		common.HexToAddress(tokenIn),
		common.HexToAddress(tokenOut),
	}

	// 4. 调用合约
	amountsOut, err := callGetAmountsOut(client, amountInWei, path)
	if err != nil {
		// 如果直接路由失败，且不是WBNB相关的交易，尝试通过WBNB路由
		if !strings.EqualFold(tokenIn, WBNBAddress) && !strings.EqualFold(tokenOut, WBNBAddress) {
			// 尝试通过WBNB的多跳路由: tokenIn -> WBNB -> tokenOut
			pathWithWBNB := []common.Address{
				common.HexToAddress(tokenIn),
				common.HexToAddress(WBNBAddress),
				common.HexToAddress(tokenOut),
			}
			amountsOut, err = callGetAmountsOut(client, amountInWei, pathWithWBNB)
			if err != nil {
				return 0, fmt.Errorf("合约调用失败(直接路由和WBNB路由都失败): %w", err)
			}
		} else {
			return 0, fmt.Errorf("合约调用失败: %w", err)
		}
	}

	// 5. 转换输出金额为可读格式
	if len(amountsOut) < 2 {
		return 0, fmt.Errorf("无效的输出数量")
	}
	// 使用最后一个金额（对于多跳路由，这是最终输出金额）
	amountOut := fromWei(amountsOut[len(amountsOut)-1], outDecimals)

	// 计算单个代币的价格（除以输入的代币数量）
	pricePerToken := amountOut / amount

	return pricePerToken, nil
}

// callGetAmountsOut 调用getAmountsOut合约方法
func callGetAmountsOut(client *ethclient.Client, amountIn *big.Int, path []common.Address) ([]*big.Int, error) {
	// 解析ABI
	contractABI, err := abi.JSON(strings.NewReader(pancakeRouterABI))
	if err != nil {
		return nil, err
	}

	// 打包调用数据
	data, err := contractABI.Pack("getAmountsOut", amountIn, path)
	if err != nil {
		return nil, err
	}

	// 创建调用消息
	routerAddr := common.HexToAddress(PancakeRouterAddress)
	callMsg := ethereum.CallMsg{
		To:   &routerAddr,
		Data: data,
	}

	// 执行调用
	result, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return nil, err
	}

	// 解析返回结果
	var amounts []*big.Int
	err = contractABI.UnpackIntoInterface(&amounts, "getAmountsOut", result)
	if err != nil {
		return nil, err
	}

	return amounts, nil
}

// getTokenDecimals 获取代币精度
func getTokenDecimals(client *ethclient.Client, tokenAddress string) (int, error) {
	// 首先检查已知代币
	addrLower := strings.ToLower(tokenAddress)
	if decimals, ok := tokenDecimals[addrLower]; ok {
		return decimals, nil
	}

	// 对于未知代币，调用合约获取精度
	decimalsABI := `[{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"}]`

	contractABI, err := abi.JSON(strings.NewReader(decimalsABI))
	if err != nil {
		return 0, err
	}

	// 方法ID (decimals())
	methodID := crypto.Keccak256([]byte("decimals()"))[:4]

	// 执行调用
	tokenAddr := common.HexToAddress(tokenAddress)
	callMsg := ethereum.CallMsg{
		To:   &tokenAddr,
		Data: methodID,
	}

	result, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		return 0, err
	}

	// 解析结果 (uint8)
	var decimals uint8
	err = contractABI.UnpackIntoInterface(&decimals, "decimals", result)
	if err != nil {
		return 0, err
	}

	// 缓存结果
	tokenDecimals[addrLower] = int(decimals)

	return int(decimals), nil
}

// toWei 转换为wei单位 (考虑代币精度)
func toWei(amount float64, decimals int) *big.Int {
	// 使用整数运算避免浮点数精度问题
	base := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	amountBig := new(big.Float).SetFloat64(amount)
	amountBig.Mul(amountBig, new(big.Float).SetInt(base))

	result := new(big.Int)
	amountBig.Int(result) // 截断小数部分
	return result
}

// fromWei 从wei单位转换 (考虑代币精度)
func fromWei(amount *big.Int, decimals int) float64 {
	// 创建10^decimals的大整数
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	// 转换为浮点数
	amountFloat := new(big.Float).SetInt(amount)
	divisorFloat := new(big.Float).SetInt(divisor)

	// 执行除法
	result, _ := new(big.Float).Quo(amountFloat, divisorFloat).Float64()
	return result
}

// MonitorPancakePrice 监控PancakeSwap上指定代币的价格
func MonitorPancakePrice(tokenAddress string, symbol string, interval time.Duration, stopCh <-chan struct{}) {
	log.Printf("开始监控 PancakeSwap 上 %s 的价格，间隔: %v", symbol, interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// 立即获取一次价格
	go func() {
		price, err := GetPancakeTokenPrice(tokenAddress, symbol)
		if err != nil {
			log.Printf("获取 %s 初始价格失败: %v", symbol, err)
		} else {
			log.Printf("%s 初始价格: $%.4f (%.8f BNB)", symbol, price.USDPrice, price.BNBPrice)
		}
	}()

	for {
		select {
		case <-ticker.C:
			go func() {
				price, err := GetPancakeTokenPrice(tokenAddress, symbol)
				if err != nil {
					log.Printf("获取 %s 价格失败: %v", symbol, err)
				} else {
					log.Printf("%s 当前价格: $%.4f (%.8f BNB)", symbol, price.USDPrice, price.BNBPrice)
				}
			}()
		case <-stopCh:
			log.Printf("停止监控 PancakeSwap 上 %s 的价格", symbol)
			return
		}
	}
}

// MonitorMultipleTokens 同时监控多个代币的价格
func MonitorMultipleTokens(tokens []TokenConfig, interval time.Duration) chan struct{} {
	log.Printf("开始监控 %d 个代币在 PancakeSwap 上的价格，间隔: %v", len(tokens), interval)

	// 创建一个全局停止信号通道
	stopCh := make(chan struct{})

	// 为每个代币启动一个监控协程
	for _, token := range tokens {
		go MonitorPancakePrice(token.Address, token.Symbol, interval, stopCh)
	}

	return stopCh
}

// PancakeSwapExample 展示如何使用PancakeSwap价格模块的示例
func PancakeSwapExample() {
	// 示例1: 获取单个代币价格
	log.Println("示例1: 获取CAKE代币价格")
	cakePrice, err := GetPancakeTokenPrice(CAKEAddress, "CAKE")
	if err != nil {
		log.Printf("获取CAKE价格失败: %v\n", err)
	} else {
		log.Printf("CAKE价格: $%.4f (%.8f BNB)\n", cakePrice.USDPrice, cakePrice.BNBPrice)
	}

	// 示例2: 获取BNB价格
	log.Println("示例2: 获取BNB价格")
	bnbPrice, err := GetPancakeTokenPrice(WBNBAddress, "BNB")
	if err != nil {
		log.Printf("获取BNB价格失败: %v\n", err)
	} else {
		log.Printf("BNB价格: $%.2f\n", bnbPrice.USDPrice)
	}

	// 示例3: 监控多个代币价格
	log.Println("示例3: 监控多个代币价格")
	tokens := []TokenConfig{
		{Symbol: "BNB", Address: WBNBAddress},
		{Symbol: "CAKE", Address: CAKEAddress},
		{Symbol: "BUSD", Address: BUSDAddress},
	}

	// 每30秒更新一次价格
	stopCh := MonitorMultipleTokens(tokens, 30*time.Second)

	// 运行5分钟后停止
	log.Println("将在5分钟后停止监控")
	time.Sleep(5 * time.Minute)
	close(stopCh)
	log.Println("已停止所有监控")
}

// SavePancakePriceToDb 将PancakeSwap价格数据保存到数据库
func SavePancakePriceToDb(tokenAddress string, symbol string, usdPrice float64, bnbPrice float64) {
	// 创建价格数据对象
	priceData := map[string]interface{}{
		"address":     tokenAddress,
		"symbol":      symbol,
		"usd_price":   usdPrice,
		"bnb_price":   bnbPrice,
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
			Decimals:        18,     // BSC上大多数代币是18位小数
			ContractAddress: tokenAddress,
			Chain:           "bsc", // BSC链
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
		// 对于PancakeSwap，我们可以更新合约地址和链信息
		if currenc.ContractAddress == "" && tokenAddress != "" {
			currenc.ContractAddress = tokenAddress
		}
		if currenc.Chain == "" {
			currenc.Chain = "bsc"
		}

		err = currency.UpdateCurrency(ctx, currenc)
		if err != nil {
			log.Printf("更新货币信息失败: %v", err)
		} else {
			log.Printf("已更新货币 %s 信息", currenc.Symbol)
		}
	}

	// 获取交易对ID
	tradingPairID := 3 // 默认值，假设PancakeSwap的exchange_id为3
	if currenc != nil && currenc.ID > 0 {
		tradingPairID = currenc.ID
	}

	// 创建 MarketData 对象
	marketData := &sqlite3.MarketData{
		ExchangeID:    3,                    // 假设 PancakeSwap 的 exchange_id 为 3，实际应该从配置或数据库中获取
		TradingPairID: int64(tradingPairID), // 使用货币ID作为交易对ID
		Timestamp:     time.Now(),
		LastPrice:     usdPrice,
		BidPrice:      0, // PancakeSwap API 不提供这些数据
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
		log.Printf("保存PancakeSwap市场数据到数据库失败: %v", err)
	} else {
		log.Printf("已保存 %s PancakeSwap市场数据到数据库，ID: %d", symbol, marketData.ID)
	}
}
