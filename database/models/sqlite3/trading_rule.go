package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries"
	"github.com/thrasher-corp/sqlboiler/queries/qm"
	"github.com/thrasher-corp/sqlboiler/strmangle"
)

// TradingRule 是表示 trading_rules 数据库表的结构体
type TradingRule struct {
	ID          int             `boil:"id" json:"id" toml:"id" yaml:"id"`
	RuleName    string          `boil:"rule_name" json:"rule_name" toml:"rule_name" yaml:"rule_name"`
	RuleType    string          `boil:"rule_type" json:"rule_type" toml:"rule_type" yaml:"rule_type"`
	BuyPrice    sql.NullFloat64 `boil:"buy_price" json:"buy_price" toml:"buy_price" yaml:"buy_price"`
	SellPrice   sql.NullFloat64 `boil:"sell_price" json:"sell_price" toml:"sell_price" yaml:"sell_price"`
	Condition   string          `boil:"condition" json:"condition" toml:"condition" yaml:"condition"`
	Priority    sql.NullFloat64 `boil:"priority" json:"priority" toml:"priority" yaml:"priority"`
	Description sql.NullString  `boil:"description" json:"description" toml:"description" yaml:"description"`
}

// Insert 使用执行器插入单条记录
func (o *TradingRule) Insert(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no trading_rule provided for insertion")
	}

	// 定义要插入的列，排除自增的 id
	tradingRuleColumns := []string{
		"rule_name",
		"rule_type",
		"buy_price",
		"sell_price",
		"condition",
		"priority",
		"description",
	}

	// 构建 SQL 查询
	query := fmt.Sprintf(
		"INSERT INTO \"trading_rules\" (\"%s\") VALUES (%s)",
		strings.Join(tradingRuleColumns, "\",\""),
		strmangle.Placeholders(dialect.UseIndexPlaceholders, len(tradingRuleColumns), 1, 1),
	)

	// 准备插入的值
	vals := []interface{}{
		o.RuleName,
		o.RuleType,
		o.BuyPrice,
		o.SellPrice,
		o.Condition,
		o.Priority,
		o.Description,
	}

	// 执行插入操作
	result, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to insert into trading_rules")
	}

	// 获取自增 ID 并赋值给结构体
	id, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to get last insert id for trading_rules")
	}
	o.ID = int(id)

	return nil
}

// Update 更新数据库中的 TradingRule 记录
func (o *TradingRule) Update(ctx context.Context, exec boil.ContextExecutor) error {
	// 检查 TradingRule 实例是否为 nil
	if o == nil {
		return errors.New("sqlite3: no trading_rule provided for update")
	}

	// 定义要更新的字段（不包括 id）
	columns := []string{
		"rule_name",
		"rule_type",
		"buy_price",
		"sell_price",
		"condition",
		"priority",
		"description",
	}

	// 构建 SQL 更新查询
	query := fmt.Sprintf(
		"UPDATE \"trading_rules\" SET \"%s\" = %s WHERE \"id\" = ?",
		strings.Join(columns, "\" = ?, \""),
		strmangle.Placeholders(dialect.UseIndexPlaceholders, len(columns), 1, 1),
	)

	// 准备更新的值
	vals := []interface{}{
		o.RuleName,
		o.RuleType,
		o.BuyPrice,
		o.SellPrice,
		o.Condition,
		o.Priority,
		o.Description,
		o.ID, // 用于 WHERE 条件
	}

	// 执行更新操作
	_, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to update trading_rules")
	}

	return nil
}

// Delete 从数据库中删除 TradingRule 记录
func (o *TradingRule) Delete(ctx context.Context, exec boil.ContextExecutor) error {
	// 检查 TradingRule 实例是否为 nil
	if o == nil {
		return errors.New("sqlite3: no trading_rule provided for deletion")
	}

	// 检查主键 id 是否已设置
	if o.ID == 0 {
		return errors.New("sqlite3: trading_rule has no primary key value for deletion")
	}

	// 构建 SQL 删除查询
	query := fmt.Sprintf(
		"DELETE FROM \"trading_rules\" WHERE \"id\" = ?",
	)

	// 执行删除操作
	_, err := exec.ExecContext(ctx, query, o.ID)
	if err != nil {
		return fmt.Errorf("sqlite3: unable to delete from trading_rules: %w", err)
	}

	return nil
}

// tradingRuleQuery 用于构建 TradingRule 记录的查询
type tradingRuleQuery struct {
	*queries.Query
}

// TradingRuleSlice 是指向 TradingRule 的指针切片别名
type TradingRuleSlice []*TradingRule

// TradingRules 使用执行器检索所有记录
func TradingRules(mods ...qm.QueryMod) tradingRuleQuery {
	mods = append(mods, qm.From("\"trading_rules\""))
	return tradingRuleQuery{NewQuery(mods...)}
}

// All 从查询中返回所有 TradingRule 记录
func (q tradingRuleQuery) All(ctx context.Context, exec boil.ContextExecutor) (TradingRuleSlice, error) {
	var o TradingRuleSlice

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "sqlite3: failed to assign all query results to TradingRule slice")
	}

	return o, nil
}
