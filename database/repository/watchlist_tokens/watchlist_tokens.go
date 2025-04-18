package watchlist_tokens

import (
	"context"
	"fmt"
	"time"

	"gocryptotrader/database"
	modelSQLite "gocryptotrader/database/models/sqlite3"

	"github.com/thrasher-corp/sqlboiler/boil"
	"github.com/thrasher-corp/sqlboiler/queries/qm"
)

// checkDatabaseSupport 检查数据库是否支持SQL操作
func checkDatabaseSupport() error {
	if database.DB.SQL == nil {
		return database.ErrDatabaseSupportDisabled
	}
	return nil
}

// GetWatchlistTokens 返回符合查询条件的代币列表
func GetWatchlistTokens(network string, limit int) ([]*modelSQLite.WatchlistToken, error) {
	if err := checkDatabaseSupport(); err != nil {
		return nil, err
	}

	var mods []qm.QueryMod
	if network != "" {
		mods = append(mods, qm.Where("network = ?", network))
	}
	mods = append(mods, qm.OrderBy("creation_time DESC"))
	if limit > 0 {
		mods = append(mods, qm.Limit(limit))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := modelSQLite.WatchlistTokens(mods...).All(ctx, database.DB.SQL)
	if err != nil {
		return nil, fmt.Errorf("获取代币列表失败: %w", err)
	}

	return result, nil
}

// GetWatchlistTokenByID 通过ID获取特定的代币记录
func GetWatchlistTokenByID(tokenID int) (*modelSQLite.WatchlistToken, error) {
	if err := checkDatabaseSupport(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return modelSQLite.FindWatchlistToken(ctx, database.DB.SQL, tokenID)
}

// GetWatchlistTokenByAddress 通过代币地址获取特定的代币记录
func GetWatchlistTokenByAddress(tokenAddress string) (*modelSQLite.WatchlistToken, error) {
	if err := checkDatabaseSupport(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return modelSQLite.FindWatchlistTokenByAddress(ctx, database.DB.SQL, tokenAddress)
}

// GetWatchlistTokensBySymbol 通过代币符号获取代币记录
func GetWatchlistTokensBySymbol(tokenSymbol string) ([]*modelSQLite.WatchlistToken, error) {
	if err := checkDatabaseSupport(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return modelSQLite.FindWatchlistTokensBySymbol(ctx, database.DB.SQL, tokenSymbol)
}

// GetWatchlistTokensByNetwork 获取特定网络的所有代币记录
func GetWatchlistTokensByNetwork(network string, limit int) ([]*modelSQLite.WatchlistToken, error) {
	if err := checkDatabaseSupport(); err != nil {
		return nil, err
	}

	var mods []qm.QueryMod
	mods = append(mods, qm.Where("network = ?", network))
	mods = append(mods, qm.OrderBy("creation_time DESC"))
	if limit > 0 {
		mods = append(mods, qm.Limit(limit))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tokens, err := modelSQLite.WatchlistTokens(mods...).All(ctx, database.DB.SQL)
	if err != nil {
		return nil, fmt.Errorf("获取代币列表失败: %w", err)
	}
	return tokens, nil
}

// GetActiveWatchlistTokens 获取所有活跃的代币记录
func GetActiveWatchlistTokens() ([]*modelSQLite.WatchlistToken, error) {
	if err := checkDatabaseSupport(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return modelSQLite.FindActiveWatchlistTokens(ctx, database.DB.SQL)
}

// AddWatchlistToken 添加新的代币到监视列表
func AddWatchlistToken(tokenSymbol, tokenAddress, network string, decimals int, creationTime, lastUpdated int64, isActive int) error {
	if err := checkDatabaseSupport(); err != nil {
		return err
	}

	token := &modelSQLite.WatchlistToken{
		TokenSymbol:  tokenSymbol,
		TokenAddress: tokenAddress,
		Network:      network,
		Decimals:     decimals,
		CreationTime: creationTime,
		LastUpdated:  lastUpdated,
		IsActive:     isActive,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return token.Insert(ctx, database.DB.SQL, boil.Infer())
}

// UpdateWatchlistToken 更新现有的代币信息
func UpdateWatchlistToken(tokenID int, tokenSymbol, tokenAddress, network string, decimals int, lastUpdated int64, isActive int) error {
	if err := checkDatabaseSupport(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	token, err := modelSQLite.FindWatchlistToken(ctx, database.DB.SQL, tokenID)
	if err != nil {
		return fmt.Errorf("获取代币记录失败: %w", err)
	}

	token.TokenSymbol = tokenSymbol
	token.TokenAddress = tokenAddress
	token.Network = network
	token.Decimals = decimals
	token.LastUpdated = lastUpdated
	token.IsActive = isActive

	return token.Update(ctx, database.DB.SQL, boil.Infer())
}

// UpdateWatchlistTokenByAddress 根据代币地址更新代币信息
func UpdateWatchlistTokenByAddress(tokenAddress string, tokenSymbol, newTokenAddress, network string, decimals int, lastUpdated int64, isActive int) error {
	if err := checkDatabaseSupport(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	token, err := modelSQLite.FindWatchlistTokenByAddress(ctx, database.DB.SQL, tokenAddress)
	if err != nil {
		return fmt.Errorf("获取代币记录失败: %w", err)
	}

	token.TokenSymbol = tokenSymbol
	token.TokenAddress = newTokenAddress
	token.Network = network
	token.Decimals = decimals
	token.LastUpdated = lastUpdated
	token.IsActive = isActive

	return token.Update(ctx, database.DB.SQL, boil.Infer())
}

// DeleteWatchlistToken 根据ID删除代币
func DeleteWatchlistToken(tokenID int) error {
	if err := checkDatabaseSupport(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	token, err := modelSQLite.FindWatchlistToken(ctx, database.DB.SQL, tokenID)
	if err != nil {
		return fmt.Errorf("获取代币记录失败: %w", err)
	}

	return token.Delete(ctx, database.DB.SQL)
}

// DeleteWatchlistTokenByAddress 根据代币地址删除代币
func DeleteWatchlistTokenByAddress(tokenAddress string) error {
	if err := checkDatabaseSupport(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return modelSQLite.DeleteWatchlistTokenByAddress(ctx, database.DB.SQL, tokenAddress)
}
