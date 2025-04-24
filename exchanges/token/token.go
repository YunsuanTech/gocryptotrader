package token

import (
	"context"
	"fmt"
	"log"
	"reflect"
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

// TokenInfo 表示token信息的结构体
type TokenInfo struct {
	Symbol    string  `field:"symbol"`
	Mcap      float64 `field:"mcap"`
	Price     float64 `field:"price"`
	Launchpad string  `field:"launchpad"`
}

// BuySOLToken 实现买入SOL token的功能
func BuySOLToken(client *Client) error {
	ctx := context.Background()
	// 获取所有token池信息
	pools, err := client.GetAllPools()
	if err != nil {
		return fmt.Errorf("获取token池信息失败: %v", err)
	}

	// 获取买入规则
	rules, err := getBuyRules("buy")
	if err != nil {
		return fmt.Errorf("获取买入规则失败: %v", err)
	}

	// 遍历所有token，检查是否符合买入规则
	for _, pool := range pools {
		tokenInfo := TokenInfo{
			Symbol:    pool.BaseAsset.Symbol,
			Mcap:      pool.BaseAsset.MCap,
			Price:     pool.BaseAsset.USDPrice,
			Launchpad: pool.BaseAsset.Launchpad,
		}

		// 检查每个规则
		for _, rule := range rules {
			if matchBuyRule(tokenInfo, rule) {
				// 查询该token是否已存在于监控表中，若已存在则跳过
				monitors, err := tokenmonitor.QueryTokenMonitors(tokenmonitor.TokenMonitorQueryOptions{TokenAddress: pool.BaseAsset.ID, Limit: 1, IsMonitoring: -1})
				if err != nil {
					log.Printf("[买入] 查询TokenMonitor失败: %v", err)
					continue
				}

				if reflect.ValueOf(monitors).Kind() == reflect.Slice && reflect.ValueOf(monitors).Len() != 0 {
					continue
				}

				// 使用醒目标记突出重要买入信息
				log.Printf("【买入信号】Token %s 符合买入规则 '%s'", tokenInfo.Symbol, rule.RuleName)
				log.Printf("【买入详情】市值: %.2f | 价格: %.6f | 规则: %s",
					tokenInfo.Mcap, tokenInfo.Price, rule.Description)

				// 创建新的TokenMonitor记录
				tokenMonitor := &sqlite3.TokenMonitor{
					TokenAddress:  pool.BaseAsset.ID,
					TokenName:     pool.BaseAsset.Symbol,
					Price:         pool.BaseAsset.USDPrice,
					TokenDecimals: pool.BaseAsset.Decimals,
					BuyAmount:     rule.BuyPrice / pool.BaseAsset.USDPrice, // 模拟结果
					Amount:        rule.BuyPrice / pool.BaseAsset.USDPrice, // 模拟结果
					BuyPrice:      pool.BaseAsset.USDPrice,
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
					TokenAddress: pool.BaseAsset.ID,
					RuleID:       rule.ID,
					Type:         "buy",
					Amount:       tokenMonitor.Amount,
					Price:        tokenMonitor.BuyPrice,
					Timestamp:    tokenMonitor.BuyTime,
					TxHash:       pool.BaseAsset.ID, // 模拟交易，暂不设置实际的交易哈希
					Status:       "confirmed",
				}

				// 插入交易记录
				if err := transactionrecord.InsertTransactionRecord(ctx, transactionRecord); err != nil {
					log.Printf("[买入] 插入交易记录失败: %v", err)
				}

				log.Printf("【买入成功】添加Token %s到监控列表，买入价格: %.6f，买入数量: %.6f",
					tokenInfo.Symbol, tokenMonitor.BuyPrice, tokenMonitor.Amount)

				log.Printf("================================================================")
			}
		}
	}

	return nil
}

func SellSOLToken(client *Client) error {
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

		// // 将常规价格更新信息降级为Debug级别，减少日志刷屏
		// log.Printf("[价格更新] Token %s address: %s 当前价格: %.6f USD, 涨幅: %.2f%%",
		// 	monitor.TokenName, monitor.TokenAddress, monitor.Price, monitor.Increase*100)
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

// getFieldValueByName 通过字段名获取TokenInfo结构体中对应字段的值
func getFieldValueByName(token TokenInfo, fieldName string) (interface{}, error) {
	// 使用反射获取结构体的值和类型
	val := reflect.ValueOf(token)
	typ := reflect.TypeOf(token)

	// 创建字段名到结构体字段的映射
	fieldMap := make(map[string]int)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		// 优先使用标签中定义的字段名
		tagName := field.Tag.Get("field")
		if tagName != "" {
			fieldMap[tagName] = i
		}
		// 同时支持直接使用结构体字段名（小写形式）
		fieldMap[strings.ToLower(field.Name)] = i
	}

	// 查找字段
	fieldIndex, ok := fieldMap[strings.ToLower(fieldName)]

	if !ok {
		return nil, fmt.Errorf("未找到字段: %s", fieldName)
	}

	// 获取字段值
	fieldValue := val.Field(fieldIndex)

	// 返回字段值
	return fieldValue.Interface(), nil
}

// compareValues 比较两个值是否满足指定的操作符条件
func compareValues(fieldValue interface{}, operator string, condValue string) (bool, error) {
	switch v := fieldValue.(type) {
	case string:
		// 字符串类型只支持相等操作符
		if operator != "=" {
			return false, fmt.Errorf("字符串字段只支持相等操作符: %s", operator)
		}
		return v == condValue, nil

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
		default:
			return false, fmt.Errorf("未知的操作符: %s", operator)
		}

	case int, int32, int64:
		// 整数类型转换为float64后比较
		intValue := reflect.ValueOf(v).Float()
		numValue, err := strconv.ParseFloat(condValue, 64)
		if err != nil {
			return false, fmt.Errorf("条件值解析错误: %s", condValue)
		}

		switch operator {
		case ">":
			return intValue > numValue, nil
		case "<":
			return intValue < numValue, nil
		case "=":
			return intValue == numValue, nil
		default:
			return false, fmt.Errorf("未知的操作符: %s", operator)
		}

	default:
		return false, fmt.Errorf("不支持的字段类型: %T", v)
	}
}

// matchBuyRule 检查token是否符合买入规则
func matchBuyRule(token TokenInfo, rule BuyRule) bool {
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
		fieldValue, err := getFieldValueByName(token, field)
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
