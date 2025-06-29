# Exchanges Handling Module

这个模块包含了各种交易所和价格数据源的处理逻辑。

## 文件结构

- `constants.go` - 公共常量定义文件
- `binance_price.go` - Binance价格数据处理
- `coingecko_price.go` - CoinGecko价格数据处理
- `gmgn_price.go` - GMGN多链价格数据处理
- `pancakeswap_price.go` - PancakeSwap价格数据处理
- `uniswap_price.go` - Uniswap价格数据处理
- `price_monitoring.go` - 价格监控功能

## 常量整理

为了减少代码重复和提高维护性，所有重复的常量已经被提取到 `constants.go` 文件中：

### 代币地址常量
- `ETHAddress` - 以太坊地址
- `WETHAddress` - Wrapped ETH地址
- `USDCAddress` - USDC地址
- `WBNBAddress` - Wrapped BNB地址
- `BUSDAddress` - BUSD地址
- `USDTAddress` - USDT地址
- `CAKEAddress` - CAKE代币地址
- `SOLAddress` - Solana地址
- `USDAddress` - USD地址

### RPC URL常量
- `BSCRPCURL` - BSC RPC URL
- `EthereumRPCURL` - 以太坊RPC URL

### 合约地址常量
- `PancakeRouterAddress` - PancakeSwap V2 Router地址

### 默认值常量
- `CommonDefaultTimeout` - 通用默认超时时间
- `DefaultAmountIn` - 默认输入金额
- `DefaultSlippage` - 默认滑点
- `DefaultFee` - 默认费用
- `CommonUniswapDefaultFee` - Uniswap默认费用

### 默认金额常量
- `DefaultSOLAmount` - 默认SOL金额
- `DefaultTokenAmount` - 默认代币金额
- `DefaultETHAmount` - 默认ETH金额
- `DefaultSOLFromAddress` - 默认SOL发送地址

### API基础URL和端点
- `GMGNBaseURL` - GMGN API基础URL
- `ExactInEndpoint` - 精确输入端点
- `ExactOutEndpoint` - 精确输出端点
- `CoinGeckoBaseURL` - CoinGecko API基础URL
- `CommonPriceEndpoint` - 通用价格端点
- `BinanceWSBaseURL` - Binance WebSocket基础URL
- `BinanceTickerPath` - Binance Ticker路径

## 使用说明

所有模块现在都使用 `constants.go` 中定义的公共常量，这样可以：

1. **减少重复代码** - 避免在多个文件中重复定义相同的常量
2. **提高维护性** - 只需在一个地方更新常量值
3. **保持一致性** - 确保所有模块使用相同的配置值
4. **便于管理** - 集中管理所有配置常量

## 注意事项

- 修改 `constants.go` 中的常量时，请确保所有依赖模块都能正常工作
- 添加新常量时，请遵循现有的命名规范
- 如果某个常量只在单个模块中使用，可以考虑保留在该模块的本地常量中