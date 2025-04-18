package token

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"gocryptotrader/config"
	watchlistsql "gocryptotrader/database/repository/watchlist_tokens"
)

// WatchlistManager 管理代币监视列表相关操作
type WatchlistManager struct {
	config *config.Config
}

// NewWatchlistManager 创建一个新的监视列表管理器
func NewWatchlistManager(cfg *config.Config) *WatchlistManager {
	return &WatchlistManager{config: cfg}
}

// WatchlistToken 代表监视列表中的代币信息
type WatchlistToken struct {
	ID           int
	TokenSymbol  string
	TokenAddress string
	Network      string
	Decimals     int
	CreationTime int64
	LastUpdated  int64
	IsActive     int
}

// processTokensResult 处理代币查询结果并转换为适当的类型
func processTokensResult(tokens interface{}, operation string) ([]*WatchlistToken, error) {
	switch v := tokens.(type) {
	case []*WatchlistToken:
		// 如果已经是正确的类型，直接返回
		return v, nil
	case interface{}:
		// 尝试从反射获取切片中的每个元素并转换
		return convertToWatchlistTokenSlice(v)
	default:
		return nil, fmt.Errorf("无法转换代币数据类型：未知类型 %T", tokens)
	}
}

// processTokenResult 处理单个代币查询结果并转换为适当的类型
func processTokenResult(token interface{}, operation string) (*WatchlistToken, error) {
	// 尝试将interface{}转换为*WatchlistToken
	tokenObj, ok := token.(*WatchlistToken)
	if ok {
		return tokenObj, nil
	}

	// 使用反射进行转换
	result := convertStructToWatchlistToken(token)
	if result == nil {
		return nil, fmt.Errorf("无法转换代币数据类型: %T", token)
	}

	return result, nil
}

// GetAllWatchlistTokens 获取所有监视列表代币
func (m *WatchlistManager) GetAllWatchlistTokens(network string, limit int) ([]*WatchlistToken, error) {
	return m.QueryWatchlistTokens(
		func() (interface{}, error) {
			return watchlistsql.GetWatchlistTokens(network, limit)
		},
		"获取代币列表",
	)
}

// QueryWatchlistTokens 通用查询函数，用于获取代币列表
func (m *WatchlistManager) QueryWatchlistTokens(queryFunc func() (interface{}, error), operation string) ([]*WatchlistToken, error) {
	tokens, err := queryFunc()
	if err != nil {
		return nil, fmt.Errorf("%s失败: %w", operation, err)
	}

	return processTokensResult(tokens, operation)
}

// QueryWatchlistToken 通用查询函数，用于获取单个代币
func (m *WatchlistManager) QueryWatchlistToken(queryFunc func() (interface{}, error), operation string) (*WatchlistToken, error) {
	token, err := queryFunc()
	if err != nil {
		return nil, fmt.Errorf("%s失败: %w", operation, err)
	}

	return processTokenResult(token, operation)
}

// convertToWatchlistTokenSlice 将接口类型转换为[]*WatchlistToken
func convertToWatchlistTokenSlice(data interface{}) ([]*WatchlistToken, error) {
	var result []*WatchlistToken

	sliceValue := reflect.ValueOf(data)
	if sliceValue.Kind() != reflect.Slice {
		return nil, fmt.Errorf("无法转换代币数据类型：不是切片类型")
	}

	// 遍历切片中的每个元素
	for i := 0; i < sliceValue.Len(); i++ {
		item := sliceValue.Index(i).Interface()

		// 尝试将每个元素转换为map并创建WatchlistToken对象
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			// 如果不是map，尝试直接转换
			tokenObj, ok := item.(*WatchlistToken)
			if !ok {
				// 如果无法直接转换，尝试使用反射获取字段值
				tokenObj = convertStructToWatchlistToken(item)
				if tokenObj == nil {
					continue
				}
			}
			result = append(result, tokenObj)
			continue
		}

		// 从map创建WatchlistToken对象
		tokenObj := mapToWatchlistToken(itemMap)
		result = append(result, tokenObj)
	}

	return result, nil
}

