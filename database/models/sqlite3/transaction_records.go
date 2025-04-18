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

// TransactionRecord is an object representing the database table.
type TransactionRecord struct {
	TransactionHash   string          `boil:"transaction_hash" json:"transaction_hash" toml:"transaction_hash" yaml:"transaction_hash"`
	InputToken        string          `boil:"input_token" json:"input_token" toml:"input_token" yaml:"input_token"`
	InputAmount       float64         `boil:"input_amount" json:"input_amount" toml:"input_amount" yaml:"input_amount"`
	OutputToken       string          `boil:"output_token" json:"output_token" toml:"output_token" yaml:"output_token"`
	OutputAmount      float64         `boil:"output_amount" json:"output_amount" toml:"output_amount" yaml:"output_amount"`
	ReceivingAddress  string          `boil:"receiving_address" json:"receiving_address" toml:"receiving_address" yaml:"receiving_address"`
	Slippage          sql.NullFloat64 `boil:"slippage" json:"slippage" toml:"slippage" yaml:"slippage"`
	NetworkIdentifier string          `boil:"network_identifier" json:"network_identifier" toml:"network_identifier" yaml:"network_identifier"`
	TransactionTime   string          `boil:"transaction_time" json:"transaction_time" toml:"transaction_time" yaml:"transaction_time"`
	Price             float64         `boil:"price" json:"price" toml:"price" yaml:"price"`
	InitiatingAddress string          `boil:"initiating_address" json:"initiating_address" toml:"initiating_address" yaml:"initiating_address"`

	R *transactionRecordR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L transactionRecordL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *TransactionRecord) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("sqlite3: no transaction_records provided for insertion")
	}

	// 定义列信息
	transactionRecordAllColumns := []string{"transaction_hash", "input_token", "input_amount", "output_token", "output_amount", "receiving_address", "slippage", "network_identifier", "transaction_time", "price", "initiating_address"}
	transactionRecordColumnsWithDefault := []string{}
	transactionRecordColumnsWithoutDefault := []string{"transaction_hash", "input_token", "input_amount", "output_token", "output_amount", "receiving_address", "slippage", "network_identifier", "transaction_time", "price", "initiating_address"}

	// 获取要插入的列
	wl, _ := columns.InsertColumnSet(
		transactionRecordAllColumns,
		transactionRecordColumnsWithDefault,
		transactionRecordColumnsWithoutDefault,
		queries.NonZeroDefaultSet(transactionRecordColumnsWithDefault, o),
	)

	// 构建SQL查询
	var query string
	if len(wl) != 0 {
		query = fmt.Sprintf("INSERT INTO \"transaction_records\" (\"%s\") VALUES (%s)",
			strings.Join(wl, "\",\""),
			strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
	} else {
		query = "INSERT INTO \"transaction_records\" () VALUES ()"
	}

	// 直接从结构体获取值
	transactionRecordType := reflect.TypeOf(TransactionRecord{})
	value := reflect.Indirect(reflect.ValueOf(o))

	// 创建值映射
	vals := make([]interface{}, len(wl))
	for i, colName := range wl {
		field, _ := transactionRecordType.FieldByNameFunc(func(fieldName string) bool {
			field, _ := transactionRecordType.FieldByName(fieldName)
			tag := field.Tag.Get("boil")
			return tag == colName
		})
		vals[i] = value.FieldByIndex(field.Index).Interface()
	}

	// 执行插入
	_, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to insert into transaction_records")
	}

	return nil
}

// transactionRecordR is where relationships are stored.
type transactionRecordR struct {
}

// NewStruct creates a new relationship struct
func (*transactionRecordR) NewStruct() *transactionRecordR {
	return &transactionRecordR{}
}

// transactionRecordL is where Load methods for each relationship are stored.
type transactionRecordL struct{}

// TransactionRecordQuery is used to build up a query for TransactionRecord records
type transactionRecordQuery struct {
	*queries.Query
}

// TransactionRecordSlice is an alias for a slice of pointers to TransactionRecord
type TransactionRecordSlice []*TransactionRecord

// TransactionRecords retrieves all the records using an executor
func TransactionRecords(mods ...qm.QueryMod) transactionRecordQuery {
	mods = append(mods, qm.From("\"transaction_records\""))
	return transactionRecordQuery{NewQuery(mods...)}
}

// FindTransactionRecord retrieves a single record by TransactionHash with an executor.
// If selectCols is empty Find will return all columns.
func FindTransactionRecord(ctx context.Context, exec boil.ContextExecutor, transactionHash string, selectCols ...string) (*TransactionRecord, error) {
	transactionRecordObj := &TransactionRecord{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"transaction_records\" where \"transaction_hash\"=$1", sel,
	)

	q := queries.Raw(query, transactionHash)

	err := q.Bind(ctx, exec, transactionRecordObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到交易哈希为 %s 的交易记录", transactionHash)
		}
		return nil, fmt.Errorf("查询交易记录失败: %v", err)
	}

	return transactionRecordObj, nil
}

