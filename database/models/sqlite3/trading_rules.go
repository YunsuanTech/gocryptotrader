package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries"
	"github.com/thrasher-corp/sqlboiler/queries/qm"
	"github.com/thrasher-corp/sqlboiler/strmangle"
)

// TradingRule is an object representing the database table.
type TradingRule struct {
	RuleID         int           `boil:"rule_id" json:"rule_id" toml:"rule_id" yaml:"rule_id"`
	TokenID        int           `boil:"token_id" json:"token_id" toml:"token_id" yaml:"token_id"`
	UserAddress    string        `boil:"user_address" json:"user_address" toml:"user_address" yaml:"user_address"`
	Direction      string        `boil:"direction" json:"direction" toml:"direction" yaml:"direction"`
	TriggerPrice   float64       `boil:"trigger_price" json:"trigger_price" toml:"trigger_price" yaml:"trigger_price"`
	Quantity       float64       `boil:"quantity" json:"quantity" toml:"quantity" yaml:"quantity"`
	Slippage       float64       `boil:"slippage" json:"slippage" toml:"slippage" yaml:"slippage"`
	ExpirationTime sql.NullInt64 `boil:"expiration_time" json:"expiration_time" toml:"expiration_time" yaml:"expiration_time"`
	IsEnabled      int           `boil:"is_enabled" json:"is_enabled" toml:"is_enabled" yaml:"is_enabled"`
	OrderType      string        `boil:"order_type" json:"order_type" toml:"order_type" yaml:"order_type"`
	CreatedAt      int64         `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	LastTriggered  sql.NullInt64 `boil:"last_triggered" json:"last_triggered" toml:"last_triggered" yaml:"last_triggered"`

	R *tradingRuleR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L tradingRuleL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *TradingRule) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("sqlite3: no trading_rules provided for insertion")
	}

	// 定义列信息
	tradingRuleAllColumns := []string{"rule_id", "token_id", "user_address", "direction", "trigger_price", "quantity", "slippage", "expiration_time", "is_enabled", "order_type", "created_at", "last_triggered"}
	tradingRuleColumnsWithDefault := []string{"rule_id"}
	tradingRuleColumnsWithoutDefault := []string{"token_id", "user_address", "direction", "trigger_price", "quantity", "slippage", "expiration_time", "is_enabled", "order_type", "created_at", "last_triggered"}

	// 获取要插入的列
	wl, _ := columns.InsertColumnSet(
		tradingRuleAllColumns,
		tradingRuleColumnsWithDefault,
		tradingRuleColumnsWithoutDefault,
		queries.NonZeroDefaultSet(tradingRuleColumnsWithDefault, o),
	)

	// 构建SQL查询
	var query string
	if len(wl) != 0 {
		query = fmt.Sprintf("INSERT INTO \"trading_rules\" (\"%s\") VALUES (%s)",
			strings.Join(wl, "\",\""),
			strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
	} else {
		query = "INSERT INTO \"trading_rules\" () VALUES ()"
	}

	// 直接从结构体获取值
	tradingRuleType := reflect.TypeOf(TradingRule{})
	value := reflect.Indirect(reflect.ValueOf(o))

	// 创建值映射
	vals := make([]interface{}, len(wl))
	for i, colName := range wl {
		field, _ := tradingRuleType.FieldByNameFunc(func(fieldName string) bool {
			field, _ := tradingRuleType.FieldByName(fieldName)
			tag := field.Tag.Get("boil")
			return tag == colName
		})
		vals[i] = value.FieldByIndex(field.Index).Interface()
	}

	// 执行插入
	_, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to insert into trading_rules")
	}

	return nil
}

// tradingRuleR is where relationships are stored.
type tradingRuleR struct {
}

// NewStruct creates a new relationship struct
func (*tradingRuleR) NewStruct() *tradingRuleR {
	return &tradingRuleR{}
}

// tradingRuleL is where Load methods for each relationship are stored.
type tradingRuleL struct{}

// TradingRuleQuery is used to build up a query for TradingRule records
type tradingRuleQuery struct {
	*queries.Query
}

// TradingRuleSlice is an alias for a slice of pointers to TradingRule
type TradingRuleSlice []*TradingRule

// TradingRules retrieves all the records using an executor
func TradingRules(mods ...qm.QueryMod) tradingRuleQuery {
	mods = append(mods, qm.From("\"trading_rules\""))
	return tradingRuleQuery{NewQuery(mods...)}
}

// FindTradingRule retrieves a single record by RuleID with an executor.
// If selectCols is empty Find will return all columns.
func FindTradingRule(ctx context.Context, exec boil.ContextExecutor, ruleID int, selectCols ...string) (*TradingRule, error) {
	tradingRuleObj := &TradingRule{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"trading_rules\" where \"rule_id\"=$1", sel,
	)

	q := queries.Raw(query, ruleID)

	err := q.Bind(ctx, exec, tradingRuleObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到ID为 %d 的交易规则", ruleID)
		}
		return nil, fmt.Errorf("查询交易规则失败: %v", err)
	}

	return tradingRuleObj, nil
}

// FindTradingRulesByTokenID retrieves records by TokenID with an executor.
func FindTradingRulesByTokenID(ctx context.Context, exec boil.ContextExecutor, tokenID int, selectCols ...string) (TradingRuleSlice, error) {
	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"trading_rules\" where \"token_id\"=$1", sel,
	)

	q := queries.Raw(query, tokenID)

	var result TradingRuleSlice
	err := q.Bind(ctx, exec, &result)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到代币ID为 %d 的交易规则", tokenID)
		}
		return nil, fmt.Errorf("查询交易规则失败: %v", err)
	}

	return result, nil
}

// FindTradingRulesByUserAddress retrieves records by UserAddress with an executor.
func FindTradingRulesByUserAddress(ctx context.Context, exec boil.ContextExecutor, userAddress string, selectCols ...string) (TradingRuleSlice, error) {
	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"trading_rules\" where \"user_address\"=$1", sel,
	)

	q := queries.Raw(query, userAddress)

	var result TradingRuleSlice
	err := q.Bind(ctx, exec, &result)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到用户地址为 %s 的交易规则", userAddress)
		}
		return nil, fmt.Errorf("查询交易规则失败: %v", err)
	}

	return result, nil
}

// FindTradingRulesByUserAndToken retrieves records by UserAddress and TokenID with an executor.
func FindTradingRulesByUserAndToken(ctx context.Context, exec boil.ContextExecutor, userAddress string, tokenID int, selectCols ...string) (TradingRuleSlice, error) {
	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"trading_rules\" where \"user_address\"=$1 AND \"token_id\"=$2", sel,
	)

	q := queries.Raw(query, userAddress, tokenID)

	var result TradingRuleSlice
	err := q.Bind(ctx, exec, &result)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到用户地址为 %s 且代币ID为 %d 的交易规则", userAddress, tokenID)
		}
		return nil, fmt.Errorf("查询交易规则失败: %v", err)
	}

	return result, nil
}

// FindActiveTradingRules retrieves all active trading rules with an executor.
func FindActiveTradingRules(ctx context.Context, exec boil.ContextExecutor, selectCols ...string) (TradingRuleSlice, error) {
	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"trading_rules\" where \"is_enabled\"=1", sel,
	)

	q := queries.Raw(query)

	var result TradingRuleSlice
	err := q.Bind(ctx, exec, &result)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到活跃的交易规则")
		}
		return nil, fmt.Errorf("查询活跃交易规则失败: %v", err)
	}

	return result, nil
}

// All returns all TradingRule records from the query.
func (q tradingRuleQuery) All(ctx context.Context, exec boil.ContextExecutor) (TradingRuleSlice, error) {
	var o TradingRuleSlice

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, fmt.Errorf("查询所有交易规则失败: %v", err)
	}

	return o, nil
}

// One returns a single TradingRule record from the query.
func (q tradingRuleQuery) One(ctx context.Context, exec boil.ContextExecutor) (*TradingRule, error) {
	o := &TradingRule{}

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到符合条件的交易规则")
		}
		return nil, fmt.Errorf("查询单条交易规则失败: %v", err)
	}

	return o, nil
}

// Update 更新交易规则记录
func (o *TradingRule) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if o == nil {
		return 0, errors.New("sqlite3: no trading_rules provided for update")
	}

	var query string
	vals := make([]interface{}, 0, 12)

	if updateColumns := columns.UpdateColumnSet([]string{
		"token_id", "user_address", "direction", "trigger_price", "quantity", "slippage",
		"expiration_time", "is_enabled", "order_type", "created_at", "last_triggered",
	}, nil); len(updateColumns) > 0 {
		vals = append(vals, o.TokenID, o.UserAddress, o.Direction, o.TriggerPrice, o.Quantity, o.Slippage,
			o.ExpirationTime, o.IsEnabled, o.OrderType, o.CreatedAt, o.LastTriggered)
		query = fmt.Sprintf(
			"UPDATE \"trading_rules\" SET %s WHERE \"rule_id\"=$%d",
			strmangle.SetParamNames("\"", "\"", 1, updateColumns),
			len(vals)+1,
		)
	} else {
		return 0, errors.New("sqlite3: unable to update trading_rules, could not build update column list")
	}

	vals = append(vals, o.RuleID)

	result, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlite3: unable to update trading_rules row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlite3: failed to get rows affected by update for trading_rules")
	}

	return rowsAff, nil
}

// Delete 删除交易规则记录
func (o *TradingRule) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("sqlite3: no TradingRule provided for delete")
	}

	query := "DELETE FROM \"trading_rules\" WHERE \"rule_id\"=$1"

	result, err := exec.ExecContext(ctx, query, o.RuleID)
	if err != nil {
		return 0, errors.Wrap(err, "sqlite3: unable to delete from trading_rules")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlite3: failed to get rows affected by delete for trading_rules")
	}

	return rowsAff, nil
}
