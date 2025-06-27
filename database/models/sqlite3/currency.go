package sqlite3

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries"
	"github.com/thrasher-corp/sqlboiler/queries/qm"
	"github.com/thrasher-corp/sqlboiler/strmangle"
)

// Currency 是表示 currencies 数据库表的结构体
type Currency struct {
	ID              int       `boil:"id" json:"id" toml:"id" yaml:"id"`
	Symbol          string    `boil:"symbol" json:"symbol" toml:"symbol" yaml:"symbol"`
	Name            string    `boil:"name" json:"name" toml:"name" yaml:"name"`
	Decimals        int       `boil:"decimals" json:"decimals" toml:"decimals" yaml:"decimals"`
	ContractAddress string    `boil:"contract_address" json:"contractAddress" toml:"contract_address" yaml:"contract_address"`
	Chain           string    `boil:"chain" json:"chain" toml:"chain" yaml:"chain"`
	IsActive        bool      `boil:"is_active" json:"isActive" toml:"is_active" yaml:"is_active"`
	CreatedAt       time.Time `boil:"created_at" json:"createdAt" toml:"created_at" yaml:"created_at"`
	UpdatedAt       time.Time `boil:"updated_at" json:"updatedAt" toml:"updated_at" yaml:"updated_at"`
}

// Insert 使用执行器插入单条记录
func (o *Currency) Insert(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no currency provided for insertion")
	}

	// 定义要插入的列，排除自增的 id
	currencyColumns := []string{
		"symbol",
		"name",
		"decimals",
		"contract_address",
		"chain",
		"is_active",
		"created_at",
		"updated_at",
	}

	// 构建 SQL 查询
	query := fmt.Sprintf(
		"INSERT INTO \"currencies\" (\"%s\") VALUES (%s)",
		strings.Join(currencyColumns, "\",\""),
		strmangle.Placeholders(dialect.UseIndexPlaceholders, len(currencyColumns), 1, 1),
	)

	// 准备插入的值
	vals := []interface{}{
		o.Symbol,
		o.Name,
		o.Decimals,
		o.ContractAddress,
		o.Chain,
		o.IsActive,
		o.CreatedAt,
		o.UpdatedAt,
	}

	// 执行插入操作
	result, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to insert into currencies")
	}

	// 获取自增 ID 并赋值给结构体
	id, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to get last insert id for currencies")
	}
	o.ID = int(id)

	return nil
}

// Update 更新数据库中的 Currency 记录
func (o *Currency) Update(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no currency provided for update")
	}

	// 定义列信息
	currencyAllColumns := []string{
		"id",
		"symbol",
		"name",
		"decimals",
		"contract_address",
		"chain",
		"is_active",
		"created_at",
		"updated_at",
	}
	currencyPrimaryKeyColumns := []string{"id"}

	// 获取要更新的列（排除主键列）
	wl := make([]string, 0, len(currencyAllColumns)-len(currencyPrimaryKeyColumns))
	for _, col := range currencyAllColumns {
		if col != "id" { // 排除主键
			wl = append(wl, col)
		}
	}

	if len(wl) == 0 {
		return errors.New("sqlite3: unable to update currencies, no columns to update")
	}

	// 构建SQL查询
	query := fmt.Sprintf(
		"UPDATE \"currencies\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, wl),
		strmangle.WhereClause("\"", "\"", len(wl)+1, currencyPrimaryKeyColumns),
	)

	// 准备更新的值
	vals := make([]interface{}, len(wl)+len(currencyPrimaryKeyColumns))

	// 添加更新列的值
	vals[0] = o.Symbol
	vals[1] = o.Name
	vals[2] = o.Decimals
	vals[3] = o.ContractAddress
	vals[4] = o.Chain
	vals[5] = o.IsActive
	vals[6] = o.CreatedAt
	vals[7] = o.UpdatedAt

	// 添加主键值用于WHERE条件
	vals[len(wl)] = o.ID

	// 执行更新操作
	_, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to update currencies")
	}

	return nil
}

// Delete 从数据库中删除 Currency 记录
func (o *Currency) Delete(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no currency provided for deletion")
	}

	// 检查主键 id 是否已设置
	if o.ID == 0 {
		return errors.New("sqlite3: currency has no primary key value for deletion")
	}

	// 构建 SQL 删除查询
	query := fmt.Sprintf(
		"DELETE FROM \"currencies\" WHERE \"id\" = ?",
	)

	// 执行删除操作
	_, err := exec.ExecContext(ctx, query, o.ID)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to delete from currencies")
	}

	return nil
}

// currencyQuery 用于构建 Currency 记录的查询
type currencyQuery struct {
	*queries.Query
}

// CurrencySlice 是指向 Currency 的指针切片别名
type CurrencySlice []*Currency

// Currencies 使用执行器检索所有记录
func Currencies(mods ...qm.QueryMod) currencyQuery {
	mods = append(mods, qm.From("\"currencies\""))
	return currencyQuery{NewQuery(mods...)}
}

// All 从查询中返回所有 Currency 记录
func (q currencyQuery) All(ctx context.Context, exec boil.ContextExecutor) (CurrencySlice, error) {
	var o CurrencySlice

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "sqlite3: failed to assign all query results to Currency slice")
	}

	return o, nil
}