package currency

import (
	"context"
	"errors"

	"gocryptotrader/database"
	modelSQLite "gocryptotrader/database/models/sqlite3"

	"github.com/thrasher-corp/sqlboiler/queries/qm"
)

// CurrencyQueryOptions 定义查询 Currency 的选项
type CurrencyQueryOptions struct {
	ID              int    // 根据 ID 查询
	Symbol          string // 根据货币符号查询
	Name            string // 根据货币名称查询
	ContractAddress string // 根据合约地址查询
	Chain           string // 根据区块链查询
	IsActive        *bool  // 根据是否激活查询
	Limit           int    // 返回记录的最大数量
}

// QueryCurrencies 根据提供的选项查询 Currency 记录
func QueryCurrencies(opts CurrencyQueryOptions) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}
	var mods []qm.QueryMod

	// 添加 ID 条件
	if opts.ID != 0 {
		mods = append(mods, qm.Where("id = ?", opts.ID))
	}
	// 添加 Symbol 条件
	if opts.Symbol != "" {
		mods = append(mods, qm.Where("symbol = ?", opts.Symbol))
	}
	// 添加 Name 条件
	if opts.Name != "" {
		mods = append(mods, qm.Where("name = ?", opts.Name))
	}
	// 添加 ContractAddress 条件
	if opts.ContractAddress != "" {
		mods = append(mods, qm.Where("contract_address = ?", opts.ContractAddress))
	}
	// 添加 Chain 条件
	if opts.Chain != "" {
		mods = append(mods, qm.Where("chain = ?", opts.Chain))
	}
	// 添加 IsActive 条件
	if opts.IsActive != nil {
		mods = append(mods, qm.Where("is_active = ?", *opts.IsActive))
	}

	// 默认按 symbol 排序
	mods = append(mods, qm.OrderBy("symbol"))

	// 设置查询限制
	if opts.Limit > 0 {
		mods = append(mods, qm.Limit(opts.Limit))
	}

	ctx := context.TODO()
	return modelSQLite.Currencies(mods...).All(ctx, database.DB.SQL)
}

// InsertCurrency 插入一条新的 Currency 记录
func InsertCurrency(ctx context.Context, c *modelSQLite.Currency) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 Currency 是否为 nil
	if c == nil {
		return errors.New("currency cannot be nil")
	}

	// 调用 Currency 的 Insert 方法执行插入
	err := c.Insert(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// UpdateCurrency 更新一条 Currency 记录
func UpdateCurrency(ctx context.Context, c *modelSQLite.Currency) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 Currency 是否为 nil
	if c == nil {
		return errors.New("currency cannot be nil")
	}

	// 调用 Currency 的 Update 方法执行更新
	err := c.Update(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// DeleteCurrency 删除一条 Currency 记录
func DeleteCurrency(ctx context.Context, c *modelSQLite.Currency) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 Currency 是否为 nil
	if c == nil {
		return errors.New("currency cannot be nil")
	}

	// 检查主键 ID 是否已设置
	if c.ID == 0 {
		return errors.New("currency ID cannot be 0")
	}

	// 调用 Currency 的 Delete 方法执行删除
	err := c.Delete(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// GetActiveCurrencies 获取所有激活的货币
func GetActiveCurrencies(ctx context.Context) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	// 添加查询条件
	mods := []qm.QueryMod{
		qm.Where("is_active = ?", true),
		qm.OrderBy("symbol"),
	}

	return modelSQLite.Currencies(mods...).All(ctx, database.DB.SQL)
}

// GetCurrencyBySymbol 根据货币符号获取货币信息
func GetCurrencyBySymbol(ctx context.Context, symbol string) (*modelSQLite.Currency, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	if symbol == "" {
		return nil, errors.New("symbol cannot be empty")
	}

	// 添加查询条件
	mods := []qm.QueryMod{
		qm.Where("symbol = ?", symbol),
		qm.Limit(1),
	}

	results, err := modelSQLite.Currencies(mods...).All(ctx, database.DB.SQL)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, errors.New("currency not found")
	}

	return results[0], nil
}