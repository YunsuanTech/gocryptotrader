package xens

import (
	"context"
	"database/sql"

	"gocryptotrader/database"
	modelSQLite "gocryptotrader/database/models/sqlite3"

	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries/qm"
)

// GetXens 获取所有Xen记录
func GetXens(limit int) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	var mods []qm.QueryMod
	mods = append(mods, qm.OrderBy("slot ASC"))
	if limit > 0 {
		mods = append(mods, qm.Limit(limit))
	}

	ctx := context.TODO()
	return modelSQLite.Xens(mods...).All(ctx, database.DB.SQL)
}

// GetXenBySlotAndChain 根据槽位和链名获取Xen记录
func GetXenBySlotAndChain(slot int, chainName string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindXen(ctx, database.DB.SQL, slot, chainName)
}

// GetXensByChainName 根据链名获取Xen记录
func GetXensByChainName(chainName string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	var mods []qm.QueryMod
	mods = append(mods, qm.Where("chain_name=?", chainName))
	mods = append(mods, qm.OrderBy("slot ASC"))

	ctx := context.TODO()
	return modelSQLite.Xens(mods...).All(ctx, database.DB.SQL)
}

// GetXensByStatus 根据状态获取Xen记录
func GetXensByStatus(status string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	var mods []qm.QueryMod
	mods = append(mods, qm.Where("status=?", status))
	mods = append(mods, qm.OrderBy("slot ASC"))

	ctx := context.TODO()
	return modelSQLite.Xens(mods...).All(ctx, database.DB.SQL)
}

// GetXensByStatus 根据状态获取Xen记录
func GetXensByStatusByChainName(status string, chainName string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	var mods []qm.QueryMod
	mods = append(mods, qm.Where("chain_name=?", chainName))
	mods = append(mods, qm.Where("status=?", status))
	mods = append(mods, qm.OrderBy("slot ASC"))

	ctx := context.TODO()
	return modelSQLite.Xens(mods...).All(ctx, database.DB.SQL)
}

// AddXen 添加新的Xen记录
func AddXen(
	slot int,
	chainName string,
	count int,
	days int,
	executionTime sql.NullTime,
	claimTime sql.NullTime,
	expectedReward sql.NullFloat64,
	ranking int,
	amp int,
	eaa float64,
	m sql.NullFloat64,
	status string,
	txID sql.NullString,
	mintFees sql.NullFloat64,
	claimFees sql.NullFloat64) error {
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 创建一个新的Xen对象
	xen := &modelSQLite.Xen{
		Slot:           slot,
		ChainName:      chainName,
		Count:          count,
		Days:           days,
		ExecutionTime:  executionTime,
		ClaimTime:      claimTime,
		ExpectedReward: expectedReward,
		Ranking:        ranking,
		Amp:            amp,
		Eaa:            eaa,
		M:              m,
		Status:         status,
		TxID:           txID,
		MintFees:       mintFees,
		ClaimFees:      claimFees,
	}

	// 插入记录到数据库
	ctx := context.TODO()
	return xen.Insert(ctx, database.DB.SQL, boil.Infer())
}

// UpdateXen 更新Xen记录
func UpdateXen(
	slot int,
	chainName string,
	count int,
	days int,
	executionTime sql.NullTime,
	claimTime sql.NullTime,
	expectedReward sql.NullFloat64,
	ranking int,
	amp int,
	eaa float64,
	m sql.NullFloat64,
	status string,
	txID sql.NullString,
	mintFees sql.NullFloat64,
	claimFees sql.NullFloat64) error {
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 获取现有记录
	ctx := context.TODO()
	existingXen, err := modelSQLite.FindXen(ctx, database.DB.SQL, slot, chainName)
	if err != nil {
		return err
	}

	// 更新记录字段
	existingXen.Count = count
	existingXen.Days = days
	existingXen.ExecutionTime = executionTime
	existingXen.ClaimTime = claimTime
	existingXen.ExpectedReward = expectedReward
	existingXen.Ranking = ranking
	existingXen.Amp = amp
	existingXen.Eaa = eaa
	existingXen.M = m
	existingXen.Status = status
	existingXen.TxID = txID
	existingXen.MintFees = mintFees
	existingXen.ClaimFees = claimFees

	// 更新数据库中的记录
	return existingXen.Update(ctx, database.DB.SQL, boil.Infer())
}

// UpdateXenStatus 更新Xen记录的状态
func UpdateXenStatus(slot int, chainName string, status string) error {
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 获取现有记录
	ctx := context.TODO()
	existingXen, err := modelSQLite.FindXen(ctx, database.DB.SQL, slot, chainName)
	if err != nil {
		return err
	}

	// 更新状态
	existingXen.Status = status

	// 更新数据库中的记录
	return existingXen.Update(ctx, database.DB.SQL, boil.Infer())
}
