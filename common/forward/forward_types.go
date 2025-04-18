package forward

// Config 定义 SOL 转发的配置参数
type Config struct {
	RPCEndpoint   string // Solana RPC 端点
	ConcurrentTxs int    // 并发交易数量

}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		RPCEndpoint:   "http://xolana.xen.network:8899",
		ConcurrentTxs: 5,
	}
}

// ForwardRequest 定义转发请求的结构
type ForwardRequest struct {
	PrivateKeyStr string   // 发送者私钥
	Addresses     []string // 接收者地址列表
	Config        *Config  // 转发配置
	AmountSOL     float64  // 每笔转账的 SOL 数量
}

// TokenForwardRequest 定义代币转发请求的结构
type TokenForwardRequest struct {
	PrivateKeyStr string   // 发送者私钥
	TokenMint     string   // 代币铸币账户地址
	Addresses     []string // 接收者地址列表
	IsToken2022   bool     // 是否为Token-2022类型代币
	Config        *Config  // 转发配置
	Amount        float64  // 每笔转账的代币数量
}
