package account

import (
	"context"
	"fmt"

	"gocryptotrader/database"
	modelSQLite "gocryptotrader/database/models/sqlite3"

	"github.com/thrasher-corp/sqlboiler/queries/qm"
)

// GetAccount returns list of accounts matching query
func GetAccount(name string, limit int) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	var mods []qm.QueryMod
	if name != "" {
		mods = append(mods, qm.Where("name = ?", name))
	}

	mods = append(mods, qm.OrderBy("id"))
	if limit > 0 {
		mods = append(mods, qm.Limit(limit))
	}

	ctx := context.TODO()
	return modelSQLite.Accounts(mods...).All(ctx, database.DB.SQL)

}

// GetAccountByID returns a specific account by ID
func GetAccountByID(id int) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindAccount(ctx, database.DB.SQL, id)

}

// GetAccountByAddress returns a specific account by address
func GetAccountByAddress(address string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	account, err := modelSQLite.FindAccountByAddress(ctx, database.DB.SQL, address)
	if err != nil {
		return nil, fmt.Errorf("获取地址信息失败: %w", err)
	}

	return account, nil
}
