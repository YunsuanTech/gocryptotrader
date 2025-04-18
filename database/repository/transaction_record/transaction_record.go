package transactionrecord

import (
	"context"
	"errors"

	"gocryptotrader/database"
	modelSQLite "gocryptotrader/database/models/sqlite3"

	"github.com/thrasher-corp/sqlboiler/queries/qm"
)

// TransactionRecordQueryOptions 定义查询 TransactionRecord 的选项
type TransactionRecordQueryOptions struct {
	TransactionID int    // 根据 transaction_id 查询
	TokenAddress  string // 根据 token_address 查询
	RuleID        int    // 根据 rule_id 查询（可为 NULL）
	Type          string // 根据 type 查询
	Status        string // 根据 status 查询
	Limit         int    // 返回记录的最大数量
}

// QueryTransactionRecords 根据提供的选项查询 TransactionRecord 记录
func QueryTransactionRecords(opts TransactionRecordQueryOptions) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}
	var mods []qm.QueryMod

	// 添加 TransactionID 条件
	if opts.TransactionID != 0 {
		mods = append(mods, qm.Where("transaction_id = ?", opts.TransactionID))
	}
	// 添加 TokenAddress 条件
	if opts.TokenAddress != "" {
		mods = append(mods, qm.Where("token_address = ?", opts.TokenAddress))
	}
	// 添加 RuleID 条件（可为 NULL）
	if opts.RuleID != 0 {
		mods = append(mods, qm.Where("rule_id = ?", opts.RuleID))
	}
	// 添加 Type 条件
	if opts.Type != "" {
		mods = append(mods, qm.Where("type = ?", opts.Type))
	}
	// 添加 Status 条件
	if opts.Status != "" {
		mods = append(mods, qm.Where("status = ?", opts.Status))
	}

	// 默认按 transaction_id 排序
	mods = append(mods, qm.OrderBy("transaction_id"))

	// 设置查询限制
	if opts.Limit > 0 {
		mods = append(mods, qm.Limit(opts.Limit))
	}

	ctx := context.TODO()
	return modelSQLite.TransactionRecords(mods...).All(ctx, database.DB.SQL)
}

// InsertTransactionRecord 插入一条新的 TransactionRecord 记录
func InsertTransactionRecord(ctx context.Context, tr *modelSQLite.TransactionRecord) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 TransactionRecord 是否为 nil
	if tr == nil {
		return errors.New("transaction record cannot be nil")
	}

	// 调用 TransactionRecord 的 Insert 方法执行插入
	err := tr.Insert(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// UpdateTransactionRecord 更新一条 TransactionRecord 记录
func UpdateTransactionRecord(ctx context.Context, tr *modelSQLite.TransactionRecord) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 TransactionRecord 是否为 nil
	if tr == nil {
		return errors.New("transaction record cannot be nil")
	}

	// 调用 TransactionRecord 的 Update 方法执行更新
	err := tr.Update(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// DeleteTransactionRecord 删除一条 TransactionRecord 记录
func DeleteTransactionRecord(ctx context.Context, tr *modelSQLite.TransactionRecord) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 TransactionRecord 是否为 nil
	if tr == nil {
		return errors.New("transaction record cannot be nil")
	}

	// 调用 TransactionRecord 的 Delete 方法执行删除
	err := tr.Delete(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}
