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

// Xen is an object representing the database table.
type Xen struct {
	Slot           int             `boil:"slot" json:"slot" toml:"slot" yaml:"slot"`
	ChainName      string          `boil:"chain_name" json:"chain_name" toml:"chain_name" yaml:"chain_name"`
	Count          int             `boil:"count" json:"count" toml:"count" yaml:"count"`
	Days           int             `boil:"days" json:"days" toml:"days" yaml:"days"`
	ExecutionTime  sql.NullTime    `boil:"execution_time" json:"execution_time" toml:"execution_time" yaml:"execution_time"`
	ClaimTime      sql.NullTime    `boil:"claim_time" json:"claim_time" toml:"claim_time" yaml:"claim_time"`
	ExpectedReward sql.NullFloat64 `boil:"expected_reward" json:"expected_reward" toml:"expected_reward" yaml:"expected_reward"`
	Ranking        int             `boil:"ranking" json:"ranking" toml:"ranking" yaml:"ranking"`
	Amp            int             `boil:"amp" json:"amp" toml:"amp" yaml:"amp"`
	Eaa            float64         `boil:"eaa" json:"eaa" toml:"eaa" yaml:"eaa"`
	M              sql.NullFloat64 `boil:"m" json:"m" toml:"m" yaml:"m"`
	Status         string          `boil:"status" json:"status" toml:"status" yaml:"status"`
	TxID           sql.NullString  `boil:"tx_id" json:"tx_id" toml:"tx_id" yaml:"tx_id"`
	MintFees       sql.NullFloat64 `boil:"mint_fees" json:"mint_fees" toml:"mint_fees" yaml:"mint_fees"`
	ClaimFees      sql.NullFloat64 `boil:"claim_fees" json:"claim_fees" toml:"claim_fees" yaml:"claim_fees"`

	R *xenR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L xenL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Xen) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("sqlite3: no xens provided for insertion")
	}

	// 定义列信息
	xenAllColumns := []string{"slot", "chain_name", "count", "days", "execution_time", "claim_time", "expected_reward", "ranking", "amp", "eaa", "m", "status", "tx_id", "mint_fees", "claim_fees"}
	xenColumnsWithDefault := []string{}
	xenColumnsWithoutDefault := []string{"slot", "chain_name", "count", "days", "execution_time", "claim_time", "expected_reward", "ranking", "amp", "eaa", "m", "status", "tx_id", "mint_fees", "claim_fees"}

	// 获取要插入的列
	wl, _ := columns.InsertColumnSet(
		xenAllColumns,
		xenColumnsWithDefault,
		xenColumnsWithoutDefault,
		queries.NonZeroDefaultSet(xenColumnsWithDefault, o),
	)

	// 构建SQL查询
	var query string
	if len(wl) != 0 {
		query = fmt.Sprintf("INSERT INTO \"xens\" (\"%s\") VALUES (%s)",
			strings.Join(wl, "\",\""),
			strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
	} else {
		query = "INSERT INTO \"xens\" () VALUES ()"
	}

	// 直接从结构体获取值
	xenType := reflect.TypeOf(Xen{})
	value := reflect.Indirect(reflect.ValueOf(o))

	// 创建值映射
	vals := make([]interface{}, len(wl))
	for i, colName := range wl {
		field, _ := xenType.FieldByNameFunc(func(fieldName string) bool {
			field, _ := xenType.FieldByName(fieldName)
			tag := field.Tag.Get("boil")
			return tag == colName
		})
		vals[i] = value.FieldByIndex(field.Index).Interface()
	}

	// 执行插入
	_, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to insert into xens")
	}

	return nil
}

// Update 使用执行器更新Xen记录
func (o *Xen) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("sqlite3: no xens provided for update")
	}

	// 定义列信息
	xenAllColumns := []string{"slot", "chain_name", "count", "days", "execution_time", "claim_time", "expected_reward", "ranking", "amp", "eaa", "m", "status", "tx_id", "mint_fees", "claim_fees"}
	xenPrimaryKeyColumns := []string{"slot", "chain_name"}

	// 获取要更新的列
	wl := columns.UpdateColumnSet(
		xenAllColumns,
		xenPrimaryKeyColumns,
	)

	if len(wl) == 0 {
		return errors.New("sqlite3: unable to update xens, no columns to update")
	}

	// 构建SQL查询
	query := fmt.Sprintf(
		"UPDATE \"xens\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, wl),
		strmangle.WhereClause("\"", "\"", len(wl)+1, xenPrimaryKeyColumns),
	)

	// 直接从结构体获取值
	xenType := reflect.TypeOf(Xen{})
	value := reflect.Indirect(reflect.ValueOf(o))

	// 创建值映射
	vals := make([]interface{}, len(wl)+len(xenPrimaryKeyColumns))
	for i, colName := range wl {
		field, _ := xenType.FieldByNameFunc(func(fieldName string) bool {
			field, _ := xenType.FieldByName(fieldName)
			tag := field.Tag.Get("boil")
			return tag == colName
		})
		vals[i] = value.FieldByIndex(field.Index).Interface()
	}

	// 添加主键值
	for i, colName := range xenPrimaryKeyColumns {
		field, _ := xenType.FieldByNameFunc(func(fieldName string) bool {
			field, _ := xenType.FieldByName(fieldName)
			tag := field.Tag.Get("boil")
			return tag == colName
		})
		vals[len(wl)+i] = value.FieldByIndex(field.Index).Interface()
	}

	// 执行更新
	_, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to update xens")
	}

	return nil
}

// xenR is where relationships are stored.
type xenR struct {
}

// NewStruct creates a new relationship struct
func (*xenR) NewStruct() *xenR {
	return &xenR{}
}

// xenL is where Load methods for each relationship are stored.
type xenL struct{}

// xenQuery is used to build up a query for Xen records
type xenQuery struct {
	*queries.Query
}

// XenSlice is an alias for a slice of pointers to Xen
type XenSlice []*Xen

// Xens retrieves all the records using an executor
func Xens(mods ...qm.QueryMod) xenQuery {
	mods = append(mods, qm.From("\"xens\""))
	return xenQuery{NewQuery(mods...)}
}

// FindXen retrieves a single record by primary key with an executor.
// If selectCols is empty Find will return all columns.
func FindXen(ctx context.Context, exec boil.ContextExecutor, slot int, chainName string, selectCols ...string) (*Xen, error) {
	xenObj := &Xen{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"xens\" where \"slot\"=$1 AND \"chain_name\"=$2", sel,
	)

	q := queries.Raw(query, slot, chainName)

	err := q.Bind(ctx, exec, xenObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlite3: unable to select from xens")
	}

	return xenObj, nil
}

// All returns all Xen records from the query.
func (q xenQuery) All(ctx context.Context, exec boil.ContextExecutor) (XenSlice, error) {
	var o XenSlice

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "sqlite3: failed to assign all query results to Xen slice")
	}

	return o, nil
}

// One returns a single Xen record from the query.
func (q xenQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Xen, error) {
	o := &Xen{}

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlite3: failed to assign one query result to Xen")
	}

	return o, nil
}
