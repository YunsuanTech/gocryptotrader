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

// WatchlistToken 是代表数据库表的对象
type WatchlistToken struct {
	TokenID      int    `boil:"token_id" json:"token_id" toml:"token_id" yaml:"token_id"`
	TokenSymbol  string `boil:"token_symbol" json:"token_symbol" toml:"token_symbol" yaml:"token_symbol"`
	TokenAddress string `boil:"token_address" json:"token_address" toml:"token_address" yaml:"token_address"`
	Network      string `boil:"network" json:"network" toml:"network" yaml:"network"`
	Decimals     int    `boil:"decimals" json:"decimals" toml:"decimals" yaml:"decimals"`
	CreationTime int64  `boil:"creation_time" json:"creation_time" toml:"creation_time" yaml:"creation_time"`
	LastUpdated  int64  `boil:"last_updated" json:"last_updated" toml:"last_updated" yaml:"last_updated"`
	IsActive     int    `boil:"is_active" json:"is_active" toml:"is_active" yaml:"is_active"`

	R *watchlistTokenR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L watchlistTokenL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

// Insert 使用执行器插入单条记录
func (o *WatchlistToken) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("sqlite3: 没有提供要插入的watchlist_tokens")
	}

	// 定义列信息
	watchlistTokenAllColumns := []string{"token_id", "token_symbol", "token_address", "network", "decimals", "creation_time", "last_updated", "is_active"}
	watchlistTokenColumnsWithDefault := []string{"token_id", "network", "is_active"}
	watchlistTokenColumnsWithoutDefault := []string{"token_symbol", "token_address", "decimals", "creation_time", "last_updated"}

	// 获取要插入的列
	wl, _ := columns.InsertColumnSet(
		watchlistTokenAllColumns,
		watchlistTokenColumnsWithDefault,
		watchlistTokenColumnsWithoutDefault,
		queries.NonZeroDefaultSet(watchlistTokenColumnsWithDefault, o),
	)

	// 构建SQL查询
	var query string
	if len(wl) != 0 {
		query = fmt.Sprintf("INSERT INTO \"watchlist_tokens\" (\"%s\") VALUES (%s)",
			strings.Join(wl, "\",\""),
			strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
	} else {
		query = "INSERT INTO \"watchlist_tokens\" () VALUES ()"
	}

	// 直接从结构体获取值
	watchlistTokenType := reflect.TypeOf(WatchlistToken{})
	value := reflect.Indirect(reflect.ValueOf(o))

	// 创建值映射
	vals := make([]interface{}, len(wl))
	for i, colName := range wl {
		field, _ := watchlistTokenType.FieldByNameFunc(func(fieldName string) bool {
			field, _ := watchlistTokenType.FieldByName(fieldName)
			tag := field.Tag.Get("boil")
			return tag == colName
		})
		vals[i] = value.FieldByIndex(field.Index).Interface()
	}

	// 执行插入
	result, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: 无法插入到watchlist_tokens")
	}

	// 获取自增ID
	id, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "sqlite3: 无法获取最后插入的ID")
	}
	o.TokenID = int(id)

	return nil
}

// Update 使用执行器更新WatchlistToken记录
func (o *WatchlistToken) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("sqlite3: 没有提供要更新的watchlist_tokens")
	}

	// 定义列信息
	watchlistTokenAllColumns := []string{"token_id", "token_symbol", "token_address", "network", "decimals", "creation_time", "last_updated", "is_active"}
	watchlistTokenPrimaryKeyColumns := []string{"token_id"}

	// 获取要更新的列
	wl := columns.UpdateColumnSet(
		watchlistTokenAllColumns,
		watchlistTokenPrimaryKeyColumns,
	)

	if len(wl) == 0 {
		return errors.New("sqlite3: 无法更新watchlist_tokens，没有列")
	}

	// 构建SQL查询
	query := fmt.Sprintf(
		"UPDATE \"watchlist_tokens\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, wl),
		strmangle.WhereClause("\"", "\"", len(wl)+1, watchlistTokenPrimaryKeyColumns),
	)

	// 直接从结构体获取值
	watchlistTokenType := reflect.TypeOf(WatchlistToken{})
	value := reflect.Indirect(reflect.ValueOf(o))

	// 创建值映射
	vals := make([]interface{}, len(wl)+len(watchlistTokenPrimaryKeyColumns))
	for i, colName := range wl {
		field, _ := watchlistTokenType.FieldByNameFunc(func(fieldName string) bool {
			field, _ := watchlistTokenType.FieldByName(fieldName)
			tag := field.Tag.Get("boil")
			return tag == colName
		})
		vals[i] = value.FieldByIndex(field.Index).Interface()
	}

	// 添加主键值
	for i, colName := range watchlistTokenPrimaryKeyColumns {
		field, _ := watchlistTokenType.FieldByNameFunc(func(fieldName string) bool {
			field, _ := watchlistTokenType.FieldByName(fieldName)
			tag := field.Tag.Get("boil")
			return tag == colName
		})
		vals[len(wl)+i] = value.FieldByIndex(field.Index).Interface()
	}

	// 执行更新
	_, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: 无法更新watchlist_tokens")
	}

	return nil
}