// convertStructToWatchlistToken 使用反射将结构体转换为WatchlistToken
func convertStructToWatchlistToken(item interface{}) *WatchlistToken {
	itemValue := reflect.ValueOf(item)
	if itemValue.Kind() == reflect.Ptr {
		itemValue = itemValue.Elem()
	}

	if itemValue.Kind() != reflect.Struct {
		return nil
	}

	// 创建新的WatchlistToken对象
	tokenObj := &WatchlistToken{}

	// 尝试获取并设置字段值
	if f := itemValue.FieldByName("ID"); f.IsValid() {
		tokenObj.ID = int(f.Int())
	}
	if f := itemValue.FieldByName("TokenSymbol"); f.IsValid() {
		if nullStr, ok := f.Interface().(sql.NullString); ok {
			if nullStr.Valid {
				tokenObj.TokenSymbol = nullStr.String
			}
		} else {
			tokenObj.TokenSymbol = f.String()
		}
	}
	if f := itemValue.FieldByName("TokenAddress"); f.IsValid() {
		if nullStr, ok := f.Interface().(sql.NullString); ok {
			if nullStr.Valid {
				tokenObj.TokenAddress = nullStr.String
			}
		} else {
			tokenObj.TokenAddress = f.String()
		}
	}
	if f := itemValue.FieldByName("Network"); f.IsValid() {
		if nullStr, ok := f.Interface().(sql.NullString); ok {
			if nullStr.Valid {
				tokenObj.Network = nullStr.String
			}
		} else {
			tokenObj.Network = f.String()
		}
	}
	if f := itemValue.FieldByName("Decimals"); f.IsValid() {
		tokenObj.Decimals = int(f.Int())
	}
	if f := itemValue.FieldByName("CreationTime"); f.IsValid() {
		tokenObj.CreationTime = f.Int()
	}
	if f := itemValue.FieldByName("LastUpdated"); f.IsValid() {
		tokenObj.LastUpdated = f.Int()
	}
	if f := itemValue.FieldByName("IsActive"); f.IsValid() {
		tokenObj.IsActive = int(f.Int())
	}

	return tokenObj
}

// mapToWatchlistToken 将map转换为WatchlistToken
func mapToWatchlistToken(itemMap map[string]interface{}) *WatchlistToken {
	tokenObj := &WatchlistToken{}

	if v, ok := itemMap["id"].(int); ok {
		tokenObj.ID = v
	}
	if v, ok := itemMap["token_symbol"].(string); ok {
		tokenObj.TokenSymbol = v
	}
	if v, ok := itemMap["token_address"].(string); ok {
		tokenObj.TokenAddress = v
	}
	if v, ok := itemMap["network"].(string); ok {
		tokenObj.Network = v
	}
	if v, ok := itemMap["decimals"].(int); ok {
		tokenObj.Decimals = v
	}
	if v, ok := itemMap["creation_time"].(int64); ok {
		tokenObj.CreationTime = v
	}
	if v, ok := itemMap["last_updated"].(int64); ok {
		tokenObj.LastUpdated = v
	}
	if v, ok := itemMap["is_active"].(int); ok {
		tokenObj.IsActive = v
	}

	return tokenObj
}

// GetWatchlistTokenByID 根据ID获取代币信息
func (m *WatchlistManager) GetWatchlistTokenByID(tokenID int) (*WatchlistToken, error) {
	return m.QueryWatchlistToken(
		func() (interface{}, error) {
			return watchlistsql.GetWatchlistTokenByID(tokenID)
		},
		"获取代币信息",
	)
}

