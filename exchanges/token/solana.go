package token

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// SolanaTokenAccount 表示Solana代币账户信息
type SolanaTokenAccount struct {
	Mint        string  `json:"mint"`        // 代币的Mint地址
	Owner       string  `json:"owner"`       // 代币所有者地址
	Account     string  `json:"account"`     // 代币账户地址
	Amount      uint64  `json:"amount"`      // 代币数量（原始值）
	Decimals    uint8   `json:"decimals"`    // 代币小数位数
	UIAmount    float64 `json:"uiAmount"`    // 用户界面显示的数量
	UIAmountStr string  `json:"uiAmountStr"` // 用户界面显示的数量（字符串格式）
}

// parsedData 解析 JSONParsed 用的本地结构体
type parsedData struct {
	Program string `json:"program"`
	Parsed  struct {
		Info struct {
			Mint        string `json:"mint"`
			Owner       string `json:"owner"`
			State       string `json:"state"`
			TokenAmount struct {
				Amount      string  `json:"amount"`
				Decimals    int     `json:"decimals"`
				UIAmount    float64 `json:"uiAmount"`
				UIAmountStr string  `json:"uiAmountString"`
			} `json:"tokenAmount"`
		} `json:"info"`
	} `json:"parsed"`
}

// GetTokenAccountsByOwner 获取指定钱包地址的所有代币账户信息
// ownerAddress: 钱包地址（Base58格式）
// rpcEndpoint: Solana RPC节点地址，如果为空则使用默认节点
// 返回代币账户列表和可能的错误
func GetTokenAccountsByOwner(ownerAddress string, rpcEndpoint string) ([]SolanaTokenAccount, error) {
	// 如果未提供RPC节点地址，则使用默认节点
	if rpcEndpoint == "" {
		rpcEndpoint = "http://xolana.xen.network:8899"
	}

	// 创建RPC客户端
	client := rpc.New(rpcEndpoint)

	// 解析钱包地址
	owner, err := solana.PublicKeyFromBase58(ownerAddress)
	if err != nil {
		return nil, fmt.Errorf("无效钱包地址: %v", err)
	}

	// 调用RPC获取代币账户
	out, err := client.GetTokenAccountsByOwner(
		context.Background(),
		owner,
		&rpc.GetTokenAccountsConfig{ProgramId: &solana.TokenProgramID},
		&rpc.GetTokenAccountsOpts{Encoding: solana.EncodingJSONParsed},
	)
	if err != nil {
		return nil, fmt.Errorf("RPC调用失败: %v", err)
	}

	// 检查是否有代币账户
	if len(out.Value) == 0 {
		return []SolanaTokenAccount{}, nil
	}

	// 解析结果
	tokenAccounts := make([]SolanaTokenAccount, 0, len(out.Value))
	for _, acc := range out.Value {
		// 提取原始JSON
		raw := acc.Account.Data.GetRawJSON()

		// 反序列化
		var pd parsedData
		if err := json.Unmarshal(raw, &pd); err != nil {
			log.Printf("解析token account JSON失败: %v", err)
			continue
		}

		// 解析数值
		amount, err := strconv.ParseUint(pd.Parsed.Info.TokenAmount.Amount, 10, 64)
		if err != nil {
			log.Printf("解析代币数量失败: %v", err)
			continue
		}

		// 创建代币账户对象
		tokenAccount := SolanaTokenAccount{
			Mint:        pd.Parsed.Info.Mint,
			Owner:       pd.Parsed.Info.Owner,
			Account:     acc.Pubkey.String(),
			Amount:      amount,
			Decimals:    uint8(pd.Parsed.Info.TokenAmount.Decimals),
			UIAmount:    pd.Parsed.Info.TokenAmount.UIAmount,
			UIAmountStr: pd.Parsed.Info.TokenAmount.UIAmountStr,
		}

		tokenAccounts = append(tokenAccounts, tokenAccount)
	}

	return tokenAccounts, nil
}

// GetTokenBalances 获取指定钱包地址的所有代币余额信息
// 这是一个便捷方法，返回格式化的代币余额信息
func GetTokenBalances(ownerAddress string, rpcEndpoint string) (map[string]float64, error) {
	// 获取代币账户
	tokenAccounts, err := GetTokenAccountsByOwner(ownerAddress, rpcEndpoint)
	if err != nil {
		return nil, err
	}

	// 创建余额映射
	balances := make(map[string]float64)
	for _, acc := range tokenAccounts {
		// 只添加余额大于0的代币
		if acc.UIAmount > 0 {
			balances[acc.Mint] = acc.UIAmount
		}
	}

	return balances, nil
}

