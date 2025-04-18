package xen

import (
	"context"
	"fmt"

	w3types "gocryptotrader/common/types"
	"gocryptotrader/common/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"math/big"

	scommon "gocryptotrader/common/common"
	"gocryptotrader/common/rpc"
	"gocryptotrader/exchanges/eth"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

var (
	GasEstimationBuffer uint64 = 1000000

	ETH_token  string = "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
	XEN_token  string = "0x06450dee7fd2fb8e39061434babcfc05599a6fb8"
	USDC_token string = "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
	USDT_token string = "0xdac17f958d2ee523a2206206994597c13d831ec7"
	ZZ_token   string = "0xc91a71a1ffa3d8b22ba615ba1b9c01b2bbbf55ad"
	GUNI_token string = "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984"
	GETH_token string = "0x3fC91A3afd70395Cd496C647d5a6CC9D4B2b7FAD"
)

// CoinToAddressMap returns a mapping from coin to address
var CoinToAddressMap = map[coinType]common.Address{
	ETH:  common.HexToAddress("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"),
	USDC: common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"),
	USDT: common.HexToAddress("0xdac17f958d2ee523a2206206994597c13d831ec7"),
}

const (
	// ETH is ether.
	ETH coinType = iota
	// USDC is the stable coin by Circle.
	USDC coinType = iota
	// USDT is the stable coin.
	USDT coinType = iota
)

type coinType int

type Web3 struct {
	Eth   *eth.Eth
	Utils *utils.Utils
	c     *rpc.Client

	network scommon.Network
}

func (w *Web3) RpcClient() *rpc.Client {
	return w.c
}

func NewWeb3(network scommon.Network) (*Web3, error) {
	return NewWeb3WithProxy(network, "")
}

func NewWeb3WithProxy(network scommon.Network, proxy string) (*Web3, error) {
	c, err := rpc.NewClient(network.URL, proxy)
	if err != nil {
		return nil, err
	}
	e := eth.NewEth(c)
	e.SetChainId(network.ChainID.Int64())

	u := utils.NewUtils()
	w := &Web3{
		Eth:   e,
		Utils: u,
		c:     c,
	}
	// Default poll timeout 2 Minutes
	w.Eth.SetTxPollTimeout(120)
	return w, nil
}

// TransactionCallMessage returns a filled Ethereum CallMsg object with suggest gas price and limit
func (w *Web3) TransactionCallMessage(
	to common.Address,
	value *big.Int,
	data common.Hash,
) (*ethereum.CallMsg, error) {
	gasPrice, err := w.Eth.GasPrice()
	if err != nil {
		return nil, err
	}

	// Estimate the gas limit for the transaction
	msg := &w3types.CallMsg{
		From:     w.Eth.Address(),
		To:       to,
		Data:     data.Bytes(),
		GasPrice: w3types.NewCallMsgBigInt(big.NewInt(int64(gasPrice))),
		Gas:      w3types.NewCallMsgBigInt(big.NewInt(w3types.MAX_GAS_LIMIT)),
		Value:    w3types.NewCallMsgBigInt(value),
	}
	gasLimit, err := w.Eth.EstimateGas(msg)
	if err != nil {
		return nil, err
	}
	limit := gasLimit + GasEstimationBuffer

	emsg := &ethereum.CallMsg{
		From:     w.Eth.Address(),
		To:       &to,
		GasPrice: big.NewInt(int64(gasPrice)),
		Value:    value,
		Data:     data.Bytes(),
		Gas:      limit,
	}

	return emsg, nil
}

// TransactionOpts return the base binding transaction options to create a new valid tx for contract deployment
func (e *Web3) TransactionOpts(
	to common.Address,
	value *big.Int,
	data common.Hash,
) (*bind.TransactOpts, error) {
	callMsg, err := e.TransactionCallMessage(to, value, data)
	if err != nil {
		fmt.Printf("callMsg, err := e.TransactionCallMessage,to:%v,value:%d,data:%v error:%v\n", to, value, data, err)
		return nil, err
	}
	nonce, err := e.Eth.GetNonce(e.Eth.Address(), nil)
	if err != nil {
		return nil, err
	}
	privateKey := e.Eth.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}
	chainId, _ := e.Eth.ChainID()
	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		return nil, err
	}
	opts.From = callMsg.From
	opts.Nonce = big.NewInt(int64(nonce))
	opts.Value = value
	opts.GasPrice = callMsg.GasPrice
	opts.GasLimit = callMsg.Gas
	opts.Context = context.Background()

	return opts, nil
}

func (e *Web3) TransactionOptsWithCtx(
	to common.Address,
	value *big.Int,
	data common.Hash,
	ctx context.Context,
) (*bind.TransactOpts, error) {
	callMsg, err := e.TransactionCallMessage(to, value, data)
	if err != nil {
		return nil, err
	}
	nonce, err := e.Eth.GetNonce(e.Eth.Address(), nil)
	if err != nil {
		return nil, err
	}
	privateKey := e.Eth.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}
	chainId, _ := e.Eth.ChainID()

	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		return nil, err
	}
	opts.From = callMsg.From
	opts.Nonce = big.NewInt(int64(nonce))
	opts.Value = value
	opts.GasPrice = callMsg.GasPrice
	opts.GasLimit = callMsg.Gas
	opts.Context = ctx

	return opts, nil
}
