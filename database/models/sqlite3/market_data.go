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

// MarketData 是表示 market_data 数据库表的结构体
type MarketData struct {
	ID             int       `boil:"id" json:"id" toml:"id" yaml:"id"`
	ExchangeID     int64     `boil:"exchange_id" json:"exchangeId" toml:"exchange_id" yaml:"exchange_id"`
	TradingPairID  int64     `boil:"trading_pair_id" json:"tradingPairId" toml:"trading_pair_id" yaml:"trading_pair_id"`
	Timestamp      time.Time `boil:"timestamp" json:"timestamp" toml:"timestamp" yaml:"timestamp"`
	LastPrice      float64   `boil:"last_price" json:"lastPrice" toml:"last_price" yaml:"last_price"`
	BidPrice       float64   `boil:"bid_price" json:"bidPrice" toml:"bid_price" yaml:"bid_price"`
	AskPrice       float64   `boil:"ask_price" json:"askPrice" toml:"ask_price" yaml:"ask_price"`
	Volume24h      float64   `boil:"volume_24h" json:"volume24h" toml:"volume_24h" yaml:"volume_24h"`
	High24h        float64   `boil:"high_24h" json:"high24h" toml:"high_24h" yaml:"high_24h"`
	Low24h         float64   `boil:"low_24h" json:"low24h" toml:"low_24h" yaml:"low_24h"`
	OpenPrice24h   float64   `boil:"open_price_24h" json:"openPrice24h" toml:"open_price_24h" yaml:"open_price_24h"`
	ClosePrice24h  float64   `boil:"close_price_24h" json:"closePrice24h" toml:"close_price_24h" yaml:"close_price_24h"`
	LiquidityUSD   float64   `boil:"liquidity_usd" json:"liquidityUSD" toml:"liquidity_usd" yaml:"liquidity_usd"`
	SlippageBPS    float64   `boil:"slippage_bps" json:"slippageBPS" toml:"slippage_bps" yaml:"slippage_bps"`
	SourceDataRaw  string    `boil:"source_data_raw" json:"sourceDataRaw" toml:"source_data_raw" yaml:"source_data_raw"`
}

// Insert 使用执行器插入单条记录
func (o *MarketData) Insert(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no market_data provided for insertion")
	}

	// 定义要插入的列，排除自增的 id
	marketDataColumns := []string{
		"exchange_id",
		"trading_pair_id",
		"timestamp",
		"last_price",
		"bid_price",
		"ask_price",
		"volume_24h",
		"high_24h",
		"low_24h",
		"open_price_24h",
		"close_price_24h",
		"liquidity_usd",
		"slippage_bps",
		"source_data_raw",
	}

	// 构建 SQL 查询
	query := fmt.Sprintf(
		"INSERT INTO \"market_data\" (\"%s\") VALUES (%s)",
		strings.Join(marketDataColumns, "\",\""),
		strmangle.Placeholders(dialect.UseIndexPlaceholders, len(marketDataColumns), 1, 1),
	)

	// 准备插入的值
	vals := []interface{}{
		o.ExchangeID,
		o.TradingPairID,
		o.Timestamp,
		o.LastPrice,
		o.BidPrice,
		o.AskPrice,
		o.Volume24h,
		o.High24h,
		o.Low24h,
		o.OpenPrice24h,
		o.ClosePrice24h,
		o.LiquidityUSD,
		o.SlippageBPS,
		o.SourceDataRaw,
	}

	// 执行插入操作
	result, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to insert into market_data")
	}

	// 获取自增 ID 并赋值给结构体
	id, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to get last insert id for market_data")
	}
	o.ID = int(id)

	return nil
}

// Update 更新数据库中的 MarketData 记录
func (o *MarketData) Update(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no market_data provided for update")
	}

	// 定义列信息
	marketDataAllColumns := []string{
		"id",
		"exchange_id",
		"trading_pair_id",
		"timestamp",
		"last_price",
		"bid_price",
		"ask_price",
		"volume_24h",
		"high_24h",
		"low_24h",
		"open_price_24h",
		"close_price_24h",
		"liquidity_usd",
		"slippage_bps",
		"source_data_raw",
	}
	marketDataPrimaryKeyColumns := []string{"id"}

	// 获取要更新的列（排除主键列）
	wl := make([]string, 0, len(marketDataAllColumns)-len(marketDataPrimaryKeyColumns))
	for _, col := range marketDataAllColumns {
		if col != "id" { // 排除主键
			wl = append(wl, col)
		}
	}

	if len(wl) == 0 {
		return errors.New("sqlite3: unable to update market_data, no columns to update")
	}

	// 构建SQL查询
	query := fmt.Sprintf(
		"UPDATE \"market_data\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, wl),
		strmangle.WhereClause("\"", "\"", len(wl)+1, marketDataPrimaryKeyColumns),
	)

	// 准备更新的值
	vals := make([]interface{}, len(wl)+len(marketDataPrimaryKeyColumns))

	// 添加更新列的值
	vals[0] = o.ExchangeID
	vals[1] = o.TradingPairID
	vals[2] = o.Timestamp
	vals[3] = o.LastPrice
	vals[4] = o.BidPrice
	vals[5] = o.AskPrice
	vals[6] = o.Volume24h
	vals[7] = o.High24h
	vals[8] = o.Low24h
	vals[9] = o.OpenPrice24h
	vals[10] = o.ClosePrice24h
	vals[11] = o.LiquidityUSD
	vals[12] = o.SlippageBPS
	vals[13] = o.SourceDataRaw

	// 添加主键值用于WHERE条件
	vals[len(wl)] = o.ID

	// 执行更新操作
	_, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to update market_data")
	}

	return nil
}

// Delete 从数据库中删除 MarketData 记录
func (o *MarketData) Delete(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no market_data provided for deletion")
	}

	// 检查主键 id 是否已设置
	if o.ID == 0 {
		return errors.New("sqlite3: market_data has no primary key value for deletion")
	}

	// 构建 SQL 删除查询
	query := fmt.Sprintf(
		"DELETE FROM \"market_data\" WHERE \"id\" = ?",
	)

	// 执行删除操作
	_, err := exec.ExecContext(ctx, query, o.ID)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to delete from market_data")
	}

	return nil
}

// marketDataQuery 用于构建 MarketData 记录的查询
type marketDataQuery struct {
	*queries.Query
}

// MarketDataSlice 是指向 MarketData 的指针切片别名
type MarketDataSlice []*MarketData

// MarketDatas 使用执行器检索所有记录
func MarketDatas(mods ...qm.QueryMod) marketDataQuery {
	mods = append(mods, qm.From("\"market_data\""))
	return marketDataQuery{NewQuery(mods...)}
}

// All 从查询中返回所有 MarketData 记录
func (q marketDataQuery) All(ctx context.Context, exec boil.ContextExecutor) (MarketDataSlice, error) {
	var o MarketDataSlice

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "sqlite3: failed to assign all query results to MarketData slice")
	}

	return o, nil
}