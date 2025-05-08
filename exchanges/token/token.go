package token

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"gocryptotrader/database/models/sqlite3"
	tokenmonitor "gocryptotrader/database/repository/token_monitor"
	tradingrule "gocryptotrader/database/repository/trading_rule"
	transactionrecord "gocryptotrader/database/repository/transaction_record"
)

// BuyRule 表示买入规则的结构体
type BuyRule struct {
	ID          int
	RuleName    string
	Condition   string
	BuyPrice    float64
	SellPrice   float64
	Priority    float64
	Description string
}

// BuySOLToken 实现买入SOL token的功能
func BuySOLToken(fetcher *GMGNFetcher) error {

	ctx := context.Background()
	// 获取所有token信息
	tokens, err := fetcher.GetLatestCompletedTokens()
	if err != nil {
		return fmt.Errorf("获取GMGN代币信息失败: %v", err)
	}

	// 获取买入规则
	rules, err := getBuyRules("buy")
	if err != nil {
		return fmt.Errorf("获取买入规则失败: %v", err)
	}

	// 遍历所有token，检查是否符合买入规则
	for _, token := range tokens {
		// 检查每个规则
		for _, rule := range rules {
			if matchBuyRuleGMGN(token, rule) {
				// 查询该token是否已存在于监控表中，若已存在则跳过
				monitors, err := tokenmonitor.QueryTokenMonitors(tokenmonitor.TokenMonitorQueryOptions{TokenAddress: token.Address, Limit: 1, IsMonitoring: -1})
				if err != nil {
					log.Printf("[买入] 查询TokenMonitor失败: %v", err)
					continue
				}

				if reflect.ValueOf(monitors).Kind() == reflect.Slice && reflect.ValueOf(monitors).Len() != 0 {
					continue
				}

				// 将价格字符串转换为float64
				priceFloat, err := strconv.ParseFloat(token.Price, 64)
				if err != nil {
					log.Printf("[买入] 价格转换失败: %v", err)
					continue
				}

				// 使用醒目标记突出重要买入信息
				log.Printf("【买入信号】Token %s 符合买入规则 '%s'", token.Symbol, rule.RuleName)
				log.Printf("【买入详情】市值: %.2f | 价格: %.6f | 规则: %s",
					token.MarketCap, priceFloat, rule.Description)

				// 创建新的TokenMonitor记录
				tokenMonitor := &sqlite3.TokenMonitor{
					TokenAddress:  token.Address,
					TokenName:     token.Symbol,
					Price:         priceFloat,
					TokenDecimals: 9,                          // 默认为9，Solana代币通常为9位小数
					BuyAmount:     rule.BuyPrice / priceFloat, // 模拟结果
					Amount:        rule.BuyPrice / priceFloat, // 模拟结果
					BuyPrice:      priceFloat,
					BuyTime:       time.Now().Unix(),
					IsMonitoring:  1,
				}

				// 插入记录到数据库
				if err := tokenmonitor.InsertTokenMonitor(ctx, tokenMonitor); err != nil {
					log.Printf("[买入] 插入TokenMonitor记录失败: %v", err)
					continue
				}

				// 创建交易记录
				transactionRecord := &sqlite3.TransactionRecord{
					TokenAddress: token.Address,
					RuleID:       rule.ID,
					Type:         "buy",
					Amount:       tokenMonitor.Amount,
					Price:        tokenMonitor.BuyPrice,
					Timestamp:    tokenMonitor.BuyTime,
					TxHash:       token.Address, // 模拟交易，暂不设置实际的交易哈希
					Status:       "confirmed",
				}

				// 插入交易记录
				if err := transactionrecord.InsertTransactionRecord(ctx, transactionRecord); err != nil {
					log.Printf("[买入] 插入交易记录失败: %v", err)
				}

				log.Printf("【买入成功】添加Token %s到监控列表，买入价格: %.6f，买入数量: %.6f",
					token.Symbol, tokenMonitor.BuyPrice, tokenMonitor.Amount)

				log.Printf("================================================================")
			}
		}
	}

	return nil
}