// GetWatchlistTokenByAddress 根据代币地址获取代币信息
func (m *WatchlistManager) GetWatchlistTokenByAddress(tokenAddress string) (*WatchlistToken, error) {
	return m.QueryWatchlistToken(
		func() (interface{}, error) {
			return watchlistsql.GetWatchlistTokenByAddress(tokenAddress)
		},
		"获取代币信息",
	)
}

// GetWatchlistTokensBySymbol 根据代币符号获取代币信息
func (m *WatchlistManager) GetWatchlistTokensBySymbol(tokenSymbol string) ([]*WatchlistToken, error) {
	return m.QueryWatchlistTokens(
		func() (interface{}, error) {
			return watchlistsql.GetWatchlistTokensBySymbol(tokenSymbol)
		},
		"获取代币信息",
	)
}

// GetWatchlistTokensByNetwork 根据网络获取代币信息
func (m *WatchlistManager) GetWatchlistTokensByNetwork(network string, limit int) ([]*WatchlistToken, error) {
	return m.QueryWatchlistTokens(
		func() (interface{}, error) {
			return watchlistsql.GetWatchlistTokensByNetwork(network, limit)
		},
		"获取代币信息",
	)
}

// GetActiveWatchlistTokens 获取所有活跃的代币信息
func (m *WatchlistManager) GetActiveWatchlistTokens() ([]*WatchlistToken, error) {
	return m.QueryWatchlistTokens(
		func() (interface{}, error) {
			return watchlistsql.GetActiveWatchlistTokens()
		},
		"获取活跃代币信息",
	)
}

// AddWatchlistToken 添加新的代币到监视列表
func (m *WatchlistManager) AddWatchlistToken(tokenSymbol, tokenAddress, network string, decimals int, creationTime, lastUpdated int64, isActive int) error {
	if err := watchlistsql.AddWatchlistToken(tokenSymbol, tokenAddress, network, decimals, creationTime, lastUpdated, isActive); err != nil {
		return fmt.Errorf("添加代币到监视列表失败: %w", err)
	}
	return nil
}

// UpdateWatchlistToken 更新监视列表中的代币信息
func (m *WatchlistManager) UpdateWatchlistToken(tokenID int, tokenSymbol, tokenAddress, network string, decimals int, isActive int) error {
	// 设置最后更新时间为当前时间
	lastUpdated := time.Now().Unix()

	// 调用repository层的更新方法
	if err := watchlistsql.UpdateWatchlistToken(tokenID, tokenSymbol, tokenAddress, network, decimals, lastUpdated, isActive); err != nil {
		return fmt.Errorf("更新代币信息失败: %w", err)
	}
	return nil
}

// UpdateWatchlistTokenByAddress 根据代币地址更新监视列表中的代币信息
func (m *WatchlistManager) UpdateWatchlistTokenByAddress(tokenAddress string, tokenSymbol, newTokenAddress, network string, decimals int, isActive int) error {
	// 设置最后更新时间为当前时间
	lastUpdated := time.Now().Unix()

	// 调用repository层的更新方法
	if err := watchlistsql.UpdateWatchlistTokenByAddress(tokenAddress, tokenSymbol, newTokenAddress, network, decimals, lastUpdated, isActive); err != nil {
		return fmt.Errorf("更新代币信息失败: %w", err)
	}
	return nil
}

// DeleteWatchlistToken 根据ID从监视列表中删除代币
func (m *WatchlistManager) DeleteWatchlistToken(tokenID int) error {
	// 调用repository层的删除方法
	if err := watchlistsql.DeleteWatchlistToken(tokenID); err != nil {
		return fmt.Errorf("删除代币失败: %w", err)
	}
	return nil
}

// DeleteWatchlistTokenByAddress 根据代币地址从监视列表中删除代币
func (m *WatchlistManager) DeleteWatchlistTokenByAddress(tokenAddress string) error {
	// 调用repository层的删除方法
	if err := watchlistsql.DeleteWatchlistTokenByAddress(tokenAddress); err != nil {
		return fmt.Errorf("删除代币失败: %w", err)
	}
	return nil
}
