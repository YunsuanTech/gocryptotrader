package token

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"gocryptotrader/config"
	tradingsql "gocryptotrader/database/repository/trading_rules"
)

// TradingManager 管理交易规则相关操作
type TradingManager struct {
	config *config.Config
}

// NewTradingManager 创建一个新的交易规则管理器
func NewTradingManager(cfg *config.Config) *TradingManager {
	return &TradingManager{config: cfg}
}

// TradingRule 代表交易规则信息
type TradingRule struct {
	RuleID         int
	TokenID        int
	UserAddress    string
	Direction      string
	TriggerPrice   float64
	Quantity       float64
	Slippage       float64
	ExpirationTime int64
	IsEnabled      int
	OrderType      string
	CreatedAt      int64
	LastTriggered  int64
}

// processRulesResult 处理交易规则查询结果并转换为适当的类型
func processRulesResult(rules interface{}, operation string) ([]*TradingRule, error) {
	switch v := rules.(type) {
	case []*TradingRule:
		// 如果已经是正确的类型，直接返回
		return v, nil
	case interface{}:
		// 尝试从反射获取切片中的每个元素并转换
		return convertToTradingRuleSlice(v)
	default:
		return nil, fmt.Errorf("无法转换交易规则数据类型：未知类型 %T", rules)
	}
}

// processRuleResult 处理单个交易规则查询结果并转换为适当的类型
func processRuleResult(rule interface{}, operation string) (*TradingRule, error) {
	// 尝试将interface{}转换为*TradingRule
	ruleObj, ok := rule.(*TradingRule)
	if ok {
		return ruleObj, nil
	}

	// 使用反射进行转换
	result := convertStructToTradingRule(rule)
	if result == nil {
		return nil, fmt.Errorf("无法转换交易规则数据类型: %T", rule)
	}

	return result, nil
}

// GetAllTradingRules 获取所有交易规则
func (m *TradingManager) GetAllTradingRules(limit int) ([]*TradingRule, error) {
	return m.QueryTradingRules(
		func() (interface{}, error) {
			return tradingsql.GetTradingRules(limit)
		},
		"获取交易规则列表",
	)
}

// QueryTradingRules 通用查询函数，用于获取交易规则列表
func (m *TradingManager) QueryTradingRules(queryFunc func() (interface{}, error), operation string) ([]*TradingRule, error) {
	rules, err := queryFunc()
	if err != nil {
		return nil, fmt.Errorf("%s失败: %w", operation, err)
	}

	return processRulesResult(rules, operation)
}

// QueryTradingRule 通用查询函数，用于获取单个交易规则
func (m *TradingManager) QueryTradingRule(queryFunc func() (interface{}, error), operation string) (*TradingRule, error) {
	rule, err := queryFunc()
	if err != nil {
		return nil, fmt.Errorf("%s失败: %w", operation, err)
	}

	return processRuleResult(rule, operation)
}

// convertToTradingRuleSlice 将接口类型转换为[]*TradingRule
func convertToTradingRuleSlice(data interface{}) ([]*TradingRule, error) {
	var result []*TradingRule

	sliceValue := reflect.ValueOf(data)
	if sliceValue.Kind() != reflect.Slice {
		return nil, fmt.Errorf("无法转换交易规则数据类型：不是切片类型")
	}

	// 遍历切片中的每个元素
	for i := 0; i < sliceValue.Len(); i++ {
		item := sliceValue.Index(i).Interface()

		// 尝试将每个元素转换为map并创建TradingRule对象
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			// 如果不是map，尝试直接转换
			ruleObj, ok := item.(*TradingRule)
			if !ok {
				// 如果无法直接转换，尝试使用反射获取字段值
				ruleObj = convertStructToTradingRule(item)
				if ruleObj == nil {
					continue
				}
			}
			result = append(result, ruleObj)
			continue
		}

		// 从map创建TradingRule对象
		ruleObj := mapToTradingRule(itemMap)
		result = append(result, ruleObj)
	}

	return result, nil
}

