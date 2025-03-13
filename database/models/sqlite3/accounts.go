package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries"
	"github.com/thrasher-corp/sqlboiler/queries/qm"
	"github.com/thrasher-corp/sqlboiler/strmangle"
)

// Account is an object representing the database table.
type Account struct {
	ID                int       `boil:"id" json:"id" toml:"id" yaml:"id"`
	Name              string    `boil:"name" json:"name" toml:"name" yaml:"name"`
	Address           string    `boil:"address" json:"address" toml:"address" yaml:"address"`
	ExchangeAddressID string    `boil:"exchange_address_id" json:"exchange_address_id" toml:"exchange_address_id" yaml:"exchange_address_id"`
	ZkAddressID       string    `boil:"zk_address_id" json:"zk_address_id" toml:"zk_address_id" yaml:"zk_address_id"`
	F4AddressID       string    `boil:"f4_address_id" json:"f4_address_id" toml:"f4_address_id" yaml:"f4_address_id"`
	OTAddressID       string    `boil:"ot_address_id" json:"ot_address_id" toml:"ot_address_id" yaml:"ot_address_id"`
	Cipher            string    `boil:"cipher" json:"cipher" toml:"cipher" yaml:"cipher"`
	Layer             int       `boil:"layer" json:"layer" toml:"layer" yaml:"layer"`
	Owner             string    `boil:"owner" json:"owner" toml:"owner" yaml:"owner"`
	ChainName         string    `boil:"chain_name" json:"chain_name" toml:"chain_name" yaml:"chain_name"`
	CreatedAt         time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt         time.Time `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`

	R *accountR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L accountL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Account) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("sqlite3: no accounts provided for insertion")
	}

	// 定义列信息
	accountAllColumns := []string{"id", "name", "address", "exchange_address_id", "zk_address_id", "f4_address_id", "ot_address_id", "cipher", "layer", "owner", "chain_name", "created_at", "updated_at"}
	accountColumnsWithDefault := []string{"created_at", "updated_at"}
	accountColumnsWithoutDefault := []string{"name", "address", "exchange_address_id", "zk_address_id", "f4_address_id", "ot_address_id", "cipher", "layer", "owner", "chain_name"}

	// 获取要插入的列
	wl, returnColumns := columns.InsertColumnSet(
		accountAllColumns,
		accountColumnsWithDefault,
		accountColumnsWithoutDefault,
		queries.NonZeroDefaultSet(accountColumnsWithDefault, o),
	)

	// 构建SQL查询
	var query string
	if len(wl) != 0 {
		query = fmt.Sprintf("INSERT INTO \"accounts\" (\"%s\") VALUES (%s)",
			strings.Join(wl, "\",\""),
			strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
	} else {
		query = "INSERT INTO \"accounts\" () VALUES ()"
	}

	// 直接从结构体获取值
	accountType := reflect.TypeOf(Account{})
	value := reflect.Indirect(reflect.ValueOf(o))

	// 创建值映射
	vals := make([]interface{}, len(wl))
	for i, colName := range wl {
		field, _ := accountType.FieldByNameFunc(func(fieldName string) bool {
			field, _ := accountType.FieldByName(fieldName)
			tag := field.Tag.Get("boil")
			return tag == colName
		})
		vals[i] = value.FieldByIndex(field.Index).Interface()
	}

	// 执行插入
	result, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to insert into account")
	}

	// 如果需要获取返回值（自增ID或默认值）
	if len(returnColumns) > 0 {
		// 对于自增ID，尝试获取
		if o.ID == 0 {
			id, err := result.LastInsertId()
			if err == nil {
				o.ID = int(id)
			}
		}

	}

	return nil
}

// accountR is where relationships are stored.
type accountR struct {
}

// NewStruct creates a new relationship struct
func (*accountR) NewStruct() *accountR {
	return &accountR{}
}

// accountL is where Load methods for each relationship are stored.
type accountL struct{}

// AccountQuery is used to build up a query for Account records
type accountQuery struct {
	*queries.Query
}

// AccountSlice is an alias for a slice of pointers to Account
type AccountSlice []*Account

// Accounts retrieves all the records using an executor
func Accounts(mods ...qm.QueryMod) accountQuery {
	mods = append(mods, qm.From("\"accounts\""))
	return accountQuery{NewQuery(mods...)}
}

// FindAccount retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindAccount(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*Account, error) {
	accountObj := &Account{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"accounts\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, accountObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "postgres: unable to select from account")
	}

	return accountObj, nil
}

// FindAccountByName retrieves a single record by Name with an executor.
// If selectCols is empty Find will return all columns.
func FindAccountByName(ctx context.Context, exec boil.ContextExecutor, name string, selectCols ...string) (*Account, error) {
	accountObj := &Account{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"accounts\" where \"name\"=$1", sel,
	)

	q := queries.Raw(query, name)

	err := q.Bind(ctx, exec, accountObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "postgres: unable to select from account")
	}

	return accountObj, nil
}

// All returns all Account records from the query.
func (q accountQuery) All(ctx context.Context, exec boil.ContextExecutor) (AccountSlice, error) {
	var o AccountSlice

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "postgres: failed to assign all query results to Account slice")
	}

	return o, nil
}

// One returns a single Account record from the query.
func (q accountQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Account, error) {
	o := &Account{}

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "postgres: failed to assign one query result to Account")
	}

	return o, nil
}