func SellSOLToken(fetcher *GMGNFetcher) error {
	ctx := context.Background()
	// 获取所有tokenmonitor记录
	monitorsInterface, err := tokenmonitor.QueryTokenMonitors(tokenmonitor.TokenMonitorQueryOptions{IsMonitoring: 1})
	if err != nil {
		log.Printf("[卖出] 查询TokenMonitor失败: %v", err)
		return err
	}

	// 获取卖出规则
	rules, err := getBuyRules("sell")
	if err != nil {
		return fmt.Errorf("获取卖出规则失败: %v", err)
	}

	// 类型断言，将 interface{} 转换为 sqlite3.TokenMonitorSlice
	monitors, ok := monitorsInterface.(sqlite3.TokenMonitorSlice)
	if !ok {
		return fmt.Errorf("monitorsInterface 不是 TokenMonitorSlice 类型")
	}

	// 遍历所有tokenmonitor记录
	for _, monitor := range monitors {

		// 获取最新价格
		tokenPrice, err := GetTokenPrice(monitor.TokenAddress)
		if err != nil {
			log.Printf("[卖出] 获取Token %s 价格失败: %v", monitor.TokenName, err)
			continue
		}

		// 更新价格
		monitor.Price = tokenPrice.USDPrice

		// 计算涨幅 (当前价格 - 买入价格) / 买入价格
		monitor.Increase = (monitor.Price - monitor.BuyPrice) / monitor.BuyPrice

		// 检查是否符合卖出规则
		for _, rule := range rules {
			// 解析规则条件
			if matchSellRule(monitor, rule) {
				// 使用醒目标记突出重要卖出信息
				log.Printf("【卖出信号】Token %s 符合卖出规则 '%s'", monitor.TokenName, rule.RuleName)
				log.Printf("【卖出详情】买入价格: %.6f USD 当前价格: %.6f USD | 涨幅: %.2f%% | 规则: %s",
					monitor.BuyPrice, monitor.Price, monitor.Increase*100, rule.Description)

				// 计算卖出数量
				sellAmount := monitor.Amount * rule.SellPrice

				// 计算卖出金额
				sellValue := sellAmount * monitor.Price

				log.Printf("[卖出] 卖出数量: %.6f, 卖出金额: %.2f USD", sellAmount, sellValue)

				// 更新token监控记录
				monitor.Amount -= sellAmount
				// 计算卖出百分比：(买入数量 - 当前数量) / 买入数量
				monitor.SellPercentage = (monitor.BuyAmount - monitor.Amount) / monitor.BuyAmount
				monitor.TotalSellPrice += sellValue
				monitor.LastSellTime = float64(time.Now().Unix())

				// 如果全部卖出，则停止监控
				if monitor.Amount <= 0 {
					monitor.IsMonitoring = 0
					log.Printf("【卖出完成】Token %s 已全部卖出，停止监控", monitor.TokenName)
				}

				// 创建交易记录
				transactionRecord := &sqlite3.TransactionRecord{
					TokenAddress: monitor.TokenAddress,
					RuleID:       rule.ID,
					Type:         "sell",
					Amount:       sellAmount,
					Price:        monitor.Price,
					Timestamp:    int64(monitor.LastSellTime),
					TxHash:       monitor.TokenAddress + "12", // 模拟交易，暂不设置实际的交易哈希
					Status:       "confirmed",
				}

				// 插入交易记录
				if err := transactionrecord.InsertTransactionRecord(ctx, transactionRecord); err != nil {
					log.Printf("[卖出] 插入交易记录失败: %v", err)
				}

				// 更新数据库记录
				if err := tokenmonitor.UpdateTokenMonitor(context.Background(), monitor); err != nil {
					log.Printf("[卖出] 更新Token %s 监控记录失败: %v", monitor.TokenName, err)
					continue
				}

				log.Printf("【卖出成功】Token %s 卖出金额: %.2f USD，剩余数量: %.6f",
					monitor.TokenName, sellValue, monitor.Amount)
				log.Printf("================================================================")
				break // 一个token只应用一条卖出规则

			}
		}

		// 无论是否卖出，都更新价格和涨幅信息
		if err := tokenmonitor.UpdateTokenMonitor(context.Background(), monitor); err != nil {
			log.Printf("[卖出] 更新Token %s 监控记录失败: %v", monitor.TokenName, err)
			continue
		}
	}

	return nil
}

