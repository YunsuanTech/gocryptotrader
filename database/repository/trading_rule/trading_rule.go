package tradingrule

import (
	"context"
	"errors"

	"gocryptotrader/database"
	modelSQLite "gocryptotrader/database/models/sqlite3"

	"github.com/thrasher-corp/sqlboiler/queries/qm"
)

// QueryOptions 定义查询交易规则的选项
type QueryOptions struct {
	ID       int    // 根据 ID 查询
	RuleName string // 根据规则名称查询
	RuleType string // 根据规则类型查询
	Limit    int    // 返回记录的最大数量
}

// QueryTradingRules 根据提供的选项查询交易规则
func QueryTradingRules(opts QueryOptions) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}
	var mods []qm.QueryMod

	// 添加 ID 条件
	if opts.ID != 0 {
		mods = append(mods, qm.Where("id = ?", opts.ID))
	}
	// 添加 RuleName 条件
	if opts.RuleName != "" {
		mods = append(mods, qm.Where("rule_name = ?", opts.RuleName))
	}
	// 添加 RuleType 条件
	if opts.RuleType != "" {
		mods = append(mods, qm.Where("rule_type =?", opts.RuleType))
	}

	// 默认按 ID 排序
	mods = append(mods, qm.OrderBy("id"))

	// 设置查询限制
	if opts.Limit > 0 {
		mods = append(mods, qm.Limit(opts.Limit))
	}

	ctx := context.TODO()
	return modelSQLite.TradingRules(mods...).All(ctx, database.DB.SQL)
}

// InsertTradingRule 插入一条新的交易规则记录
func InsertTradingRule(ctx context.Context, rule *modelSQLite.TradingRule) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的规则是否为 nil
	if rule == nil {
		return errors.New("trading rule cannot be nil")
	}

	// 调用 TradingRule 的 Insert 方法执行插入
	err := rule.Insert(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// UpdateTradingRule 修改一条新的交易规则记录
func UpdateTradingRule(ctx context.Context, rule *modelSQLite.TradingRule) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的规则是否为 nil
	if rule == nil {
		return errors.New("trading rule cannot be nil")
	}

	// 调用 TradingRule 的 Insert 方法执行插入
	err := rule.Update(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// UpdateTradingRule 修改一条新的交易规则记录
func DeleteTradingRule(ctx context.Context, rule *modelSQLite.TradingRule) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的规则是否为 nil
	if rule == nil {
		return errors.New("trading rule cannot be ID = 0")
	}

	// 调用 TradingRule 的 Insert 方法执行插入
	err := rule.Delete(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}
