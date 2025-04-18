package transaction_records

import (
	"context"
	"database/sql"

	"gocryptotrader/database"
	modelSQLite "gocryptotrader/database/models/sqlite3"

	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries/qm"
)

// GetTransactionRecords 获取所有交易记录
func GetTransactionRecords(limit int) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	var mods []qm.QueryMod
	mods = append(mods, qm.OrderBy("transaction_time DESC"))
	if limit > 0 {
		mods = append(mods, qm.Limit(limit))
	}

	ctx := context.TODO()
	return modelSQLite.TransactionRecords(mods...).All(ctx, database.DB.SQL)
}

// GetTransactionRecordByHash 根据交易哈希获取交易记录
func GetTransactionRecordByHash(transactionHash string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindTransactionRecord(ctx, database.DB.SQL, transactionHash)
}

// GetTransactionRecordsByInitiatingAddress 根据发起地址获取交易记录
func GetTransactionRecordsByInitiatingAddress(initiatingAddress string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindTransactionRecordsByInitiatingAddress(ctx, database.DB.SQL, initiatingAddress)
}

// GetTransactionRecordsByReceivingAddress 根据接收地址获取交易记录
func GetTransactionRecordsByReceivingAddress(receivingAddress string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindTransactionRecordsByReceivingAddress(ctx, database.DB.SQL, receivingAddress)
}

// GetTransactionRecordsByNetwork 根据网络标识获取交易记录
func GetTransactionRecordsByNetwork(networkIdentifier string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindTransactionRecordsByNetwork(ctx, database.DB.SQL, networkIdentifier)
}

// GetTransactionRecordsByTimeRange 根据时间范围获取交易记录
func GetTransactionRecordsByTimeRange(startTime, endTime string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindTransactionRecordsByTimeRange(ctx, database.DB.SQL, startTime, endTime)
}

// GetTransactionRecordsByToken 根据代币获取交易记录
func GetTransactionRecordsByToken(token string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindTransactionRecordsByToken(ctx, database.DB.SQL, token)
}

// AddTransactionRecord 添加新的交易记录
func AddTransactionRecord(
	transactionHash, inputToken string,
	inputAmount float64,
	outputToken string,
	outputAmount float64,
	receivingAddress string,
	slippage float64,
	networkIdentifier, transactionTime string,
	price float64,
	initiatingAddress string) error {
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 创建一个新的TransactionRecord对象
	record := &modelSQLite.TransactionRecord{
		TransactionHash:   transactionHash,
		InputToken:        inputToken,
		InputAmount:       inputAmount,
		OutputToken:       outputToken,
		OutputAmount:      outputAmount,
		ReceivingAddress:  receivingAddress,
		Slippage:          sql.NullFloat64{Float64: slippage, Valid: true},
		NetworkIdentifier: networkIdentifier,
		TransactionTime:   transactionTime,
		Price:             price,
		InitiatingAddress: initiatingAddress,
	}

	// 插入记录到数据库
	ctx := context.TODO()
	return record.Insert(ctx, database.DB.SQL, boil.Infer())
}

// DeleteTransactionRecord 删除交易记录
func DeleteTransactionRecord(transactionHash string) error {
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 获取现有记录
	ctx := context.TODO()
	existingRecord, err := modelSQLite.FindTransactionRecord(ctx, database.DB.SQL, transactionHash)
	if err != nil {
		return err
	}

	// 从数据库删除记录
	_, err = existingRecord.Delete(ctx, database.DB.SQL)
	return err
}
