package exchange

import (
	"context"
	"errors"

	"gocryptotrader/database"
	modelSQLite "gocryptotrader/database/models/sqlite3"

	"github.com/thrasher-corp/sqlboiler/queries/qm"
)

// ExchangeQueryOptions 定义查询 Exchange 的选项
type ExchangeQueryOptions struct {
	ID             int     // 根据 ID 查询
	Name           string  // 根据交易所名称查询
	Type           string  // 根据交易所类型查询 (CEX, DEX, Aggregator)
	APIKeyRequired *bool   // 根据是否需要API密钥查询
	IsActive       *bool   // 根据是否激活查询
	Limit          int     // 返回记录的最大数量
}

// QueryExchanges 根据提供的选项查询 Exchange 记录
func QueryExchanges(opts ExchangeQueryOptions) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}
	var mods []qm.QueryMod

	// 添加 ID 条件
	if opts.ID != 0 {
		mods = append(mods, qm.Where("id = ?", opts.ID))
	}
	// 添加 Name 条件
	if opts.Name != "" {
		mods = append(mods, qm.Where("name = ?", opts.Name))
	}
	// 添加 Type 条件
	if opts.Type != "" {
		mods = append(mods, qm.Where("type = ?", opts.Type))
	}
	// 添加 APIKeyRequired 条件
	if opts.APIKeyRequired != nil {
		mods = append(mods, qm.Where("api_key_required = ?", *opts.APIKeyRequired))
	}
	// 添加 IsActive 条件
	if opts.IsActive != nil {
		mods = append(mods, qm.Where("is_active = ?", *opts.IsActive))
	}

	// 默认按 name 排序
	mods = append(mods, qm.OrderBy("name"))

	// 设置查询限制
	if opts.Limit > 0 {
		mods = append(mods, qm.Limit(opts.Limit))
	}

	ctx := context.TODO()
	return modelSQLite.Exchanges(mods...).All(ctx, database.DB.SQL)
}

// InsertExchange 插入一条新的 Exchange 记录
func InsertExchange(ctx context.Context, e *modelSQLite.Exchange) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 Exchange 是否为 nil
	if e == nil {
		return errors.New("exchange cannot be nil")
	}

	// 调用 Exchange 的 Insert 方法执行插入
	err := e.Insert(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// UpdateExchange 更新一条 Exchange 记录
func UpdateExchange(ctx context.Context, e *modelSQLite.Exchange) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 Exchange 是否为 nil
	if e == nil {
		return errors.New("exchange cannot be nil")
	}

	// 调用 Exchange 的 Update 方法执行更新
	err := e.Update(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// DeleteExchange 删除一条 Exchange 记录
func DeleteExchange(ctx context.Context, e *modelSQLite.Exchange) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 Exchange 是否为 nil
	if e == nil {
		return errors.New("exchange cannot be nil")
	}

	// 检查主键 ID 是否已设置
	if e.ID == 0 {
		return errors.New("exchange ID cannot be 0")
	}

	// 调用 Exchange 的 Delete 方法执行删除
	err := e.Delete(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// GetActiveExchanges 获取所有激活的交易所
func GetActiveExchanges(ctx context.Context) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	// 添加查询条件
	mods := []qm.QueryMod{
		qm.Where("is_active = ?", true),
		qm.OrderBy("name"),
	}

	return modelSQLite.Exchanges(mods...).All(ctx, database.DB.SQL)
}

// GetExchangeByName 根据交易所名称获取交易所信息
func GetExchangeByName(ctx context.Context, name string) (*modelSQLite.Exchange, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	// 添加查询条件
	mods := []qm.QueryMod{
		qm.Where("name = ?", name),
		qm.Limit(1),
	}

	results, err := modelSQLite.Exchanges(mods...).All(ctx, database.DB.SQL)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, errors.New("exchange not found")
	}

	return results[0], nil
}

// GetExchangesByType 根据交易所类型获取交易所列表
func GetExchangesByType(ctx context.Context, exchangeType string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	if exchangeType == "" {
		return nil, errors.New("exchange type cannot be empty")
	}

	// 添加查询条件
	mods := []qm.QueryMod{
		qm.Where("type = ?", exchangeType),
		qm.OrderBy("name"),
	}

	return modelSQLite.Exchanges(mods...).All(ctx, database.DB.SQL)
}