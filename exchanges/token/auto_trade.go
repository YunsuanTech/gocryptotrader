package token

// import (
// 	"fmt"
// 	"log"
// 	"math"
// 	"strconv"
// 	"sync"
// 	"time"

// 	"gocryptotrader/config"
// )

// // AutoTradeManager 管理自动交易相关操作
// type AutoTradeManager struct {
// 	config         *config.Config
// 	watchlistMgr   *WatchlistManager
// 	tradingMgr     *TradingManager
// 	privateKey     string
// 	running        bool
// 	checkInterval  time.Duration
// 	executionLock  sync.Mutex
// 	processedRules map[int]time.Time
// 	stopChan       chan struct{}
// }

// // AutoTradeConfig 自动交易配置
// type AutoTradeConfig struct {
// 	PrivateKey    string        // Solana钱包私钥
// 	CheckInterval time.Duration // 检查交易规则的时间间隔
// }

// // DefaultAutoTradeConfig 返回默认的自动交易配置
// func DefaultAutoTradeConfig() *AutoTradeConfig {
// 	return &AutoTradeConfig{
// 		PrivateKey:    "", // 需要用户提供
// 		CheckInterval: 1 * time.Minute,
// 	}
// }

// // NewAutoTradeManager 创建一个新的自动交易管理器
// func NewAutoTradeManager(cfg *config.Config, autoConfig *AutoTradeConfig) *AutoTradeManager {
// 	if autoConfig == nil {
// 		autoConfig = DefaultAutoTradeConfig()
// 	}

// 	return &AutoTradeManager{
// 		config:         cfg,
// 		watchlistMgr:   NewWatchlistManager(cfg),
// 		tradingMgr:     NewTradingManager(cfg),
// 		privateKey:     autoConfig.PrivateKey,
// 		checkInterval:  autoConfig.CheckInterval,
// 		processedRules: make(map[int]time.Time),
// 		stopChan:       make(chan struct{}),
// 	}
// }

// // SetPrivateKey 设置私钥
// func (m *AutoTradeManager) SetPrivateKey(privateKey string) {
// 	m.privateKey = privateKey
// }

// // Start 启动自动交易服务
// func (m *AutoTradeManager) Start() error {
// 	if m.privateKey == "" {
// 		return fmt.Errorf("未设置私钥，无法启动自动交易服务")
// 	}

// 	m.executionLock.Lock()
// 	defer m.executionLock.Unlock()

// 	if m.running {
// 		return fmt.Errorf("自动交易服务已在运行中")
// 	}

// 	m.running = true
// 	m.stopChan = make(chan struct{})

// 	go m.monitorLoop()

// 	log.Println("自动交易服务已启动")
// 	return nil
// }

// // Stop 停止自动交易服务
// func (m *AutoTradeManager) Stop() {
// 	m.executionLock.Lock()
// 	defer m.executionLock.Unlock()

// 	if !m.running {
// 		return
// 	}

// 	close(m.stopChan)
// 	m.running = false
// 	log.Println("自动交易服务已停止")
// }

// // monitorLoop 监控循环，定期检查交易规则
// func (m *AutoTradeManager) monitorLoop() {
// 	ticker := time.NewTicker(m.checkInterval)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-m.stopChan:
// 			return
// 		case <-ticker.C:
// 			if err := m.checkAndExecuteTrades(); err != nil {
// 				log.Printf("执行自动交易检查时出错: %v\n", err)
// 			}
// 		}
// 	}
// }

// // checkAndExecuteTrades 检查并执行符合条件的交易
// func (m *AutoTradeManager) checkAndExecuteTrades() error {
// 	// 获取活跃的代币列表
// 	activeTokens, err := m.watchlistMgr.GetActiveWatchlistTokens()
// 	if err != nil {
// 		return fmt.Errorf("获取活跃代币列表失败: %w", err)
// 	}

// 	// 获取活跃的交易规则
// 	activeRules, err := m.tradingMgr.GetActiveTradingRules()
// 	if err != nil {
// 		return fmt.Errorf("获取活跃交易规则失败: %w", err)
// 	}

// 	// 为每个代币创建一个映射，以便快速查找
// 	tokenMap := make(map[int]*WatchlistToken)
// 	for _, token := range activeTokens {
// 		tokenMap[token.ID] = token
// 	}

// 	// 检查每个交易规则
// 	for _, rule := range activeRules {
// 		// 跳过已过期的规则
// 		if rule.ExpirationTime > 0 && time.Now().Unix() > rule.ExpirationTime {
// 			continue
// 		}

// 		// 获取对应的代币信息
// 		token, exists := tokenMap[rule.TokenID]
// 		if !exists {
// 			log.Printf("警告: 找不到交易规则 %d 对应的代币 ID %d\n", rule.RuleID, rule.TokenID)
// 			continue
// 		}

// 		// 检查网络是否为Solana
// 		if token.Network != "solana" {
// 			log.Printf("警告: 代币 %s 不在Solana网络上，跳过\n", token.TokenSymbol)
// 			continue
// 		}

