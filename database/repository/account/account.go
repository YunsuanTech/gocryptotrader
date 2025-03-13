package account

import (
	"context"

	"gocryptotrader/database"
	modelSQLite "gocryptotrader/database/models/sqlite3"
	"gocryptotrader/database/repository"
	"gocryptotrader/log"

	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries/qm"
)

// Account inserts a new account to database
func Account(name, address, exchangeAddressID, zkAddressID, f4AddressID, otAddressID, cipher string, layer int, owner, chainName string) error {
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	ctx = boil.SkipTimestamps(ctx)

	tx, err := database.DB.SQL.BeginTx(ctx, nil)
	if err != nil {
		log.Errorf(log.Global, "Account transaction begin failed: %v", err)
		return err
	}

	if repository.GetSQLDialect() == database.DBSQLite3 {
		var tempAccount = modelSQLite.Account{
			Name:              name,
			Address:           address,
			ExchangeAddressID: exchangeAddressID,
			ZkAddressID:       zkAddressID,
			F4AddressID:       f4AddressID,
			OTAddressID:       otAddressID,
			Cipher:            cipher,
			Layer:             layer,
			Owner:             owner,
			ChainName:         chainName,
		}
		// 注意：SQLite版本的Account结构体使用的是gorm标签，可能需要调整Insert方法
		// 这里假设有Insert方法，实际使用时可能需要调整
		err = tempAccount.Insert(ctx, tx, boil.Infer())
	}

	if err != nil {
		log.Errorf(log.Global, "Account insert failed: %v", err)
		err = tx.Rollback()
		if err != nil {
			log.Errorf(log.Global, "Account Transaction rollback failed: %v", err)
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Errorf(log.Global, "Account Transaction commit failed: %v", err)
		return err
	}

	return nil
}

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

// GetAccountByName returns a specific account by name
func GetAccountByName(name string) (interface{}, error) {
	if database.DB.SQL == nil {
		return nil, database.ErrDatabaseSupportDisabled
	}

	ctx := context.TODO()
	return modelSQLite.FindAccountByName(ctx, database.DB.SQL, name)

}
