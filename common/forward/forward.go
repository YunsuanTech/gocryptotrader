package forward

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"gocryptotrader/config"
	"gocryptotrader/database/repository/solana_transfer"
	"gocryptotrader/log"

	"github.com/donutnomad/solana-web3/web3"
	"github.com/donutnomad/solana-web3/web3kit"
)

// Manager 管理SOL转发相关操作
type Manager struct {
	config *config.Config
}

// New 创建一个新的SOL转发管理器
func New(cfg *config.Config) *Manager {
	return &Manager{config: cfg}
}

// TransferSOL 将 SOL 发送到多个地址
func (m *Manager) TransferSOL(ctx context.Context, req *ForwardRequest) ([]string, error) {
	// 解析私钥
	privateKey := web3.Keypair.FromBase58(req.PrivateKeyStr)

	// 将 SOL 金额转换为 lamports（1 SOL = 10^9 lamports）
	amountLamport := uint64(req.AmountSOL * 1e9)

	// 创建 RPC 客户端
	client, err := web3.NewConnection(req.Config.RPCEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 RPC 客户端失败: %w", err)
	}

	// 并发发送交易
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, req.Config.ConcurrentTxs)
	var mu sync.Mutex
	var txSignatures []string

	for i := 0; i < len(req.Addresses); i++ {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			// 解析目标地址
			toStr := req.Addresses[idx]
			to := web3.MustPublicKey(toStr)
			if err != nil {
				log.Warnf(log.Global, "无效地址: %s，已跳过", toStr)
				return
			}

			// 发送转账交易
			signature, err := web3kit.Token.Transfer(
				ctx,
				client,
				privateKey,           // 签名密钥
				privateKey,           // 发送者
				to,                   // 接收者
				web3.PublicKey{},     // 无 token mint（原生 SOL 转账）
				amountLamport,        // 转账金额（lamports）
				web3.SystemProgramID, // 系统程序 ID
				true,                 // 创建关联账户（SOL 不需要，设为 true 无影响）
				web3.ConfirmOptions{
					SkipPreflight: web3.Ref(false), // 不跳过预检
				},
			)
			if err != nil {
				log.Errorf(log.Global, "发送交易失败 (to: %s): %v", toStr, err)
				return
			}

			signatureStr := signature.String()
			log.Infof(log.Global, "交易已发送: %s", signatureStr)

			err = solana_transfer.RecordSOLTransfer(
				signatureStr,
				req.Config.RPCEndpoint,
				time.Now(),
				privateKey.PublicKey().String(),
				toStr,
				req.AmountSOL,
			)
			if err != nil {
				log.Errorf(log.Global, "记录交易失败: %v", err)
			}

			mu.Lock()
			txSignatures = append(txSignatures, signatureStr)
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	return txSignatures, nil
}

// TransferToken 将代币发送到多个地址
func (m *Manager) TransferToken(ctx context.Context, req *TokenForwardRequest) ([]string, error) {
	// 解析私钥
	privateKey := web3.Keypair.FromBase58(req.PrivateKeyStr)

	// 解析代币铸币地址
	tokenMint := web3.MustPublicKey(req.TokenMint)

	// 创建 RPC 客户端
	commitment := web3.CommitmentProcessed
	client, err := web3.NewConnection(req.Config.RPCEndpoint, &web3.ConnectionConfig{
		Commitment: &commitment,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 RPC 客户端失败: %w", err)
	}

	// 计算转账金额（根据代币精度）
	amountRaw := uint64(req.Amount * math.Pow10(int(9)))

	var TokenProgram = web3.TokenProgram2022ID

	if !req.IsToken2022 {
		TokenProgram = web3.TokenProgramID
	}

	// 并发发送交易
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, req.Config.ConcurrentTxs)
	var mu sync.Mutex
	var txSignatures []string

	for i := 0; i < len(req.Addresses); i++ {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			// 解析目标地址
			toStr := req.Addresses[idx]
			to := web3.MustPublicKey(toStr)

			// 发送代币转账交易
			signature, err := web3kit.Token.Transfer(
				ctx,
				client,
				privateKey,   // 签名密钥
				privateKey,   // 发送者
				to,           // 接收者
				tokenMint,    // 代币铸币地址
				amountRaw,    // 转账金额
				TokenProgram, // 代币程序 ID，支持 Token-2022
				true,         // 自动创建关联账户
				web3.ConfirmOptions{
					SkipPreflight:       web3.Ref(false), // 不跳过预检
					PreflightCommitment: &commitment,     // 预检确认级别
					Commitment:          &commitment,     // 交易确认级别
				},
			)
			if err != nil {
				log.Errorf(log.Global, "发送交易失败 (to: %s): %v", toStr, err)
				return
			}

			signatureStr := signature.String()
			log.Infof(log.Global, "交易已发送: %s", signatureStr)

			err = solana_transfer.RecordTokenTransfer(
				signatureStr,
				req.Config.RPCEndpoint,
				time.Now(),
				privateKey.PublicKey().String(),
				toStr,
				req.Amount,
				tokenMint.String(),
			)
			if err != nil {
				log.Errorf(log.Global, "记录交易失败: %v", err)
			}

			mu.Lock()
			txSignatures = append(txSignatures, signatureStr)
			mu.Unlock()
		}(i)
	}

	wg.Wait()
	return txSignatures, nil
}

// ReadAddressesFromFile 从文件中读取目标地址列表
func ReadAddressesFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var addresses []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		addr := scanner.Text()
		if addr != "" {
			addresses = append(addresses, addr)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件出错: %w", err)
	}
	return addresses, nil
}