// getBuyRules 获取所有买卖规则
func getBuyRules(ruleType string) ([]BuyRule, error) {
	opts := tradingrule.QueryOptions{
		RuleType: ruleType,
	}

	result, err := tradingrule.QueryTradingRules(opts)
	if err != nil {
		return nil, err
	}

	// 将结果转换为TradingRuleSlice类型
	tradingRules, ok := result.(sqlite3.TradingRuleSlice)
	if !ok {
		return nil, fmt.Errorf("无法将结果转换为TradingRuleSlice类型")
	}

	// 将TradingRuleSlice转换为BuyRule切片
	rules := make([]BuyRule, 0, len(tradingRules))
	for _, tr := range tradingRules {
		buyRule := BuyRule{
			ID:          tr.ID,
			RuleName:    tr.RuleName,
			Condition:   tr.Condition,
			BuyPrice:    tr.BuyPrice,
			SellPrice:   tr.SellPrice,
			Priority:    tr.Priority,
			Description: tr.Description,
		}
		rules = append(rules, buyRule)
	}

	return rules, nil
}

// getFieldValueByNameGMGN 通过字段名获取GMGNToken结构体中对应字段的值
func getFieldValueByNameGMGN(token GMGNToken, fieldName string) (interface{}, error) {
	// 使用反射获取结构体的值和类型
	val := reflect.ValueOf(token)
	typ := reflect.TypeOf(token)

	// 创建字段名到结构体字段的映射
	fieldMap := make(map[string]int)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		// 优先使用标签中定义的字段名
		tagName := field.Tag.Get("json")
		if tagName != "" {
			// 去除json标签中可能的逗号及后面的内容
			tagParts := strings.Split(tagName, ",")
			fieldMap[strings.ToLower(tagParts[0])] = i
		}
		// 同时支持直接使用结构体字段名（小写形式）
		fieldMap[strings.ToLower(field.Name)] = i
	}

	// // 处理特殊字段映射
	// switch strings.ToLower(fieldName) {
	// case "mcap":
	// 	fieldName = "market_cap"
	// case "price":
	// 	fieldName = "price"
	// case "holdercount":
	// 	fieldName = "holder_count"
	// }

	// 查找字段
	fieldIndex, ok := fieldMap[strings.ToLower(fieldName)]

	if !ok {
		return nil, fmt.Errorf("未找到字段: %s", fieldName)
	}

	// 获取字段值
	fieldValue := val.Field(fieldIndex)

	// 特殊处理price字段，将字符串转换为float64
	if strings.ToLower(fieldName) == "price" && fieldValue.Kind() == reflect.String {
		priceStr := fieldValue.String()
		priceFloat, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, fmt.Errorf("价格转换失败: %v", err)
		}
		return priceFloat, nil
	}

	// 返回字段值
	return fieldValue.Interface(), nil
}

