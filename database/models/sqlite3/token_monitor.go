package sqlite3

import (
	"context"
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
	TokenAddress   string  `boil:"token_address" json:"tokenAddress" toml:"token_address" yaml:"token_address"`
	TokenName      string  `boil:"token_name" json:"tokenName" toml:"token_name" yaml:"token_name"`
	Price          float64 `boil:"price" json:"price" toml:"price" yaml:"price"`
	TokenDecimals  uint8   `boil:"token_decimals" json:"tokenDecimals" toml:"token_decimals" yaml:"token_decimals"`
	BuyAmount      float64 `boil:"buy_amount" json:"buyAmount" toml:"buy_amount" yaml:"buy_amount"`
	BuyPrice       float64 `boil:"buy_price" json:"buyPrice" toml:"buy_price" yaml:"buy_price"`
	BuyTime        int64   `boil:"buy_time" json:"buyTime" toml:"buy_time" yaml:"buy_time"`
	SellPercentage float64 `boil:"sell_percentage" json:"sellPercentage" toml:"sell_percentage" yaml:"sell_percentage"`
	TotalSellPrice float64 `boil:"total_sell_price" json:"totalSellPrice" toml:"total_sell_price" yaml:"total_sell_price"`
	LastSellTime   float64 `boil:"last_sell_time" json:"lastSellTime" toml:"last_sell_time" yaml:"last_sell_time"`
	IsMonitoring   int     `boil:"is_monitoring" json:"isMonitoring" toml:"is_monitoring" yaml:"is_monitoring"`
	Amount         float64 `boil:"amount" json:"amount" toml:"amount" yaml:"amount"`
	Increase       float64 `boil:"increase" json:"increase" toml:"increase" yaml:"increase"`
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
		"amount",
		"buy_amount",
		"buy_price",
		"buy_time",
		"sell_percentage",
		"total_sell_price",
		"increase",
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
		o.Amount,
		o.BuyAmount,
		o.BuyPrice,
		o.BuyTime,
		o.SellPercentage,
		o.TotalSellPrice,
		o.Increase,
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

	// 定义列信息
	tokenMonitorAllColumns := []string{
		"token_address",
		"token_name",
		"price",
		"token_decimals",
		"amount",
		"buy_amount",
		"buy_price",
		"buy_time",
		"sell_percentage",
		"total_sell_price",
		"increase",
		"last_sell_time",
		"is_monitoring",
	}
	tokenMonitorPrimaryKeyColumns := []string{"token_address"}

	// 获取要更新的列（排除主键列）
	wl := make([]string, 0, len(tokenMonitorAllColumns)-len(tokenMonitorPrimaryKeyColumns))
	for _, col := range tokenMonitorAllColumns {
		if col != "token_address" { // 排除主键
			wl = append(wl, col)
		}
	}

	if len(wl) == 0 {
		return errors.New("sqlite3: unable to update token_monitor, no columns to update")
	}

	// 构建SQL查询
	query := fmt.Sprintf(
		"UPDATE \"token_monitor\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, wl),
		strmangle.WhereClause("\"", "\"", len(wl)+1, tokenMonitorPrimaryKeyColumns),
	)

	// 准备更新的值
	vals := make([]interface{}, len(wl)+len(tokenMonitorPrimaryKeyColumns))

	// 添加更新列的值
	vals[0] = o.TokenName
	vals[1] = o.Price
	vals[2] = o.TokenDecimals
	vals[3] = o.Amount
	vals[4] = o.BuyAmount
	vals[5] = o.BuyPrice
	vals[6] = o.BuyTime
	vals[7] = o.SellPercentage
	vals[8] = o.TotalSellPrice
	vals[9] = o.Increase
	vals[10] = o.LastSellTime
	vals[11] = o.IsMonitoring

	// 添加主键值用于WHERE条件
	vals[len(wl)] = o.TokenAddress

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