// FindTransactionRecordsByInitiatingAddress retrieves records by InitiatingAddress with an executor.
func FindTransactionRecordsByInitiatingAddress(ctx context.Context, exec boil.ContextExecutor, initiatingAddress string, selectCols ...string) (TransactionRecordSlice, error) {
	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"transaction_records\" where \"initiating_address\"=$1", sel,
	)

	q := queries.Raw(query, initiatingAddress)

	var result TransactionRecordSlice
	err := q.Bind(ctx, exec, &result)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到发起地址为 %s 的交易记录", initiatingAddress)
		}
		return nil, fmt.Errorf("查询交易记录失败: %v", err)
	}

	return result, nil
}

// FindTransactionRecordsByReceivingAddress retrieves records by ReceivingAddress with an executor.
func FindTransactionRecordsByReceivingAddress(ctx context.Context, exec boil.ContextExecutor, receivingAddress string, selectCols ...string) (TransactionRecordSlice, error) {
	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"transaction_records\" where \"receiving_address\"=$1", sel,
	)

	q := queries.Raw(query, receivingAddress)

	var result TransactionRecordSlice
	err := q.Bind(ctx, exec, &result)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到接收地址为 %s 的交易记录", receivingAddress)
		}
		return nil, fmt.Errorf("查询交易记录失败: %v", err)
	}

	return result, nil
}

// FindTransactionRecordsByNetwork retrieves records by NetworkIdentifier with an executor.
func FindTransactionRecordsByNetwork(ctx context.Context, exec boil.ContextExecutor, networkIdentifier string, selectCols ...string) (TransactionRecordSlice, error) {
	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"transaction_records\" where \"network_identifier\"=$1", sel,
	)

	q := queries.Raw(query, networkIdentifier)

	var result TransactionRecordSlice
	err := q.Bind(ctx, exec, &result)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到网络标识为 %s 的交易记录", networkIdentifier)
		}
		return nil, fmt.Errorf("查询交易记录失败: %v", err)
	}

	return result, nil
}

// FindTransactionRecordsByTimeRange retrieves records within a time range with an executor.
func FindTransactionRecordsByTimeRange(ctx context.Context, exec boil.ContextExecutor, startTime, endTime string, selectCols ...string) (TransactionRecordSlice, error) {
	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"transaction_records\" where \"transaction_time\" >= $1 AND \"transaction_time\" <= $2", sel,
	)

	q := queries.Raw(query, startTime, endTime)

	var result TransactionRecordSlice
	err := q.Bind(ctx, exec, &result)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到时间范围在 %s 至 %s 的交易记录", startTime, endTime)
		}
		return nil, fmt.Errorf("查询交易记录失败: %v", err)
	}

	return result, nil
}

// FindTransactionRecordsByToken retrieves records by input or output token with an executor.
func FindTransactionRecordsByToken(ctx context.Context, exec boil.ContextExecutor, token string, selectCols ...string) (TransactionRecordSlice, error) {
	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"transaction_records\" where \"input_token\"=$1 OR \"output_token\"=$1", sel,
	)

	q := queries.Raw(query, token)

	var result TransactionRecordSlice
	err := q.Bind(ctx, exec, &result)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到涉及代币 %s 的交易记录", token)
		}
		return nil, fmt.Errorf("查询交易记录失败: %v", err)
	}

	return result, nil
}

// All returns all TransactionRecord records from the query.
func (q transactionRecordQuery) All(ctx context.Context, exec boil.ContextExecutor) (TransactionRecordSlice, error) {
	var o TransactionRecordSlice

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, fmt.Errorf("查询所有交易记录失败: %v", err)
	}

	return o, nil
}

// One returns a single TransactionRecord record from the query.
func (q transactionRecordQuery) One(ctx context.Context, exec boil.ContextExecutor) (*TransactionRecord, error) {
	o := &TransactionRecord{}

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, fmt.Errorf("未找到符合条件的交易记录")
		}
		return nil, fmt.Errorf("查询单条交易记录失败: %v", err)
	}

	return o, nil
}

// Delete 删除指定交易哈希的记录
func (o *TransactionRecord) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("sqlite3: no transaction_records provided for delete")
	}

	query := "DELETE FROM \"transaction_records\" WHERE \"transaction_hash\"=?"
	result, err := exec.ExecContext(ctx, query, o.TransactionHash)
	if err != nil {
		return 0, errors.Wrap(err, "sqlite3: unable to delete from transaction_records")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlite3: failed to get rows affected")
	}

	return rowsAff, nil
}