// PrintTokenAccounts 打印指定钱包地址的所有代币账户信息
// 这是一个便捷方法，用于调试和显示
func PrintTokenAccounts(ownerAddress string, rpcEndpoint string) error {
	// 获取代币账户
	tokenAccounts, err := GetTokenAccountsByOwner(ownerAddress, rpcEndpoint)
	if err != nil {
		return err
	}

	// 打印账户信息
	if len(tokenAccounts) == 0 {
		fmt.Println("未找到任何代币账户。")
		return nil
	}

	fmt.Printf("找到 %d 个代币账户:\n", len(tokenAccounts))
	for i, acc := range tokenAccounts {
		fmt.Printf("[%d] Mint: %s\n", i+1, acc.Mint)
		fmt.Printf("    Account: %s\n", acc.Account)
		fmt.Printf("    Amount: %d (decimals=%d)\n", acc.Amount, acc.Decimals)
		fmt.Printf("    UI Amount: %s\n\n", acc.UIAmountStr)
	}

	return nil
}

// TokenBalanceWithPrice 表示代币余额和价格信息
type TokenBalanceWithPrice struct {
	// 代币信息
	Mint        string  `json:"mint"`          // 代币的Mint地址
	Amount      uint64  `json:"amount"`        // 代币数量（原始值）
	Decimals    uint8   `json:"decimals"`      // 代币小数位数
	UIAmount    float64 `json:"ui_amount"`     // 用户界面显示的数量
	UIAmountStr string  `json:"ui_amount_str"` // 用户界面显示的数量（字符串格式）

	// 价格信息
	USDPrice   float64   `json:"usd_price"`   // 代币的USD价格
	SOLPrice   float64   `json:"sol_price"`   // 代币的SOL价格
	USDValue   float64   `json:"usd_value"`   // 持有代币的USD总价值
	SOLValue   float64   `json:"sol_value"`   // 持有代币的SOL总价值
	LastUpdate time.Time `json:"last_update"` // 价格更新时间
}

// GetTokenBalanceAndPrice 获取指定钱包地址持有的特定代币数量和价格信息
// ownerAddress: 钱包地址（Base58格式）
// tokenAddress: 代币地址（Mint地址）
// rpcEndpoint: Solana RPC节点地址，如果为空则使用默认节点
// 返回代币余额和价格信息，以及可能的错误
func GetTokenBalanceAndPrice(ownerAddress string, tokenAddress string, rpcEndpoint string) (*TokenBalanceWithPrice, error) {
	if ownerAddress == "" || tokenAddress == "" {
		return nil, fmt.Errorf("钱包地址和代币地址不能为空")
	}

	// 获取代币账户
	tokenAccounts, err := GetTokenAccountsByOwner(ownerAddress, rpcEndpoint)
	if err != nil {
		return nil, fmt.Errorf("获取代币账户失败: %w", err)
	}

	// 查找指定代币
	var targetToken *SolanaTokenAccount
	for _, acc := range tokenAccounts {
		if acc.Mint == tokenAddress {
			targetToken = &acc
			break
		}
	}

	// 如果未找到指定代币
	if targetToken == nil {
		return nil, fmt.Errorf("未找到指定代币: %s", tokenAddress)
	}

	// 获取代币价格
	tokenPrice, err := GetTokenPrice(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("获取代币价格失败: %w", err)
	}

	// 计算持有代币的总价值
	usdValue := targetToken.UIAmount * tokenPrice.USDPrice
	solValue := targetToken.UIAmount * tokenPrice.SOLPrice

	// 构建返回结果
	result := &TokenBalanceWithPrice{
		Mint:        targetToken.Mint,
		Amount:      targetToken.Amount,
		Decimals:    targetToken.Decimals,
		UIAmount:    targetToken.UIAmount,
		UIAmountStr: targetToken.UIAmountStr,
		USDPrice:    tokenPrice.USDPrice,
		SOLPrice:    tokenPrice.SOLPrice,
		USDValue:    usdValue,
		SOLValue:    solValue,
		LastUpdate:  tokenPrice.LastUpdate,
	}

	return result, nil
}
