package account

import (
	"fmt"
	"reflect"

	accountsql "gocryptotrader/database/repository/account"
)

// Manager 管理账户相关操作
type Manager struct{}

// New 创建一个新的账户管理器
func New() *Manager {
	return &Manager{}
}

// Accounts 获取所有账户信息
func (m *Manager) Accounts() ([]*Account, error) {
	accounts, err := accountsql.GetAccount("", 0)
	if err != nil {
		return nil, fmt.Errorf("获取账户列表失败: %w", err)
	}

	// 检查返回类型并进行适当的转换
	switch v := accounts.(type) {
	case []*Account:
		// 如果已经是正确的类型，直接返回
		return v, nil
	case interface{}:
		// 尝试从反射获取切片中的每个元素并转换
		return convertToAccountSlice(v)
	default:
		return nil, fmt.Errorf("无法转换账户数据类型：未知类型 %T", accounts)
	}
}

// convertToAccountSlice 将接口类型转换为[]*Account
func convertToAccountSlice(data interface{}) ([]*Account, error) {
	var result []*Account

	sliceValue := reflect.ValueOf(data)
	if sliceValue.Kind() != reflect.Slice {
		return nil, fmt.Errorf("无法转换账户数据类型：不是切片类型")
	}

	// 遍历切片中的每个元素
	for i := 0; i < sliceValue.Len(); i++ {
		item := sliceValue.Index(i).Interface()

		// 尝试将每个元素转换为map并创建Account对象
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			// 如果不是map，尝试直接转换
			accObj, ok := item.(*Account)
			if !ok {
				// 如果无法直接转换，尝试使用反射获取字段值
				accObj = convertStructToAccount(item)
				if accObj == nil {
					continue
				}
			}
			result = append(result, accObj)
			continue
		}

		// 从map创建Account对象
		accObj := mapToAccount(itemMap)
		result = append(result, accObj)
	}

	return result, nil
}

// convertStructToAccount 使用反射将结构体转换为Account
func convertStructToAccount(item interface{}) *Account {
	itemValue := reflect.ValueOf(item)
	if itemValue.Kind() == reflect.Ptr {
		itemValue = itemValue.Elem()
	}

	if itemValue.Kind() != reflect.Struct {
		return nil
	}

	// 创建新的Account对象
	accObj := &Account{}

	// 尝试获取并设置字段值
	if f := itemValue.FieldByName("ID"); f.IsValid() {
		accObj.ID = int(f.Int())
	}
	if f := itemValue.FieldByName("Name"); f.IsValid() {
		accObj.Name = f.String()
	}
	if f := itemValue.FieldByName("Address"); f.IsValid() {
		accObj.Address = f.String()
	}
	if f := itemValue.FieldByName("ExchangeAddressID"); f.IsValid() {
		accObj.ExchangeAddressID = f.String()
	}
	if f := itemValue.FieldByName("ZkAddressID"); f.IsValid() {
		accObj.ZkAddressID = f.String()
	}
	if f := itemValue.FieldByName("F4AddressID"); f.IsValid() {
		accObj.F4AddressID = f.String()
	}
	if f := itemValue.FieldByName("OTAddressID"); f.IsValid() {
		accObj.OTAddressID = f.String()
	}
	if f := itemValue.FieldByName("Cipher"); f.IsValid() {
		accObj.Cipher = f.String()
	}
	if f := itemValue.FieldByName("Layer"); f.IsValid() {
		accObj.Layer = int(f.Int())
	}
	if f := itemValue.FieldByName("Owner"); f.IsValid() {
		accObj.Owner = f.String()
	}
	if f := itemValue.FieldByName("ChainName"); f.IsValid() {
		accObj.ChainName = f.String()
	}

	return accObj
}

// mapToAccount 将map转换为Account
func mapToAccount(itemMap map[string]interface{}) *Account {
	accObj := &Account{}

	if v, ok := itemMap["id"].(int); ok {
		accObj.ID = v
	}
	if v, ok := itemMap["name"].(string); ok {
		accObj.Name = v
	}
	if v, ok := itemMap["address"].(string); ok {
		accObj.Address = v
	}
	if v, ok := itemMap["exchange_address_id"].(string); ok {
		accObj.ExchangeAddressID = v
	}
	if v, ok := itemMap["zk_address_id"].(string); ok {
		accObj.ZkAddressID = v
	}
	if v, ok := itemMap["f4_address_id"].(string); ok {
		accObj.F4AddressID = v
	}
	if v, ok := itemMap["ot_address_id"].(string); ok {
		accObj.OTAddressID = v
	}
	if v, ok := itemMap["cipher"].(string); ok {
		accObj.Cipher = v
	}
	if v, ok := itemMap["layer"].(int); ok {
		accObj.Layer = v
	}
	if v, ok := itemMap["owner"].(string); ok {
		accObj.Owner = v
	}
	if v, ok := itemMap["chain_name"].(string); ok {
		accObj.ChainName = v
	}

	return accObj
}

// GetAccountByID 根据ID获取账户信息
func (m *Manager) GetAccountByID(id int) (*Account, error) {
	account, err := accountsql.GetAccountByID(id)
	if err != nil {
		return nil, fmt.Errorf("获取账户信息失败: %w", err)
	}

	// 将interface{}转换为*Account
	accountObj, ok := account.(*Account)
	if !ok {
		return nil, fmt.Errorf("无法转换账户数据类型: %T", account)
	}

	return accountObj, nil
}

// GetAccountByName 根据名称获取账户信息
func (m *Manager) GetAccountByName(name string) (*Account, error) {
	account, err := accountsql.GetAccountByName(name)
	if err != nil {
		return nil, fmt.Errorf("获取账户信息失败: %w", err)
	}

	// 将interface{}转换为*Account
	accountObj, ok := account.(*Account)
	if !ok {
		return nil, fmt.Errorf("无法转换账户数据类型: %T", account)
	}

	return accountObj, nil
}

// CreateAccount 创建新账户
// 参数验证并创建新的账户记录
func (m *Manager) CreateAccount(name, address, exchangeAddressID, zkAddressID, f4AddressID, otAddressID, cipher string, layer int, owner, chainName string) error {
	// 参数验证
	if err := validateAccountParams(name, address, layer); err != nil {
		return err
	}

	// 检查账户名是否已存在
	existingAccount, err := m.GetAccountByName(name)
	if err == nil && existingAccount != nil {
		return fmt.Errorf("账户名 '%s' 已存在", name)
	}

	// 调用数据库层创建账户
	err = accountsql.Account(name, address, exchangeAddressID, zkAddressID, f4AddressID, otAddressID, cipher, layer, owner, chainName)
	if err != nil {
		return fmt.Errorf("创建账户失败: %w", err)
	}

	return nil
}

// validateAccountParams 验证账户参数
func validateAccountParams(name, address string, layer int) error {
	if name == "" {
		return fmt.Errorf("账户名称不能为空")
	}

	if address == "" {
		return fmt.Errorf("地址不能为空")
	}

	if layer < 0 {
		return fmt.Errorf("层级不能为负数")
	}

	return nil
}
