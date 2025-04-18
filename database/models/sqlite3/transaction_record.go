package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries"
	"github.com/thrasher-corp/sqlboiler/queries/qm"
	"github.com/thrasher-corp/sqlboiler/strmangle"
)

// TransactionRecord 是表示 transaction_records 数据库表的结构体
type TransactionRecord struct {
	TransactionID int            `boil:"transaction_id" json:"transaction_id" toml:"transaction_id" yaml:"transaction_id"`
	TokenAddress  string         `boil:"token_address" json:"token_address" toml:"token_address" yaml:"token_address"`
	RuleID        sql.NullInt64  `boil:"rule_id" json:"rule_id" toml:"rule_id" yaml:"rule_id"`
	Type          string         `boil:"type" json:"type" toml:"type" yaml:"type"`
	Amount        float64        `boil:"amount" json:"amount" toml:"amount" yaml:"amount"`
	Price         float64        `boil:"price" json:"price" toml:"price" yaml:"price"`
	Timestamp     int64          `boil:"timestamp" json:"timestamp" toml:"timestamp" yaml:"timestamp"`
	TxHash        sql.NullString `boil:"tx_hash" json:"tx_hash" toml:"tx_hash" yaml:"tx_hash"`
	Status        string         `boil:"status" json:"status" toml:"status" yaml:"status"`
}

// Insert 使用执行器插入单条记录
func (o *TransactionRecord) Insert(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no transaction_record provided for insertion")
	}

	// 定义要插入的列，排除自增的 transaction_id
	transactionRecordColumns := []string{
		"token_address",
		"rule_id",
		"type",
		"amount",
		"price",
		"timestamp",
		"tx_hash",
		"status",
	}

	// 构建 SQL 查询
	query := fmt.Sprintf(
		"INSERT INTO \"transaction_records\" (\"%s\") VALUES (%s)",
		strings.Join(transactionRecordColumns, "\",\""),
		strmangle.Placeholders(dialect.UseIndexPlaceholders, len(transactionRecordColumns), 1, 1),
	)

	// 准备插入的值
	vals := []interface{}{
		o.TokenAddress,
		o.RuleID,
		o.Type,
		o.Amount,
		o.Price,
		o.Timestamp,
		o.TxHash,
		o.Status,
	}

	// 执行插入操作
	result, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to insert into transaction_records")
	}

	// 获取自增 ID 并赋值给结构体
	id, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to get last insert id for transaction_records")
	}
	o.TransactionID = int(id)

	return nil
}

// Update 更新数据库中的 TransactionRecord 记录
func (o *TransactionRecord) Update(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no transaction_record provided for update")
	}

	// 定义要更新的字段（不包括 transaction_id）
	columns := []string{
		"token_address",
		"rule_id",
		"type",
		"amount",
		"price",
		"timestamp",
		"tx_hash",
		"status",
	}

	// 构建 SQL 更新查询
	query := fmt.Sprintf(
		"UPDATE \"transaction_records\" SET \"%s\" = %s WHERE \"transaction_id\" = ?",
		strings.Join(columns, "\" = ?, \""),
		strmangle.Placeholders(dialect.UseIndexPlaceholders, len(columns), 1, 1),
	)

	// 准备更新的值
	vals := []interface{}{
		o.TokenAddress,
		o.RuleID,
		o.Type,
		o.Amount,
		o.Price,
		o.Timestamp,
		o.TxHash,
		o.Status,
		o.TransactionID, // 用于 WHERE 条件
	}

	// 执行更新操作
	_, err := exec.ExecContext(ctx, query, vals...)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to update transaction_records")
	}

	return nil
}

// Delete 从数据库中删除 TransactionRecord 记录
func (o *TransactionRecord) Delete(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil {
		return errors.New("sqlite3: no transaction_record provided for deletion")
	}

	// 检查主键 transaction_id 是否已设置
	if o.TransactionID == 0 {
		return errors.New("sqlite3: transaction_record has no primary key value for deletion")
	}

	// 构建 SQL 删除查询
	query := fmt.Sprintf(
		"DELETE FROM \"transaction_records\" WHERE \"transaction_id\" = ?",
	)

	// 执行删除操作
	_, err := exec.ExecContext(ctx, query, o.TransactionID)
	if err != nil {
		return errors.Wrap(err, "sqlite3: unable to delete from transaction_records")
	}

	return nil
}

// transactionRecordQuery 用于构建 TransactionRecord 记录的查询
type transactionRecordQuery struct {
	*queries.Query
}

// TransactionRecordSlice 是指向 TransactionRecord 的指针切片别名
type TransactionRecordSlice []*TransactionRecord

// TransactionRecords 使用执行器检索所有记录
func TransactionRecords(mods ...qm.QueryMod) transactionRecordQuery {
	mods = append(mods, qm.From("\"transaction_records\""))
	return transactionRecordQuery{NewQuery(mods...)}
}

// All 从查询中返回所有 TransactionRecord 记录
func (q transactionRecordQuery) All(ctx context.Context, exec boil.ContextExecutor) (TransactionRecordSlice, error) {
	var o TransactionRecordSlice

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "sqlite3: failed to assign all query results to TransactionRecord slice")
	}

	return o, nil
}