// Delete 使用执行器删除WatchlistToken记录
func (o *WatchlistToken) Delete(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: 没有提供要删除的watchlist_tokens")
	}

	// 构建SQL查询
	query := "DELETE FROM \"watchlist_tokens\" WHERE \"token_id\"=$1"

	// 执行删除
	_, err := exec.ExecContext(ctx, query, o.TokenID)
	if err != nil {
		return errors.Wrap(err, "sqlite3: 无法删除watchlist_tokens记录")
	}

	return nil
}

// watchlistTokenR 是存储关系的地方
type watchlistTokenR struct {
}

// NewStruct 创建一个新的关系结构体
func (*watchlistTokenR) NewStruct() *watchlistTokenR {
	return &watchlistTokenR{}
}

// watchlistTokenL 是存储每个关系的Load方法的地方
type watchlistTokenL struct{}

// WatchlistTokenSlice 是指向WatchlistToken的指针切片的别名
type WatchlistTokenSlice []*WatchlistToken

// WatchlistTokens 使用执行器检索所有记录
func WatchlistTokens(mods ...qm.QueryMod) watchlistTokenQuery {
	mods = append(mods, qm.From("\"watchlist_tokens\""))
	return watchlistTokenQuery{NewQuery(mods...)}
}

// watchlistTokenQuery 用于构建WatchlistToken记录的查询
type watchlistTokenQuery struct {
	*queries.Query
}

// All 返回查询的所有记录
func (q watchlistTokenQuery) All(ctx context.Context, exec boil.ContextExecutor) (WatchlistTokenSlice, error) {
	var o []*WatchlistToken

	if err := q.Bind(ctx, exec, &o); err != nil {
		return nil, errors.Wrap(err, "sqlite3: 无法查询所有watchlist_tokens记录")
	}

	return o, nil
}

// FindWatchlistToken 通过TokenID使用执行器检索单个记录
func FindWatchlistToken(ctx context.Context, exec boil.ContextExecutor, tokenID int, selectCols ...string) (*WatchlistToken, error) {
	watchlistTokenObj := &WatchlistToken{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"watchlist_tokens\" where \"token_id\"=$1", sel,
	)

	q := queries.Raw(query, tokenID)

	err := q.Bind(ctx, exec, watchlistTokenObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到ID为 %d 的代币", tokenID)
		}
		return nil, fmt.Errorf("查询代币失败: %v", err)
	}

	return watchlistTokenObj, nil
}

// FindWatchlistTokenByAddress 通过TokenAddress使用执行器检索单个记录
func FindWatchlistTokenByAddress(ctx context.Context, exec boil.ContextExecutor, tokenAddress string, selectCols ...string) (*WatchlistToken, error) {
	watchlistTokenObj := &WatchlistToken{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"watchlist_tokens\" where \"token_address\"=$1", sel,
	)

	q := queries.Raw(query, tokenAddress)

	err := q.Bind(ctx, exec, watchlistTokenObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到地址为 %s 的代币", tokenAddress)
		}
		return nil, fmt.Errorf("查询代币失败: %v", err)
	}

	return watchlistTokenObj, nil
}

// FindWatchlistTokensBySymbol 通过TokenSymbol使用执行器检索记录
func FindWatchlistTokensBySymbol(ctx context.Context, exec boil.ContextExecutor, tokenSymbol string, selectCols ...string) (WatchlistTokenSlice, error) {
	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"watchlist_tokens\" where \"token_symbol\"=$1", sel,
	)

	q := queries.Raw(query, tokenSymbol)

	var result WatchlistTokenSlice
	err := q.Bind(ctx, exec, &result)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到符号为 %s 的代币", tokenSymbol)
		}
		return nil, fmt.Errorf("查询代币失败: %v", err)
	}

	return result, nil
}

// FindActiveWatchlistTokens 使用执行器检索所有活跃的记录
func FindActiveWatchlistTokens(ctx context.Context, exec boil.ContextExecutor, selectCols ...string) (WatchlistTokenSlice, error) {
	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"watchlist_tokens\" where \"is_active\"=1", sel,
	)

	q := queries.Raw(query)

	var result WatchlistTokenSlice
	err := q.Bind(ctx, exec, &result)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到活跃的代币")
		}
		return nil, fmt.Errorf("查询活跃代币失败: %v", err)
	}

	return result, nil
}

// DeleteWatchlistTokenByAddress 通过TokenAddress使用执行器删除记录
func DeleteWatchlistTokenByAddress(ctx context.Context, exec boil.ContextExecutor, tokenAddress string) error {
	query := "DELETE FROM \"watchlist_tokens\" WHERE \"token_address\"=$1"

	// 执行删除
	_, err := exec.ExecContext(ctx, query, tokenAddress)
	if err != nil {
		return errors.Wrap(err, "sqlite3: 无法删除watchlist_tokens记录")
	}

	return nil
}