// compareValues 比较两个值是否满足指定的操作符条件
// 支持字符串和数值类型的比较，字符串会尝试先转换为数值进行比较
// 如果转换失败，则按字符串的自然排序进行比较
// 支持的操作符: >, <, >=, <=, =, !=
func compareValues(fieldValue interface{}, operator string, condValue string) (bool, error) {
	switch v := fieldValue.(type) {
	case string:
		// 尝试将字符串转换为数值进行比较
		fieldFloat, fieldErr := strconv.ParseFloat(v, 64)
		condFloat, condErr := strconv.ParseFloat(condValue, 64)

		// 如果两个值都可以转换为数值，则按数值比较
		if fieldErr == nil && condErr == nil {
			switch operator {
			case ">":
				return fieldFloat > condFloat, nil
			case "<":
				return fieldFloat < condFloat, nil
			case ">=":
				return fieldFloat >= condFloat, nil
			case "<=":
				return fieldFloat <= condFloat, nil
			case "=":
				return fieldFloat == condFloat, nil
			case "!=":
				return fieldFloat != condFloat, nil
			default:
				return false, fmt.Errorf("未知的操作符: %s", operator)
			}
		}

		// 如果不能转换为数值，则按字符串比较
		switch operator {
		case ">":
			return v > condValue, nil
		case "<":
			return v < condValue, nil
		case ">=":
			return v >= condValue, nil
		case "<=":
			return v <= condValue, nil
		case "=":
			return v == condValue, nil
		case "!=":
			return v != condValue, nil
		default:
			return false, fmt.Errorf("未知的操作符: %s", operator)
		}

	case float64:
		// 数值类型支持比较操作符
		numValue, err := strconv.ParseFloat(condValue, 64)
		if err != nil {
			return false, fmt.Errorf("条件值解析错误: %s", condValue)
		}

		switch operator {
		case ">":
			return v > numValue, nil
		case "<":
			return v < numValue, nil
		case ">=":
			return v >= numValue, nil
		case "<=":
			return v <= numValue, nil
		case "=":
			return v == numValue, nil
		case "!=":
			return v != numValue, nil
		default:
			return false, fmt.Errorf("未知的操作符: %s", operator)
		}

	case int, int32, int64:
		// 整数类型转换为float64后比较
		var intValue float64
		switch iv := v.(type) {
		case int:
			intValue = float64(iv)
		case int32:
			intValue = float64(iv)
		case int64:
			intValue = float64(iv)
		}

		numValue, err := strconv.ParseFloat(condValue, 64)
		if err != nil {
			return false, fmt.Errorf("条件值解析错误: %s", condValue)
		}

		switch operator {
		case ">":
			return intValue > numValue, nil
		case "<":
			return intValue < numValue, nil
		case ">=":
			return intValue >= numValue, nil
		case "<=":
			return intValue <= numValue, nil
		case "=":
			return intValue == numValue, nil
		case "!=":
			return intValue != numValue, nil
		default:
			return false, fmt.Errorf("未知的操作符: %s", operator)
		}

	default:
		return false, fmt.Errorf("不支持的字段类型: %T", v)
	}
}

// matchBuyRuleGMGN 检查GMGNToken是否符合买入规则
func matchBuyRuleGMGN(token GMGNToken, rule BuyRule) bool {
	// 按and分割多个条件
	conditions := strings.Split(rule.Condition, " and ")

	// 检查所有条件是否都满足
	for _, condition := range conditions {
		condParts := strings.Split(strings.TrimSpace(condition), " ")
		if len(condParts) != 3 {
			log.Printf("规则条件格式错误: %s\n", condition)
			return false
		}

		field := condParts[0]
		operator := condParts[1]
		value := condParts[2]

		// 通过反射获取字段值
		fieldValue, err := getFieldValueByNameGMGN(token, field)
		if err != nil {
			log.Printf("%v\n", err)
			return false
		}

		// 比较值是否满足条件
		result, err := compareValues(fieldValue, operator, value)
		if err != nil {
			log.Printf("%v\n", err)
			return false
		}

		if !result {
			return false
		}
	}

	// 所有条件都满足
	return true
}

