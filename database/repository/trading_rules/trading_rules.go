package trading_rules

import (
	"context"
	"database/sql"
	"time"

	"gocryptotrader/database"
	modelSQLite "gocryptotrader/database/models/sqlite3"

	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries/qm"
)

// GetTradingRules 获取所有交易规则
func GetTradingRules(limit int) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	var mods []qm.QueryMod
	mods = append(mods, qm.OrderBy("created_at DESC"))
	if limit > 0 {
		mods = append(mods, qm.Limit(limit))
	}

	ctx := context.TODO()
	return modelSQLite.TradingRules(mods...).All(ctx, database.DB.SQL)
}

// GetTradingRuleByID 根据ID获取交易规则
func GetTradingRuleByID(ruleID int) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindTradingRule(ctx, database.DB.SQL, ruleID)
}

// GetTradingRulesByTokenID 根据代币ID获取交易规则
func GetTradingRulesByTokenID(tokenID int) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindTradingRulesByTokenID(ctx, database.DB.SQL, tokenID)
}

// GetTradingRulesByUserAddress 根据用户地址获取交易规则
func GetTradingRulesByUserAddress(userAddress string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindTradingRulesByUserAddress(ctx, database.DB.SQL, userAddress)
}

// GetTradingRulesByUserAndToken 根据用户地址和代币ID获取交易规则
func GetTradingRulesByUserAndToken(userAddress string, tokenID int) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindTradingRulesByUserAndToken(ctx, database.DB.SQL, userAddress, tokenID)
}

// GetActiveTradingRules 获取所有活跃的交易规则
func GetActiveTradingRules() (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindActiveTradingRules(ctx, database.DB.SQL)
}

// AddTradingRule 添加新的交易规则
func AddTradingRule(tokenID int, userAddress, direction string, triggerPrice, quantity, slippage float64, expirationTime, createdAt, lastTriggered int64, isEnabled int, orderType string) error {
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 创建一个新的TradingRule对象
	rule := &modelSQLite.TradingRule{
		TokenID:        tokenID,
		UserAddress:    userAddress,
		Direction:      direction,
		TriggerPrice:   triggerPrice,
		Quantity:       quantity,
		Slippage:       slippage,
		ExpirationTime: sql.NullInt64{Int64: expirationTime, Valid: expirationTime != 0},
		IsEnabled:      isEnabled,
		OrderType:      orderType,
		CreatedAt:      createdAt,
		LastTriggered:  sql.NullInt64{Int64: lastTriggered, Valid: lastTriggered != 0},
	}

	// 插入记录到数据库
	ctx := context.TODO()
	return rule.Insert(ctx, database.DB.SQL, boil.Infer())
}

// UpdateTradingRule 更新交易规则
func UpdateTradingRule(ruleID, tokenID int, userAddress, direction string, triggerPrice, quantity, slippage float64, expirationTime int64, isEnabled int, orderType string, lastTriggered int64) error {
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 获取现有规则
	ctx := context.TODO()
	existingRule, err := modelSQLite.FindTradingRule(ctx, database.DB.SQL, ruleID)
	if err != nil {
		return err
	}

	// 更新规则字段
	existingRule.TokenID = tokenID
	existingRule.UserAddress = userAddress
	existingRule.Direction = direction
	existingRule.TriggerPrice = triggerPrice
	existingRule.Quantity = quantity
	existingRule.Slippage = slippage
	existingRule.ExpirationTime = sql.NullInt64{Int64: expirationTime, Valid: expirationTime != 0}
	existingRule.IsEnabled = isEnabled
	existingRule.OrderType = orderType
	existingRule.LastTriggered = sql.NullInt64{Int64: lastTriggered, Valid: lastTriggered != 0}

	// 更新数据库记录
	_, err = existingRule.Update(ctx, database.DB.SQL, boil.Infer())
	return err
}

// DeleteTradingRule 删除交易规则
func DeleteTradingRule(ruleID int) error {
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 获取现有规则
	ctx := context.TODO()
	existingRule, err := modelSQLite.FindTradingRule(ctx, database.DB.SQL, ruleID)
	if err != nil {
		return err
	}

	// 从数据库删除记录
	_, err = existingRule.Delete(ctx, database.DB.SQL)
	return err
}

// UpdateLastTriggered 更新规则的最后触发时间
func UpdateLastTriggered(ruleID int) error {
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 获取现有规则
	ctx := context.TODO()
	existingRule, err := modelSQLite.FindTradingRule(ctx, database.DB.SQL, ruleID)
	if err != nil {
		return err
	}

	// 更新最后触发时间为当前时间
	existingRule.LastTriggered = sql.NullInt64{Int64: time.Now().Unix(), Valid: true}

	// 更新数据库记录
	_, err = existingRule.Update(ctx, database.DB.SQL, boil.Infer())
	return err
}
