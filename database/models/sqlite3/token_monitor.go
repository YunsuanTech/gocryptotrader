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

// TokenMonitor 是表示 token_monitor 数据库表的结构体
type TokenMonitor struct {
	TokenAddress   string          `boil:"token_address" json:"token_address" toml:"token_address" yaml:"token_address"`
	TokenName      string          `boil:"token_name" json:"token_name" toml:"token_name" yaml:"token_name"`
	Price          sql.NullFloat64 `boil:"price" json:"price" toml:"price" yaml:"price"`
	TokenDecimals  int             `boil:"token_decimals" json:"token_decimals" toml:"token_decimals" yaml:"token_decimals"`
	BuyAmount      sql.NullFloat64 `boil:"buy_amount" json:"buy_amount" toml:"buy_amount" yaml:"buy_amount"`
	BuyPrice       sql.NullFloat64 `boil:"buy_price" json:"buy_price" toml:"buy_price" yaml:"buy_price"`
	BuyTime        sql.NullInt64   `boil:"buy_time" json:"buy_time" toml:"buy_time" yaml:"buy_time"`
	SellPercentage sql.NullFloat64 `boil:"sell_percentage" json:"sell_percentage" toml:"sell_percentage" yaml:"sell_percentage"`
	TotalSellPrice sql.NullFloat64 `boil:"total_sell_price" json:"total_sell_price" toml:"total_sell_price" yaml:"total_sell_price"`
	LastSellTime   sql.NullInt64   `boil:"last_sell_time" json:"last_sell_time" toml:"last_sell_time" yaml:"last_sell_time"`
	IsMonitoring   int             `boil:"is_monitoring" json:"is_monitoring" toml:"is_monitoring" yaml:"is_monitoring"`
}

// Insert 使用执行器插入单条记录
func (o *TokenMonitor) Insert(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no token_monitor provided for insertion")
	}

	// 定义要插入的列
	tokenMonitorColumns := []string{
		"token_address",
		"token_name",
		"price",
		"token_decimals",
		"buy_amount",
		"buy_price",
		"buy_time",
		"sell_percentage",
		"total_sell_price",
		"last_sell_time",
		"is_monitoring",
	}

	// 构建 SQL 查询
	query := fmt.Sprintf(
		"INSERT INTO \"token_monitor\" (\"%s\") VALUES (%s)",
		strings.Join(tokenMonitorColumns, "\",\""),
		strmangle.Placeholders(dialect.UseIndexPlaceholders, len(tokenMonitorColumns), 1, 1),
	)

	// 准备插入的值
	vals := []interface{}{
		o.TokenAddress,
		o.TokenName,
		o.Price,
		o.TokenDecimals,
		o.BuyAmount,
		o.BuyPrice,
		o.BuyTime,
		o.SellPercentage,
		o.TotalSellPrice,
		o.LastSellTime,
		o.IsMonitoring,
	}

	// 执行插入操作
	_, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to insert into token_monitor")
	}

	return nil
}

// Update 更新数据库中的 TokenMonitor 记录
func (o *TokenMonitor) Update(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no token_monitor provided for update")
	}

	// 定义要更新的字段（不包括主键 token_address）
	columns := []string{
		"token_name",
		"price",
		"token_decimals",
		"buy_amount",
		"buy_price",
		"buy_time",
		"sell_percentage",
		"total_sell_price",
		"last_sell_time",
		"is_monitoring",
	}

	// 构建 SQL 更新查询
	query := fmt.Sprintf(
		"UPDATE \"token_monitor\" SET \"%s\" = %s WHERE \"token_address\" = ?",
		strings.Join(columns, "\" = ?, \""),
		strmangle.Placeholders(dialect.UseIndexPlaceholders, len(columns), 1, 1),
	)

	// 准备更新的值
	vals := []interface{}{
		o.TokenName,
		o.Price,
		o.TokenDecimals,
		o.BuyAmount,
		o.BuyPrice,
		o.BuyTime,
		o.SellPercentage,
		o.TotalSellPrice,
		o.LastSellTime,
		o.IsMonitoring,
		o.TokenAddress, // 用于 WHERE 条件
	}

	// 执行更新操作
	_, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to update token_monitor")
	}

	return nil
}

// Delete 从数据库中删除 TokenMonitor 记录
func (o *TokenMonitor) Delete(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no token_monitor provided for deletion")
	}

	// 检查主键 token_address 是否已设置
	if o.TokenAddress == "" {
		return errors.New("sqlite3: token_monitor has no primary key value for deletion")
	}

	// 构建 SQL 删除查询
	query := fmt.Sprintf(
		"DELETE FROM \"token_monitor\" WHERE \"token_address\" = ?",
	)

	// 执行删除操作
	_, err := exec.ExecContext(ctx, query, o.TokenAddress)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to delete from token_monitor")
	}

	return nil
}

// tokenMonitorQuery 用于构建 TokenMonitor 记录的查询
type tokenMonitorQuery struct {
	*queries.Query
}

// TokenMonitorSlice 是指向 TokenMonitor 的指针切片别名
type TokenMonitorSlice []*TokenMonitor

// TokenMonitors 使用执行器检索所有记录
func TokenMonitors(mods ...qm.QueryMod) tokenMonitorQuery {
	mods = append(mods, qm.From("\"token_monitor\""))
	return tokenMonitorQuery{NewQuery(mods...)}
}

// All 从查询中返回所有 TokenMonitor 记录
func (q tokenMonitorQuery) All(ctx context.Context, exec boil.ContextExecutor) (TokenMonitorSlice, error) {
	var o TokenMonitorSlice

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "sqlite3: failed to assign all query results to TokenMonitor slice")
	}

	return o, nil
}