// matchSellRule 检查token是否符合卖出规则
func matchSellRule(monitor *sqlite3.TokenMonitor, rule BuyRule) bool {
	// 按and分割多个条件
	conditions := strings.Split(rule.Condition, " and ")

	// 检查所有条件是否都满足
	for _, condition := range conditions {
		condParts := strings.Split(strings.TrimSpace(condition), " ")
		if len(condParts) != 3 {
			log.Printf("规则条件格式错误: %s\n", condition)
			return false
		}

		field := condParts[0]
		operator := condParts[1]
		value := condParts[2]

		// 处理特殊字段
		var fieldValue interface{}
		var err error

		// 处理特殊字段：sell_percentage 和 increase
		if field == "sell_percentage" {
			fieldValue = monitor.SellPercentage
		} else if field == "increase" {
			fieldValue = monitor.Increase
		} else {
			log.Printf("卖出规则不支持的字段: %s\n", field)
			return false
		}

		// 比较值是否满足条件
		result, err := compareValues(fieldValue, operator, value)
		if err != nil {
			log.Printf("%v\n", err)
			return false
		}

		if !result {
			return false
		}
	}

	// 所有条件都满足
	return true
}

// TradeTokenBySignal 根据交易信号执行买入或卖出操作
// 输入代币地址、买入价格和交易比例，通过AnalyzeKlineSignal方法判断是否应该买入或卖出
// 根据信号结果创建相应的交易记录并保存到数据库中
// 支持多次买入卖出，不需要将数据添加到监控列表，只创建交易记录
// tradeRatio: 交易比例，范围0-1，表示买入或卖出的比例，默认为1（全部）
func TradeTokenBySignal(tokenAddress string, buyPrice float64, tradeRatio ...float64) error {
	ctx := context.Background()

	// 设置交易比例，默认为1（全部）
	ratio := 1.0
	if len(tradeRatio) > 0 && tradeRatio[0] > 0 && tradeRatio[0] <= 1 {
		ratio = tradeRatio[0]
	}

	// 获取交易信号
	signal, err := AnalyzeKlineSignal(tokenAddress)
	if err != nil {
		return fmt.Errorf("获取交易信号失败: %v", err)
	}

	// 获取最新价格
	tokenPrice, err := GetTokenPrice(tokenAddress)
	if err != nil {
		return fmt.Errorf("获取代币价格失败: %v", err)
	}

	// 根据信号执行相应操作
	switch signal {
	case SignalBuy:
		// 计算买入数量
		amount := (buyPrice * ratio) / tokenPrice.USDPrice

		// 获取当前持仓信息
		recordsInterface, err := transactionrecord.QueryTransactionRecords(transactionrecord.TransactionRecordQueryOptions{
			TokenAddress: tokenAddress,
			Status:       "confirmed",
		})

		var currentHoldAmount float64
		var avgBuyPrice float64
		var totalCost float64

		if err == nil {
			if records, ok := recordsInterface.(sqlite3.TransactionRecordSlice); ok {
				// 计算当前持有数量和总成本
				for _, record := range records {
					if record.Type == "buy" {
						currentHoldAmount += record.Amount
						totalCost += record.Amount * record.Price
					} else if record.Type == "sell" {
						currentHoldAmount -= record.Amount
						totalCost -= record.Amount * record.Price
					}
				}

				// 计算平均买入价格
				if currentHoldAmount > 0 {
					avgBuyPrice = totalCost / currentHoldAmount
				}
			}
		}

		// 使用醒目标记突出重要买入信息
		log.Printf("【买入信号】Token %s 符合买入信号", tokenAddress)
		log.Printf("【当前持仓】持有数量: %.6f | 平均买入价: %.6f USD", currentHoldAmount, avgBuyPrice)
		log.Printf("【买入详情】价格: %.6f USD | 买入数量: %.6f | 买入比例: %.2f%%",
			tokenPrice.USDPrice, amount, ratio*100)

		// 创建交易记录
		transactionRecord := &sqlite3.TransactionRecord{
			TokenAddress: tokenAddress,
			Type:         "buy",
			Amount:       amount,
			Price:        tokenPrice.USDPrice,
			Timestamp:    time.Now().Unix(),
			TxHash:       tokenAddress + "_" + strconv.FormatInt(time.Now().Unix(), 10), // 模拟交易哈希
			Status:       "confirmed",
		}

		// 插入交易记录
		if err := transactionrecord.InsertTransactionRecord(ctx, transactionRecord); err != nil {
			return fmt.Errorf("插入交易记录失败: %v", err)
		}

		// 更新持仓统计
		newHoldAmount := currentHoldAmount + amount
		newAvgPrice := (totalCost + (amount * tokenPrice.USDPrice)) / newHoldAmount
		increase := (tokenPrice.USDPrice - newAvgPrice) / newAvgPrice

		log.Printf("【买入成功】Token %s", tokenAddress)
		log.Printf("【最新持仓】持有数量: %.6f | 平均成本: %.6f USD | 当前涨幅: %.2f%%",
			newHoldAmount, newAvgPrice, increase*100)
		log.Printf("================================================================")

		// 更新监控记录（可选，不影响主要功能）
		monitorsInterface, err := tokenmonitor.QueryTokenMonitors(tokenmonitor.TokenMonitorQueryOptions{
			TokenAddress: tokenAddress,
			Limit:        1,
		})

		if err == nil {
			if monitors, ok := monitorsInterface.(sqlite3.TokenMonitorSlice); ok && len(monitors) > 0 {
				// 更新现有记录
				monitor := monitors[0]
				monitor.Amount = newHoldAmount
				monitor.BuyAmount += amount
				monitor.BuyPrice = newAvgPrice
				monitor.Price = tokenPrice.USDPrice
				monitor.Increase = increase
				_ = tokenmonitor.UpdateTokenMonitor(ctx, monitor)
			} else {
				// 创建新的监控记录
				tokenMonitor := &sqlite3.TokenMonitor{
					TokenAddress:  tokenAddress,
					TokenName:     tokenAddress[:8] + "...", // 简化显示
					Price:         tokenPrice.USDPrice,
					TokenDecimals: 9,
					BuyAmount:     amount,
					Amount:        amount,
					BuyPrice:      tokenPrice.USDPrice,
					BuyTime:       time.Now().Unix(),
					IsMonitoring:  0,
				}
				_ = tokenmonitor.InsertTokenMonitor(ctx, tokenMonitor)
			}
		}

	case SignalSell:
		// 获取该代币的交易记录，计算持有数量
		recordsInterface, err := transactionrecord.QueryTransactionRecords(transactionrecord.TransactionRecordQueryOptions{
			TokenAddress: tokenAddress,
			Status:       "confirmed",
		})
		if err != nil {
			return fmt.Errorf("查询交易记录失败: %v", err)
		}

		// 类型断言
		records, ok := recordsInterface.(sqlite3.TransactionRecordSlice)
		if !ok {
			return fmt.Errorf("交易记录类型断言失败")
		}

		// 创建交易记录的副本并按时间戳排序（FIFO顺序）
		sortedRecords := make([]sqlite3.TransactionRecord, len(records))
		for i, record := range records {
			sortedRecords[i] = *record
		}
		sort.Slice(sortedRecords, func(i, j int) bool {
			return sortedRecords[i].Timestamp < sortedRecords[j].Timestamp
		})

		// 计算当前持仓情况
		var (
			holdAmount     float64 // 当前持有数量
			totalCost      float64 // 总成本
			totalSold      float64 // 已卖出数量
			totalSoldValue float64 // 已卖出总价值
		)

		// 使用FIFO方法计算当前持仓
		for _, record := range sortedRecords {
			if record.Type == "buy" {
				holdAmount += record.Amount
				totalCost += record.Amount * record.Price
			} else if record.Type == "sell" {
				holdAmount -= record.Amount
				totalSold += record.Amount
				totalSoldValue += record.Amount * record.Price
			}
		}

		// 检查是否有足够的代币可卖
		if holdAmount <= 0 {
			return fmt.Errorf("没有足够的代币 %s 可卖出", tokenAddress)
		}

		// 计算平均买入价格和当前持仓成本
		avgBuyPrice := totalCost / (totalSold + holdAmount) // 总买入均价
		currentAvgCost := 0.0
		if holdAmount > 0 {
			currentAvgCost = (totalCost - (totalSold * avgBuyPrice)) / holdAmount // 当前持仓均价
		}

		// 计算卖出数量和预期收益
		sellAmount := holdAmount * ratio
		if sellAmount <= 0 {
			return fmt.Errorf("卖出数量必须大于0")
		}

		// 计算本次卖出的相关指标
		sellValue := sellAmount * tokenPrice.USDPrice                             // 卖出金额
		sellCost := sellAmount * currentAvgCost                                   // 卖出成本
		sellProfit := sellValue - sellCost                                        // 卖出收益
		sellProfitRate := (tokenPrice.USDPrice - currentAvgCost) / currentAvgCost // 收益率

		// 使用醒目标记突出重要卖出信息
		log.Printf("【卖出信号】Token %s 符合卖出信号", tokenAddress)
		log.Printf("【当前持仓】持有数量: %.6f | 持仓均价: %.6f USD", holdAmount, currentAvgCost)
		log.Printf("【卖出详情】卖出数量: %.6f | 卖出价格: %.6f USD | 卖出比例: %.2f%%",
			sellAmount, tokenPrice.USDPrice, ratio*100)
		log.Printf("【收益分析】卖出金额: %.2f USD | 卖出成本: %.2f USD | 收益: %.2f USD (%.2f%%)",
			sellValue, sellCost, sellProfit, sellProfitRate*100)

		// 创建交易记录
		transactionRecord := &sqlite3.TransactionRecord{
			TokenAddress: tokenAddress,
			Type:         "sell",
			Amount:       sellAmount,
			Price:        tokenPrice.USDPrice,
			Timestamp:    time.Now().Unix(),
			TxHash:       tokenAddress + "_" + strconv.FormatInt(time.Now().Unix(), 10),
			Status:       "confirmed",
		}

		// 插入交易记录
		if err := transactionrecord.InsertTransactionRecord(ctx, transactionRecord); err != nil {
			return fmt.Errorf("插入交易记录失败: %v", err)
		}

		// 计算卖出后的持仓情况
		newHoldAmount := holdAmount - sellAmount
		newTotalCost := totalCost - (sellAmount * currentAvgCost)
		newAvgCost := 0.0
		if newHoldAmount > 0 {
			newAvgCost = newTotalCost / newHoldAmount
		}

		log.Printf("【卖出成功】Token %s", tokenAddress)
		log.Printf("【最新持仓】持有数量: %.6f | 持仓均价: %.6f USD | 当前价格: %.6f USD",
			newHoldAmount, newAvgCost, tokenPrice.USDPrice)
		log.Printf("================================================================")

		// 更新监控记录
		monitorsInterface, err := tokenmonitor.QueryTokenMonitors(tokenmonitor.TokenMonitorQueryOptions{
			TokenAddress: tokenAddress,
			Limit:        1,
		})

		if err == nil {
			if monitors, ok := monitorsInterface.(sqlite3.TokenMonitorSlice); ok && len(monitors) > 0 {
				monitor := monitors[0]
				monitor.Amount = newHoldAmount
				monitor.Price = tokenPrice.USDPrice
				monitor.BuyPrice = newAvgCost
				monitor.Increase = (tokenPrice.USDPrice - newAvgCost) / newAvgCost
				monitor.SellPercentage = (monitor.BuyAmount - newHoldAmount) / monitor.BuyAmount
				monitor.TotalSellPrice += sellValue
				monitor.LastSellTime = float64(time.Now().Unix())

				// 如果全部卖出，停止监控
				if newHoldAmount <= 0 {
					monitor.IsMonitoring = 0
				}

				_ = tokenmonitor.UpdateTokenMonitor(ctx, monitor)
			}
		}

	case SignalHold:
		log.Printf("【持有信号】Token %s 当前无交易信号，保持持有", tokenAddress)

		// 可选：更新价格信息
		monitorsInterface, err := tokenmonitor.QueryTokenMonitors(tokenmonitor.TokenMonitorQueryOptions{
			TokenAddress: tokenAddress,
			Limit:        1,
		})

		if err == nil {
			monitors, ok := monitorsInterface.(sqlite3.TokenMonitorSlice)
			if ok && len(monitors) > 0 {
				// 更新价格信息
				monitor := monitors[0]
				monitor.Price = tokenPrice.USDPrice
				monitor.Increase = (monitor.Price - monitor.BuyPrice) / monitor.BuyPrice

				// 更新数据库记录（可选，不影响主要功能）
				_ = tokenmonitor.UpdateTokenMonitor(ctx, monitor)

				log.Printf("【价格更新】Token %s 当前价格: %.6f USD | 涨幅: %.2f%%",
					tokenAddress, monitor.Price, monitor.Increase*100)
			}
		}

		// 获取该代币的交易记录，计算持有数量和平均买入价格（用于日志显示）
		recordsInterface, err := transactionrecord.QueryTransactionRecords(transactionrecord.TransactionRecordQueryOptions{
			TokenAddress: tokenAddress,
			Status:       "confirmed",
		})

		if err == nil {
			records, ok := recordsInterface.(sqlite3.TransactionRecordSlice)
			if ok {
				// 计算当前持有数量
				var holdAmount float64

				// 先计算总的买入和卖出数量
				for _, record := range records {
					if record.Type == "buy" {
						holdAmount += record.Amount
					} else if record.Type == "sell" {
						holdAmount -= record.Amount
					}
				}

				if holdAmount > 0 {
					// 使用FIFO方法计算当前持有代币的平均买入价格
					var currentHoldValue float64

					// 创建交易记录的副本并按时间戳排序
					sortedRecords := make([]sqlite3.TransactionRecord, len(records))
					// 将 TransactionRecordSlice 转换为 []sqlite3.TransactionRecord
					for i, record := range records {
						sortedRecords[i] = *record
					}

					// 先找出所有卖出记录，计算总卖出量
					// 按时间戳降序排序，最新的交易在前面
					sort.Slice(sortedRecords, func(i, j int) bool {
						return sortedRecords[i].Timestamp > sortedRecords[j].Timestamp
					})

					var totalSellAmount float64
					for _, record := range sortedRecords {
						if record.Type == "sell" {
							totalSellAmount += record.Amount
						}
					}

					// 再按时间顺序处理买入记录，计算当前持有的代币的成本
					// 按时间戳升序排序，最早的交易在前面
					sort.Slice(sortedRecords, func(i, j int) bool {
						return sortedRecords[i].Timestamp < sortedRecords[j].Timestamp
					})

					var processedSellAmount float64
					for _, record := range sortedRecords {
						if record.Type == "buy" {
							buyAmount := record.Amount
							// 如果有卖出，则减去已卖出的部分
							if processedSellAmount < totalSellAmount {
								// 如果当前买入记录的数量小于等于剩余需要处理的卖出数量
								if buyAmount <= (totalSellAmount - processedSellAmount) {
									// 这笔买入已全部卖出，跳过
									processedSellAmount += buyAmount
									continue
								} else {
									// 这笔买入部分卖出，只计算剩余部分
									remainingBuyAmount := buyAmount - (totalSellAmount - processedSellAmount)
									currentHoldValue += remainingBuyAmount * record.Price
									processedSellAmount = totalSellAmount // 所有卖出都已处理完
								}
							} else {
								// 没有卖出或卖出已处理完，全部计入当前持有
								currentHoldValue += buyAmount * record.Price
							}
						}
					}

					// 计算平均买入价格和涨幅
					avgBuyPrice := currentHoldValue / holdAmount
					increase := (tokenPrice.USDPrice - avgBuyPrice) / avgBuyPrice

					log.Printf("【持仓信息】Token %s 持有数量: %.6f | 平均买入价: %.6f | 当前价格: %.6f | 涨幅: %.2f%%",
						tokenAddress, holdAmount, avgBuyPrice, tokenPrice.USDPrice, increase*100)
				}
			}
		}
	}

	return nil
}