// convertStructToTradingRule 使用反射将结构体转换为TradingRule
func convertStructToTradingRule(item interface{}) *TradingRule {
	itemValue := reflect.ValueOf(item)
	if itemValue.Kind() == reflect.Ptr {
		itemValue = itemValue.Elem()
	}

	if itemValue.Kind() != reflect.Struct {
		return nil
	}

	// 创建新的TradingRule对象
	ruleObj := &TradingRule{}

	// 尝试获取并设置字段值
	if f := itemValue.FieldByName("RuleID"); f.IsValid() {
		if f.Kind() == reflect.Int || f.Kind() == reflect.Int8 || f.Kind() == reflect.Int16 || f.Kind() == reflect.Int32 {
			ruleObj.RuleID = int(f.Int())
		} else if f.Kind() == reflect.Int64 {
			ruleObj.RuleID = int(f.Int())
		}
	}
	if f := itemValue.FieldByName("TokenID"); f.IsValid() {
		if f.Kind() == reflect.Int || f.Kind() == reflect.Int8 || f.Kind() == reflect.Int16 || f.Kind() == reflect.Int32 {
			ruleObj.TokenID = int(f.Int())
		} else if f.Kind() == reflect.Int64 {
			ruleObj.TokenID = int(f.Int())
		}
	}
	if f := itemValue.FieldByName("UserAddress"); f.IsValid() {
		if nullStr, ok := f.Interface().(sql.NullString); ok {
			if nullStr.Valid {
				ruleObj.UserAddress = nullStr.String
			}
		} else {
			ruleObj.UserAddress = f.String()
		}
	}
	if f := itemValue.FieldByName("Direction"); f.IsValid() {
		if nullStr, ok := f.Interface().(sql.NullString); ok {
			if nullStr.Valid {
				ruleObj.Direction = nullStr.String
			}
		} else {
			ruleObj.Direction = f.String()
		}
	}
	if f := itemValue.FieldByName("TriggerPrice"); f.IsValid() {
		if f.Kind() == reflect.Float32 || f.Kind() == reflect.Float64 {
			ruleObj.TriggerPrice = f.Float()
		}
	}
	if f := itemValue.FieldByName("Quantity"); f.IsValid() {
		if f.Kind() == reflect.Float32 || f.Kind() == reflect.Float64 {
			ruleObj.Quantity = f.Float()
		}
	}
	if f := itemValue.FieldByName("Slippage"); f.IsValid() {
		if f.Kind() == reflect.Float32 || f.Kind() == reflect.Float64 {
			ruleObj.Slippage = f.Float()
		}
	}
	if f := itemValue.FieldByName("ExpirationTime"); f.IsValid() {
		if f.Kind() == reflect.Int64 {
			ruleObj.ExpirationTime = f.Int()
		} else if f.Kind() == reflect.Int || f.Kind() == reflect.Int8 || f.Kind() == reflect.Int16 || f.Kind() == reflect.Int32 {
			ruleObj.ExpirationTime = int64(f.Int())
		}
	}
	if f := itemValue.FieldByName("IsEnabled"); f.IsValid() {
		if f.Kind() == reflect.Int || f.Kind() == reflect.Int8 || f.Kind() == reflect.Int16 || f.Kind() == reflect.Int32 {
			ruleObj.IsEnabled = int(f.Int())
		} else if f.Kind() == reflect.Int64 {
			ruleObj.IsEnabled = int(f.Int())
		}
	}
	if f := itemValue.FieldByName("OrderType"); f.IsValid() {
		if nullStr, ok := f.Interface().(sql.NullString); ok {
			if nullStr.Valid {
				ruleObj.OrderType = nullStr.String
			}
		} else {
			ruleObj.OrderType = f.String()
		}
	}
	if f := itemValue.FieldByName("CreatedAt"); f.IsValid() {
		if f.Kind() == reflect.Int64 {
			ruleObj.CreatedAt = f.Int()
		} else if f.Kind() == reflect.Int || f.Kind() == reflect.Int8 || f.Kind() == reflect.Int16 || f.Kind() == reflect.Int32 {
			ruleObj.CreatedAt = int64(f.Int())
		}
	}
	if f := itemValue.FieldByName("LastTriggered"); f.IsValid() {
		if f.Kind() == reflect.Int64 {
			ruleObj.LastTriggered = f.Int()
		} else if f.Kind() == reflect.Int || f.Kind() == reflect.Int8 || f.Kind() == reflect.Int16 || f.Kind() == reflect.Int32 {
			ruleObj.LastTriggered = int64(f.Int())
		}
	}

	return ruleObj
}