// 		// 检查是否最近处理过该规则（防止重复执行）
// 		if lastProcessed, ok := m.processedRules[rule.RuleID]; ok {
// 			if time.Since(lastProcessed) < 5*time.Minute {
// 				continue // 跳过最近5分钟内处理过的规则
// 			}
// 		}

// 		// 获取当前价格
// 		price, err := GetTokenPrice(token.TokenAddress)
// 		if err != nil {
// 			log.Printf("获取代币 %s 价格失败: %v\n", token.TokenSymbol, err)
// 			continue
// 		}

// 		// 检查价格触发条件
// 		if m.shouldExecuteTrade(rule, price.USDPrice) {
// 			// 执行交易
// 			if err := m.executeTrade(rule, token, price); err != nil {
// 				log.Printf("执行交易失败 (规则ID: %d, 代币: %s): %v\n",
// 					rule.RuleID, token.TokenSymbol, err)
// 			} else {
// 				// 记录处理时间
// 				m.processedRules[rule.RuleID] = time.Now()

// 				// 更新规则的最后触发时间
// 				m.updateRuleLastTriggered(rule.RuleID)
// 			}
// 		}
// 	}

// 	return nil
// }

// // shouldExecuteTrade 判断是否应该执行交易
// func (m *AutoTradeManager) shouldExecuteTrade(rule *TradingRule, currentPrice float64) bool {
// 	switch rule.Direction {
// 	case "BUY":
// 		// 当前价格低于或等于触发价格时买入
// 		return currentPrice <= rule.TriggerPrice
// 	case "SELL":
// 		// 当前价格高于或等于触发价格时卖出
// 		return currentPrice >= rule.TriggerPrice
// 	default:
// 		log.Printf("未知的交易方向: %s\n", rule.Direction)
// 		return false
// 	}
// }

// // executeTrade 执行交易
// func (m *AutoTradeManager) executeTrade(rule *TradingRule, token *WatchlistToken, price *TokenPrice) error {
// 	// 准备交易配置
// 	config := DefaultSolanaSwapConfig()

// 	// 设置交易参数
// 	config.Slippage = rule.Slippage
// 	config.FromAddress = rule.UserAddress

// 	// 根据交易方向设置输入和输出代币
// 	if rule.Direction == "BUY" {
// 		// 买入: SOL -> Token
// 		config.InputToken = SolAddress
// 		config.OutputToken = token.TokenAddress

// 		log.Printf("USD价格: %.4f, SOL价格: %.4f, 数量: %.4f", price.USDPrice, price.SOLPrice, rule.Quantity)
// 		// 计算SOL数量 (根据规则中的数量和当前价格)
// 		solAmount := rule.Quantity / price.SOLPrice
// 		log.Printf("SOL数量: %.4f", solAmount)
// 		config.Amount = convertToLamports(solAmount)
// 	} else {
// 		// 卖出: Token -> SOL
// 		config.InputToken = token.TokenAddress
// 		config.OutputToken = SolAddress

// 		// 使用规则中指定的代币数量
// 		tokenAmount := rule.Quantity
// 		config.Amount = convertToTokenAmount(tokenAmount, token.Decimals)
// 	}

// 	// 执行交易
// 	result, err := ExecuteSolanaSwap(m.privateKey, config)
// 	if err != nil {
// 		return fmt.Errorf("执行Solana交换失败: %w", err)
// 	}

// 	// 记录交易结果
// 	log.Printf("交易成功执行 (规则ID: %d, 代币: %s, 方向: %s, 交易哈希: %s)\n",
// 		rule.RuleID, token.TokenSymbol, rule.Direction, result.TransactionHash)

// 	return nil
// }

// // updateRuleLastTriggered 更新规则的最后触发时间
// func (m *AutoTradeManager) updateRuleLastTriggered(ruleID int) {
// 	// 获取现有规则
// 	rule, err := m.tradingMgr.GetTradingRuleByID(ruleID)
// 	if err != nil {
// 		log.Printf("获取交易规则失败: %v\n", err)
// 		return
// 	}

// 	// 更新规则，保持其他字段不变，只更新LastTriggered
// 	err = m.tradingMgr.UpdateTradingRule(
// 		rule.RuleID,
// 		rule.TokenID,
// 		rule.UserAddress,
// 		rule.Direction,
// 		rule.TriggerPrice,
// 		rule.Quantity,
// 		rule.Slippage,
// 		rule.ExpirationTime,
// 		rule.IsEnabled,
// 		rule.OrderType,
// 	)

// 	if err != nil {
// 		log.Printf("更新交易规则失败: %v\n", err)
// 	}
// }

// // convertToLamports 将SOL数量转换为lamports（SOL的最小单位，1 SOL = 10^9 lamports）
// func convertToLamports(solAmount float64) string {
// 	lamports := int64(solAmount * 1e9)
// 	return strconv.FormatInt(lamports, 10)
// }

// // convertToTokenAmount 将代币数量转换为代币的最小单位
// func convertToTokenAmount(amount float64, decimals int) string {
// 	tokenAmount := int64(amount * math.Pow10(decimals))
// 	return strconv.FormatInt(tokenAmount, 10)
// }
