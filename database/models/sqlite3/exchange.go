package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries"
	"github.com/thrasher-corp/sqlboiler/queries/qm"
	"github.com/thrasher-corp/sqlboiler/strmangle"
)

// Exchange 是表示 exchanges 数据库表的结构体
type Exchange struct {
	ID             int            `boil:"id" json:"id" toml:"id" yaml:"id"`
	Name           string         `boil:"name" json:"name" toml:"name" yaml:"name"`
	Type           string         `boil:"type" json:"type" toml:"type" yaml:"type"`
	WebsiteURL     sql.NullString `boil:"website_url" json:"websiteUrl" toml:"website_url" yaml:"website_url"`
	APIBaseURL     sql.NullString `boil:"api_base_url" json:"apiBaseUrl" toml:"api_base_url" yaml:"api_base_url"`
	APIKeyRequired bool           `boil:"api_key_required" json:"apiKeyRequired" toml:"api_key_required" yaml:"api_key_required"`
	IsActive       bool           `boil:"is_active" json:"isActive" toml:"is_active" yaml:"is_active"`
	CreatedAt      time.Time      `boil:"created_at" json:"createdAt" toml:"created_at" yaml:"created_at"`
	UpdatedAt      time.Time      `boil:"updated_at" json:"updatedAt" toml:"updated_at" yaml:"updated_at"`
}

// Insert 使用执行器插入单条记录
func (o *Exchange) Insert(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no exchange provided for insertion")
	}

	// 定义要插入的列，排除自增的 id
	exchangeColumns := []string{
		"name",
		"type",
		"website_url",
		"api_base_url",
		"api_key_required",
		"is_active",
		"created_at",
		"updated_at",
	}

	// 构建 SQL 查询
	query := fmt.Sprintf(
		"INSERT INTO \"exchanges\" (\"%s\") VALUES (%s)",
		strings.Join(exchangeColumns, "\",\""),
		strmangle.Placeholders(dialect.UseIndexPlaceholders, len(exchangeColumns), 1, 1),
	)

	// 准备插入的值
	vals := []interface{}{
		o.Name,
		o.Type,
		o.WebsiteURL,
		o.APIBaseURL,
		o.APIKeyRequired,
		o.IsActive,
		o.CreatedAt,
		o.UpdatedAt,
	}

	// 执行插入操作
	result, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to insert into exchanges")
	}

	// 获取自增 ID 并赋值给结构体
	id, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to get last insert id for exchanges")
	}
	o.ID = int(id)

	return nil
}

// Update 更新数据库中的 Exchange 记录
func (o *Exchange) Update(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no exchange provided for update")
	}

	// 定义列信息
	exchangeAllColumns := []string{
		"id",
		"name",
		"type",
		"website_url",
		"api_base_url",
		"api_key_required",
		"is_active",
		"created_at",
		"updated_at",
	}
	exchangePrimaryKeyColumns := []string{"id"}

	// 获取要更新的列（排除主键列）
	wl := make([]string, 0, len(exchangeAllColumns)-len(exchangePrimaryKeyColumns))
	for _, col := range exchangeAllColumns {
		if col != "id" { // 排除主键
			wl = append(wl, col)
		}
	}

	if len(wl) == 0 {
		return errors.New("sqlite3: unable to update exchanges, no columns to update")
	}

	// 构建SQL查询
	query := fmt.Sprintf(
		"UPDATE \"exchanges\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, wl),
		strmangle.WhereClause("\"", "\"", len(wl)+1, exchangePrimaryKeyColumns),
	)

	// 准备更新的值
	vals := make([]interface{}, len(wl)+len(exchangePrimaryKeyColumns))

	// 添加更新列的值
	vals[0] = o.Name
	vals[1] = o.Type
	vals[2] = o.WebsiteURL
	vals[3] = o.APIBaseURL
	vals[4] = o.APIKeyRequired
	vals[5] = o.IsActive
	vals[6] = o.CreatedAt
	vals[7] = o.UpdatedAt

	// 添加主键值用于WHERE条件
	vals[len(wl)] = o.ID

	// 执行更新操作
	_, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to update exchanges")
	}

	return nil
}

// Delete 从数据库中删除 Exchange 记录
func (o *Exchange) Delete(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no exchange provided for deletion")
	}

	// 检查主键 id 是否已设置
	if o.ID == 0 {
		return errors.New("sqlite3: exchange has no primary key value for deletion")
	}

	// 构建 SQL 删除查询
	query := fmt.Sprintf(
		"DELETE FROM \"exchanges\" WHERE \"id\" = ?",
	)

	// 执行删除操作
	_, err := exec.ExecContext(ctx, query, o.ID)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to delete from exchanges")
	}

	return nil
}

// exchangeQuery 用于构建 Exchange 记录的查询
type exchangeQuery struct {
	*queries.Query
}

// ExchangeSlice 是指向 Exchange 的指针切片别名
type ExchangeSlice []*Exchange

// Exchanges 使用执行器检索所有记录
func Exchanges(mods ...qm.QueryMod) exchangeQuery {
	mods = append(mods, qm.From("\"exchanges\""))
	return exchangeQuery{NewQuery(mods...)}
}

// All 从查询中返回所有 Exchange 记录
func (q exchangeQuery) All(ctx context.Context, exec boil.ContextExecutor) (ExchangeSlice, error) {
	var o ExchangeSlice

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "sqlite3: failed to assign all query results to Exchange slice")
	}

	return o, nil
}

// One 从查询中返回一条 Exchange 记录
func (q exchangeQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Exchange, error) {
	o := &Exchange{}

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlite3: failed to assign one query result to Exchange")
	}

	return o, nil
}