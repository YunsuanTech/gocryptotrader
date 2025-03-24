package forward

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"sync"

	"gocryptotrader/config"
	"gocryptotrader/log"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
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
	privateKey, err := solana.PrivateKeyFromBase58(req.PrivateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("无效的私钥: %w", err)
	}
	from := privateKey.PublicKey()

	// 将 SOL 金额转换为 lamports（1 SOL = 10^9 lamports）
	amountLamport := uint64(req.Config.AmountSOL * 1e9)

	// 创建 RPC 客户端
	rpcClient := rpc.New(req.Config.RPCEndpoint)

	// 获取最新的 blockhash
	recent, err := rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("获取最新 blockhash 失败: %w", err)
	}

	// 将地址分组，每组生成一笔交易
	var txs []*solana.Transaction
	for i := 0; i < len(req.Addresses); i += req.Config.MaxInstructionsPerTx {
		end := i + req.Config.MaxInstructionsPerTx
		if end > len(req.Addresses) {
			end = len(req.Addresses)
		}
		group := req.Addresses[i:end]

		var instructions []solana.Instruction
		for _, toStr := range group {
			to, err := solana.PublicKeyFromBase58(toStr)
			if err != nil {
				log.Warnf(log.Global, "无效地址: %s，已跳过", toStr)
				continue
			}
			ix := system.NewTransferInstruction(amountLamport, from, to).Build()
			instructions = append(instructions, ix)
		}

		tx, err := solana.NewTransaction(instructions, recent.Value.Blockhash, solana.TransactionPayer(from))
		if err != nil {
			log.Errorf(log.Global, "创建交易失败: %v", err)
			continue
		}
		txs = append(txs, tx)
	}

	// 并发发送交易
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, req.Config.ConcurrentTxs)
	var mu sync.Mutex
	var txSignatures []string

	for _, tx := range txs {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(tx *solana.Transaction) {
			defer wg.Done()
			defer func() { <-semaphore }()

			_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
				if key.Equals(from) {
					return &privateKey
				}
				return nil
			})
			if err != nil {
				log.Errorf(log.Global, "签名交易失败: %v", err)
				return
			}

			txSig, err := rpcClient.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
				SkipPreflight:       false,
				PreflightCommitment: rpc.CommitmentFinalized,
			})
			if err != nil {
				log.Errorf(log.Global, "发送交易失败: %v", err)
				return
			}
			log.Infof(log.Global, "交易已发送: %s", txSig)

			mu.Lock()
			txSignatures = append(txSignatures, txSig.String())
			mu.Unlock()
		}(tx)
	}

	wg.Wait()
	return txSignatures, nil
}

// TransferToken 将代币发送到多个地址
func (m *Manager) TransferToken(ctx context.Context, req *TokenForwardRequest) ([]string, error) {
	// 解析私钥
	privateKey, err := solana.PrivateKeyFromBase58(req.PrivateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("无效的私钥: %w", err)
	}
	from := privateKey.PublicKey()

	// 解析代币铸币地址
	tokenMint, err := solana.PublicKeyFromBase58(req.TokenMint)
	if err != nil {
		return nil, fmt.Errorf("无效的代币铸币地址: %w", err)
	}

	// 创建 RPC 客户端
	rpcClient := rpc.New(req.Config.RPCEndpoint)

	// 获取代币精度信息
	tokenInfo, err := rpcClient.GetTokenSupply(ctx, tokenMint, rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("获取代币信息失败: %w", err)
	}

	// 计算转账金额（根据代币精度）
	decimals := tokenInfo.Value.Decimals
	amountRaw := uint64(req.Config.Amount * math.Pow10(int(decimals)))
	// 获取发送者的代币账户
	senderTokenAccount, _, err := solana.FindAssociatedTokenAddress(from, tokenMint)
	if err != nil {
		return nil, fmt.Errorf("查找发送者代币账户失败: %w", err)
	}

	// 获取最新的 blockhash
	recent, err := rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("获取最新 blockhash 失败: %w", err)
	}

	// 将地址分组，每组生成一笔交易
	var txs []*solana.Transaction
	for i := 0; i < len(req.Addresses); i += req.Config.MaxInstructionsPerTx {
		end := i + req.Config.MaxInstructionsPerTx
		if end > len(req.Addresses) {
			end = len(req.Addresses)
		}
		group := req.Addresses[i:end]

		var instructions []solana.Instruction

		for _, toStr := range group {
			to, err := solana.PublicKeyFromBase58(toStr)
			if err != nil {
				log.Warnf(log.Global, "无效地址: %s，已跳过", toStr)
				continue
			}

			// 使用标准 ATA
			recipientTokenAccount, _, err := solana.FindAssociatedTokenAddress(to, tokenMint)
			if err != nil {
				log.Warnf(log.Global, "查找接收者代币账户失败: %v，已跳过", toStr)
				continue
			}
			// 检查账户是否存在
			accountInfo, err := rpcClient.GetAccountInfo(ctx, recipientTokenAccount)
			if err != nil || accountInfo.Value == nil || accountInfo.Value.Owner.IsZero() {
				if req.Config.CreateAccountIfNotExist {
					createIx := associatedtokenaccount.NewCreateInstruction(from, to, tokenMint).Build()
					instructions = append(instructions, createIx)
				} else {
					log.Warnf(log.Global, "接收者代币账户不存在且未配置自动创建: %s，已跳过", toStr)
					continue
				}
			}

			// 创建转账指令
			transferIx := token.NewTransferInstruction(
				amountRaw,
				senderTokenAccount,
				recipientTokenAccount,
				from,
				[]solana.PublicKey{},
			).Build()
			instructions = append(instructions, transferIx)
		}

		if len(instructions) > 0 {
			tx, err := solana.NewTransaction(instructions, recent.Value.Blockhash, solana.TransactionPayer(from))
			if err != nil {
				log.Errorf(log.Global, "创建交易失败: %v", err)
				continue
			}
			fmt.Println("Required signers:", tx.Signatures) // 打印签名者

			txs = append(txs, tx)
		}
	}

	// 并发发送交易
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, req.Config.ConcurrentTxs)
	var mu sync.Mutex
	var txSignatures []string

	for _, tx := range txs {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(tx *solana.Transaction) {
			defer wg.Done()
			defer func() { <-semaphore }()

			_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
				if key.Equals(from) {
					return &privateKey
				}
				return nil
			})
			if err != nil {
				log.Errorf(log.Global, "签名交易失败: %v", err)
				return
			}

			txSig, err := rpcClient.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
				SkipPreflight:       false,
				PreflightCommitment: rpc.CommitmentFinalized,
			})
			if err != nil {
				log.Errorf(log.Global, "发送交易失败: %v", err)
				return
			}
			log.Infof(log.Global, "交易已发送: %s", txSig)

			mu.Lock()
			txSignatures = append(txSignatures, txSig.String())
			mu.Unlock()
		}(tx)
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
