package token

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gagliardetto/solana-go"
)

// SolanaSwapConfig 包含Solana代币交换所需的配置参数
type SolanaSwapConfig struct {
	APIHost        string        // API服务器地址
	InputToken     string        // 输入代币地址
	OutputToken    string        // 输出代币地址
	Amount         string        // 交易金额（以lamports为单位）
	FromAddress    string        // 发送方地址
	Slippage       float64       // 滑点百分比
	PollInterval   time.Duration // 轮询间隔
	RequestTimeout time.Duration // 请求超时时间
}

// DefaultSolanaSwapConfig 返回默认的Solana交换配置
func DefaultSolanaSwapConfig() *SolanaSwapConfig {
	return &SolanaSwapConfig{
		APIHost:        "https://gmgn.ai",
		InputToken:     "So11111111111111111111111111111111111111112",
		OutputToken:    "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		Amount:         "10000000",
		FromAddress:    "3yQKDUwAgobWxooGbFWaGQttRV6AuUWw7zz4ip3iTfzc",
		Slippage:       0.5,
		PollInterval:   1 * time.Second,
		RequestTimeout: 10 * time.Second,
	}
}

// API 响应结构体
type (
	RouteResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			RawTx struct {
				SwapTransaction      string `json:"swapTransaction"`
				LastValidBlockHeight int    `json:"lastValidBlockHeight"`
			} `json:"raw_tx"`
		} `json:"data"`
	}

	SubmitResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Hash string `json:"hash"`
		} `json:"data"`
	}

	StatusResponse struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Success bool `json:"success"`
			Expired bool `json:"expired"`
		} `json:"data"`
	}
)

// SolanaSwapResult 包含Solana代币交换的结果信息
type SolanaSwapResult struct {
	SignedTransaction string `json:"signed_transaction"` // 签名后的交易
	TransactionHash   string `json:"transaction_hash"`   // 交易哈希
	Status            string `json:"status"`             // 交易状态
}

// ExecuteSolanaSwap 执行Solana代币交换操作
// privateKey: 用户的私钥（Base58格式）
// config: 交换配置参数
// 返回交换结果和可能的错误
func ExecuteSolanaSwap(privateKey string, config *SolanaSwapConfig) (*SolanaSwapResult, error) {
	if privateKey == "" {
		return nil, fmt.Errorf("必须提供私钥")
	}

	// 使用默认配置（如果未提供）
	if config == nil {
		config = DefaultSolanaSwapConfig()
	}

	// 初始化钱包
	wallet, err := solana.WalletFromPrivateKeyBase58(privateKey)
	if err != nil {
		return nil, fmt.Errorf("钱包初始化失败: %w", err)
	}

	// 获取交易路由
	route, err := getSwapRoute(config)
	if err != nil {
		return nil, fmt.Errorf("获取路由失败: %w", err)
	}

	// 签名交易
	signedTx, err := signTransaction(wallet.PrivateKey, route.Data.RawTx.SwapTransaction)
	if err != nil {
		return nil, fmt.Errorf("交易签名失败: %w", err)
	}

	// 提交交易
	submitResp, err := submitTransaction(signedTx, config)
	if err != nil {
		return nil, fmt.Errorf("提交交易失败: %w", err)
	}

	// 轮询交易状态
	if err := pollTransactionStatus(submitResp.Data.Hash, route.Data.RawTx.LastValidBlockHeight, config); err != nil {
		return nil, fmt.Errorf("状态检查失败: %w", err)
	}

	// 交易成功后，返回结果
	result := &SolanaSwapResult{
		SignedTransaction: signedTx,
		TransactionHash:   submitResp.Data.Hash,
		Status:            "success",
	}

	return result, nil
}

// 获取交易路由
func getSwapRoute(config *SolanaSwapConfig) (*RouteResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer cancel()

	url := fmt.Sprintf("%s/defi/router/v1/sol/tx/get_swap_route?token_in_address=%s&token_out_address=%s&in_amount=%s&from_address=%s&slippage=%f",
		config.APIHost, config.InputToken, config.OutputToken, config.Amount, config.FromAddress, config.Slippage)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("非200状态码: %d", resp.StatusCode)
	}

	var route RouteResponse
	if err := json.NewDecoder(resp.Body).Decode(&route); err != nil {
		return nil, err
	}

	if route.Code != 0 {
		return nil, fmt.Errorf("API错误: %s", route.Msg)
	}

	return &route, nil
}

// 签名交易
func signTransaction(privateKey solana.PrivateKey, swapTxBase64 string) (string, error) {
	// 解码Base64
	txBytes, err := base64.StdEncoding.DecodeString(swapTxBase64)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	// 反序列化交易
	tx, err := solana.TransactionFromBytes(txBytes)
	if err != nil {
		return "", fmt.Errorf("交易反序列化失败: %w", err)
	}

	// 签名交易
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(privateKey.PublicKey()) {
			return &privateKey
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}

	// 序列化签名后的交易
	signedTxBytes, err := tx.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("交易序列化失败: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signedTxBytes), nil
}

// 提交交易
func submitTransaction(signedTx string, config *SolanaSwapConfig) (*SubmitResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer cancel()

	body, _ := json.Marshal(map[string]string{"signed_tx": signedTx})
	req, _ := http.NewRequestWithContext(ctx, "POST", config.APIHost+"/defi/router/v1/sol/tx/submit_signed_transaction", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("提交失败: %s", string(data))
	}

	var submitResp SubmitResponse
	if err := json.NewDecoder(resp.Body).Decode(&submitResp); err != nil {
		return nil, err
	}

	if submitResp.Code != 0 {
		return nil, fmt.Errorf("提交错误: %s", submitResp.Msg)
	}

	return &submitResp, nil
}

// 轮询交易状态
func pollTransactionStatus(hash string, lastValidHeight int, config *SolanaSwapConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ticker := time.NewTicker(config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("状态检查超时")

		case <-ticker.C:
			status, err := getTransactionStatus(hash, lastValidHeight, config)
			if err != nil {
				log.Printf("状态检查失败: %v", err)
				continue
			}

			if status.Data.Success {
				log.Println("交易已确认")
				return nil
			}
			if status.Data.Expired {
				return fmt.Errorf("交易已过期")
			}
		}
	}
}

// 获取交易状态
func getTransactionStatus(hash string, lastValidHeight int, config *SolanaSwapConfig) (*StatusResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer cancel()

	url := fmt.Sprintf("%s/defi/router/v1/sol/tx/get_transaction_status?hash=%s&last_valid_height=%d",
		config.APIHost, hash, lastValidHeight)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var status StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	if status.Code != 0 {
		return nil, fmt.Errorf("状态检查错误: %s", status.Msg)
	}

	return &status, nil
}