// mapToTradingRule 将map转换为TradingRule
func mapToTradingRule(itemMap map[string]interface{}) *TradingRule {
	ruleObj := &TradingRule{}

	if v, ok := itemMap["rule_id"].(int); ok {
		ruleObj.RuleID = v
	}
	if v, ok := itemMap["token_id"].(int); ok {
		ruleObj.TokenID = v
	}
	if v, ok := itemMap["user_address"].(string); ok {
		ruleObj.UserAddress = v
	}
	if v, ok := itemMap["direction"].(string); ok {
		ruleObj.Direction = v
	}
	if v, ok := itemMap["trigger_price"].(float64); ok {
		ruleObj.TriggerPrice = v
	}
	if v, ok := itemMap["quantity"].(float64); ok {
		ruleObj.Quantity = v
	}
	if v, ok := itemMap["slippage"].(float64); ok {
		ruleObj.Slippage = v
	}
	if v, ok := itemMap["expiration_time"].(int64); ok {
		ruleObj.ExpirationTime = v
	}
	if v, ok := itemMap["is_enabled"].(int); ok {
		ruleObj.IsEnabled = v
	}
	if v, ok := itemMap["order_type"].(string); ok {
		ruleObj.OrderType = v
	}
	if v, ok := itemMap["created_at"].(int64); ok {
		ruleObj.CreatedAt = v
	}
	if v, ok := itemMap["last_triggered"].(int64); ok {
		ruleObj.LastTriggered = v
	}

	return ruleObj
}

// GetTradingRuleByID 根据ID获取交易规则
func (m *TradingManager) GetTradingRuleByID(ruleID int) (*TradingRule, error) {
	return m.QueryTradingRule(
		func() (interface{}, error) {
			return tradingsql.GetTradingRuleByID(ruleID)
		},
		"获取交易规则信息",
	)
}

// GetTradingRulesByTokenID 根据代币ID获取交易规则
func (m *TradingManager) GetTradingRulesByTokenID(tokenID int) ([]*TradingRule, error) {
	return m.QueryTradingRules(
		func() (interface{}, error) {
			return tradingsql.GetTradingRulesByTokenID(tokenID)
		},
		"获取交易规则信息",
	)
}

// GetTradingRulesByUserAddress 根据用户地址获取交易规则
func (m *TradingManager) GetTradingRulesByUserAddress(userAddress string) ([]*TradingRule, error) {
	return m.QueryTradingRules(
		func() (interface{}, error) {
			return tradingsql.GetTradingRulesByUserAddress(userAddress)
		},
		"获取交易规则信息",
	)
}

// GetTradingRulesByUserAndToken 根据用户地址和代币ID获取交易规则
func (m *TradingManager) GetTradingRulesByUserAndToken(userAddress string, tokenID int) ([]*TradingRule, error) {
	return m.QueryTradingRules(
		func() (interface{}, error) {
			return tradingsql.GetTradingRulesByUserAndToken(userAddress, tokenID)
		},
		"获取交易规则信息",
	)
}

// GetActiveTradingRules 获取所有活跃的交易规则
func (m *TradingManager) GetActiveTradingRules() ([]*TradingRule, error) {
	return m.QueryTradingRules(
		func() (interface{}, error) {
			return tradingsql.GetActiveTradingRules()
		},
		"获取活跃交易规则信息",
	)
}

// AddTradingRule 添加新的交易规则
func (m *TradingManager) AddTradingRule(tokenID int, userAddress, direction string, triggerPrice, quantity, slippage float64, expirationTime int64, isEnabled int, orderType string) error {
	// 设置创建时间为当前时间
	createdAt := time.Now().Unix()
	// 初始化最后触发时间为0
	lastTriggered := int64(0)

	// 调用repository层的添加方法
	if err := tradingsql.AddTradingRule(tokenID, userAddress, direction, triggerPrice, quantity, slippage, expirationTime, createdAt, lastTriggered, isEnabled, orderType); err != nil {
		return fmt.Errorf("添加交易规则失败: %w", err)
	}
	return nil
}

// UpdateTradingRule 更新交易规则
func (m *TradingManager) UpdateTradingRule(ruleID, tokenID int, userAddress, direction string, triggerPrice, quantity, slippage float64, expirationTime int64, isEnabled int, orderType string) error {
	// 获取现有规则信息以验证其存在性
	existingRule, err := m.GetTradingRuleByID(ruleID)
	if err != nil {
		return fmt.Errorf("获取现有交易规则失败: %w", err)
	}

	// 调用repository层的更新方法
	err = tradingsql.UpdateTradingRule(ruleID, tokenID, userAddress, direction, triggerPrice, quantity, slippage, expirationTime, isEnabled, orderType, existingRule.LastTriggered)
	if err != nil {
		return fmt.Errorf("更新交易规则失败: %w", err)
	}

	return nil
}

// DeleteTradingRule 删除交易规则
func (m *TradingManager) DeleteTradingRule(ruleID int) error {
	// 获取现有规则信息以验证其存在性
	_, err := m.GetTradingRuleByID(ruleID)
	if err != nil {
		return fmt.Errorf("获取交易规则失败: %w", err)
	}

	// 调用repository层的删除方法
	err = tradingsql.DeleteTradingRule(ruleID)
	if err != nil {
		return fmt.Errorf("删除交易规则失败: %w", err)
	}

	return nil
}
