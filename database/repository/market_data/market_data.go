package marketdata

import (
	"context"
	"errors"

	"gocryptotrader/database"
	modelSQLite "gocryptotrader/database/models/sqlite3"

	"github.com/thrasher-corp/sqlboiler/queries/qm"
)

// MarketDataQueryOptions 定义查询 MarketData 的选项
type MarketDataQueryOptions struct {
	ID            int       // 根据 ID 查询
	ExchangeID    int64     // 根据交易所ID查询
	TradingPairID int64     // 根据交易对ID查询
	MinLastPrice  float64   // 根据最低最新价格查询
	MaxLastPrice  float64   // 根据最高最新价格查询
	MinBidPrice   float64   // 根据最低买入价格查询
	MaxBidPrice   float64   // 根据最高买入价格查询
	MinAskPrice   float64   // 根据最低卖出价格查询
	MaxAskPrice   float64   // 根据最高卖出价格查询
	MinVolume24h  float64   // 根据最低24小时交易量查询
	MaxVolume24h  float64   // 根据最高24小时交易量查询
	Limit         int       // 返回记录的最大数量
}

// QueryMarketData 根据提供的选项查询 MarketData 记录
func QueryMarketData(opts MarketDataQueryOptions) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}
	var mods []qm.QueryMod

	// 添加 ID 条件
	if opts.ID != 0 {
		mods = append(mods, qm.Where("id = ?", opts.ID))
	}
	// 添加 ExchangeID 条件
	if opts.ExchangeID != 0 {
		mods = append(mods, qm.Where("exchange_id = ?", opts.ExchangeID))
	}
	// 添加 TradingPairID 条件
	if opts.TradingPairID != 0 {
		mods = append(mods, qm.Where("trading_pair_id = ?", opts.TradingPairID))
	}
	// 添加最新价格范围条件
	if opts.MinLastPrice > 0 {
		mods = append(mods, qm.Where("last_price >= ?", opts.MinLastPrice))
	}
	if opts.MaxLastPrice > 0 {
		mods = append(mods, qm.Where("last_price <= ?", opts.MaxLastPrice))
	}
	// 添加买入价格范围条件
	if opts.MinBidPrice > 0 {
		mods = append(mods, qm.Where("bid_price >= ?", opts.MinBidPrice))
	}
	if opts.MaxBidPrice > 0 {
		mods = append(mods, qm.Where("bid_price <= ?", opts.MaxBidPrice))
	}
	// 添加卖出价格范围条件
	if opts.MinAskPrice > 0 {
		mods = append(mods, qm.Where("ask_price >= ?", opts.MinAskPrice))
	}
	if opts.MaxAskPrice > 0 {
		mods = append(mods, qm.Where("ask_price <= ?", opts.MaxAskPrice))
	}
	// 添加交易量范围条件
	if opts.MinVolume24h > 0 {
		mods = append(mods, qm.Where("volume_24h >= ?", opts.MinVolume24h))
	}
	if opts.MaxVolume24h > 0 {
		mods = append(mods, qm.Where("volume_24h <= ?", opts.MaxVolume24h))
	}

	// 默认按 exchange_id 和 trading_pair_id 排序
	mods = append(mods, qm.OrderBy("exchange_id, trading_pair_id"))

	// 设置查询限制
	if opts.Limit > 0 {
		mods = append(mods, qm.Limit(opts.Limit))
	}

	ctx := context.TODO()
	return modelSQLite.MarketDatas(mods...).All(ctx, database.DB.SQL)
}

// InsertMarketData 插入一条新的 MarketData 记录
func InsertMarketData(ctx context.Context, md *modelSQLite.MarketData) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 MarketData 是否为 nil
	if md == nil {
		return errors.New("market data cannot be nil")
	}

	// 调用 MarketData 的 Insert 方法执行插入
	err := md.Insert(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// UpdateMarketData 更新一条 MarketData 记录
func UpdateMarketData(ctx context.Context, md *modelSQLite.MarketData) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 MarketData 是否为 nil
	if md == nil {
		return errors.New("market data cannot be nil")
	}

	// 调用 MarketData 的 Update 方法执行更新
	err := md.Update(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// DeleteMarketData 删除一条 MarketData 记录
func DeleteMarketData(ctx context.Context, md *modelSQLite.MarketData) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 MarketData 是否为 nil
	if md == nil {
		return errors.New("market data cannot be nil")
	}

	// 检查主键 ID 是否已设置
	if md.ID == 0 {
		return errors.New("market data ID cannot be 0")
	}

	// 调用 MarketData 的 Delete 方法执行删除
	err := md.Delete(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// GetLatestMarketData 获取每个交易所和交易对的最新市场数据
func GetLatestMarketData(ctx context.Context) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	// 使用子查询获取每个交易所和交易对的最新记录
	subQuery := qm.SQL("id IN (SELECT MAX(id) FROM market_data GROUP BY exchange_id, trading_pair_id)")

	// 添加排序
	mods := []qm.QueryMod{
		subQuery,
		qm.OrderBy("exchange_id, trading_pair_id"),
	}

	return modelSQLite.MarketDatas(mods...).All(ctx, database.DB.SQL)
}