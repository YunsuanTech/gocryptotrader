package tokenmonitor

import (
	"context"
	"errors"

	"gocryptotrader/database"
	modelSQLite "gocryptotrader/database/models/sqlite3"

	"github.com/thrasher-corp/sqlboiler/queries/qm"
)

// TokenMonitorQueryOptions 定义查询 TokenMonitor 的选项
type TokenMonitorQueryOptions struct {
	TokenAddress string // 根据 token_address 查询
	TokenName    string // 根据 token_name 查询
	Limit        int    // 返回记录的最大数量
}

// QueryTokenMonitors 根据提供的选项查询 TokenMonitor 记录
func QueryTokenMonitors(opts TokenMonitorQueryOptions) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}
	var mods []qm.QueryMod

	// 添加 TokenAddress 条件
	if opts.TokenAddress != "" {
		mods = append(mods, qm.Where("token_address = ?", opts.TokenAddress))
	}
	// 添加 TokenName 条件
	if opts.TokenName != "" {
		mods = append(mods, qm.Where("token_name = ?", opts.TokenName))
	}

	// 默认按 token_address 排序
	mods = append(mods, qm.OrderBy("token_address"))

	// 设置查询限制
	if opts.Limit > 0 {
		mods = append(mods, qm.Limit(opts.Limit))
	}

	ctx := context.TODO()
	return modelSQLite.TokenMonitors(mods...).All(ctx, database.DB.SQL)
}

// InsertTokenMonitor 插入一条新的 TokenMonitor 记录
func InsertTokenMonitor(ctx context.Context, tm *modelSQLite.TokenMonitor) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 TokenMonitor 是否为 nil
	if tm == nil {
		return errors.New("token monitor cannot be nil")
	}

	// 调用 TokenMonitor 的 Insert 方法执行插入
	err := tm.Insert(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// UpdateTokenMonitor 更新一条 TokenMonitor 记录
func UpdateTokenMonitor(ctx context.Context, tm *modelSQLite.TokenMonitor) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 TokenMonitor 是否为 nil
	if tm == nil {
		return errors.New("token monitor cannot be nil")
	}

	// 调用 TokenMonitor 的 Update 方法执行更新
	err := tm.Update(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}

// DeleteTokenMonitor 删除一条 TokenMonitor 记录
func DeleteTokenMonitor(ctx context.Context, tm *modelSQLite.TokenMonitor) error {
	// 检查数据库连接是否可用
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	// 检查传入的 TokenMonitor 是否为 nil
	if tm == nil {
		return errors.New("token monitor cannot be nil")
	}

	// 调用 TokenMonitor 的 Delete 方法执行删除
	err := tm.Delete(ctx, database.DB.SQL)
	if err != nil {
		return err
	}

	return nil
}
