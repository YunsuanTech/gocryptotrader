package xen

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	w3types "gocryptotrader/common/types"

	w3common "gocryptotrader/common/common"

	"gocryptotrader/common/utils"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// XENInfo representative execute proxy contract to mint claim-rank Info
type XENInfo struct {

	// Unique identifier of the key within the xen object.
	ID string `json:"id,omitempty"`

	// Solt of within the Proxy contract is from 0 beginning
	Solt int32 `json:"solt,omitempty"`

	// Address representative the proxy Contract address
	Address int32 `json:"address,omitempty"`

	// Start representative start solt currently
	Start int64 `json:"start,omitempty"`

	// Term representative the xen contract mint days
	Term int32 `json:"term,omitempty"`

	// Owner representative the proxy contract deployer
	Owner common.Address `json:"owner,omitempty"`

	// exectue mint claimRank hash of transaction
	tx types.Transaction `json:"tx,omitempty"`
}

var XenConfigs = map[string]Config{
	"CBJIE": {
		XenAddress:    XEN_CB_ADDRESS,
		XenABI:        XEN_ABI,
		ProxyABI:      PROXY2_ABI,
		ProxyAddress:  PROXY_CBSNIPER_JIE_ADDRESS,
		RewardAddress: RewardJieAddress,
		Shared:        100,
		Cipher:        "cDRPMTx+DWBGhgXLVi27Aqh+CZttWC0++CC1VJ//omDTXG355su+EsS4XuuHMH7+F5hKsUxFv4fEPtumq9HyhwhDrHvLffzvPTwSCSLDzNNqgoZTySQLhAdXgCkjFL15VWVQXSibV8/gk3dvrcHOaYh3rM+iMqUPOsjNjM96hDdLSrE69clmlRa7d3Ez3w/0Ji6bXP8zQmQO7FU+sf5S8LPl0YCHwzvvMOKqQrjLKmd5v5JPf1yt2SqWiHJQN30nzp9x1LYABHOL/x4yxgw+jDxGPbuEva6OW3xbiFWbVmnmK5YqcOPNVOtRDG3eYylsli1eVThwlWBxK3ItXmi7w1Ecyy8AkBIbSHzaw6IGyvwolFRFtFlURsjO2bwsLLGrA1sewzYhnIGDMmA/OfKOurwlwpMd1ufRb9GgoUpoKLng4ZPxJg3X8d18lxIYCVGZsmdfwDEasJgbjUNsaRuCzZ9cm5RssDsnbyBOpzgi1WFCUxlq3iy9Usv//N/whdHJ",
		Keys: map[string]string{
			"default": "06cb90517d35061a38342e9cbcb6",
		},
	},
	"sniper": {
		XenAddress:    XEN_ADDRESS,
		XenABI:        XEN_ABI,
		ProxyAddress:  PROXY_XJIE_ADDRESS,
		ProxyABI:      PROXY4_ABI,
		XenftABI:      XENFT_ABI,
		XenftAddress:  XENFT_ADDRESS,
		VMPXAddress:   VMPX_ETH_ADDRESS,
		VMPXABI:       VMPX_ETH_ABI,
		RewardAddress: RewardJieAddress,
		Shared:        100,
		Keys: map[string]string{
			"default": "06cb90517d35061a38342e9cbcb6",
		},
	},
}

type Config struct {
	XenABI        string
	XenftABI      string
	ProxyABI      string
	ProxyAddress  string
	XenAddress    string
	XenftAddress  string
	Keys          map[string]string
	ChainID       *big.Int
	MaxBaseFee    float64
	RewardAddress string
	Shared        int
	Cipher        string
	VMPXAddress   string
	VMPXABI       string
}

// XenMintRank sends a transaction to claim a reward for minting in the Xen system
// Parameters:
//
//	ctx - the context for the function call
//	name - the name of the configuration to use
//	client - the web3 client object
//	privateKey - the private key of the account to use
//	begin - the index of the first reward to claim
//	count - the number of rewards to claim
//	timeoutInSeconds - the timeout for the transaction
//	gas - the gas price to use for the transaction
//	params - optional parameters for the contract function call
//
// Returns:
//
//	An error if the transaction fails, otherwise nil
func BaseMint(ctx context.Context, client *Web3, xenConfig Config, privateKey string, nonce uint64, gas, tips string, maxGas float64, begin, count int, params ...interface{}) (string, *big.Float, error) {
	key := privateKey
	client.Eth.SetAccount(key)

	gasPrice := new(big.Int)
	gasPrice, ok := gasPrice.SetString(gas, 10)
	if !ok {
		log.Fatalf("Cannot convert gas to big.Int  %v", gas)
	}

	priority := new(big.Int)
	priority, _ = priority.SetString(tips, 10)

	// Load the ABI for the Xen contract
	xenABI, err := abi.JSON(strings.NewReader(xenConfig.XenABI))
	if err != nil {
		return "0x00", big.NewFloat(0), fmt.Errorf("Cannot initialize Xen ABI: %v", err)
	}

	// Get the function signature for the claimRank method
	fn := xenABI.Methods["claimRank"]
	// Convert the optional parameters to the correct Go types
	goParams, err := client.Utils.ConvertArguments(fn.Inputs, params)
	if err != nil {
		return "0x00", big.NewFloat(0), fmt.Errorf("Failed to convert arguments: %v", err)
	}
	// Pack the arguments into a byte array for the contract call
	data, err := xenABI.Pack("claimRank", goParams...)
	if err != nil {
		return "0x00", big.NewFloat(0), fmt.Errorf("Failed to pack values: %v", err)
	}

	// Load the ABI for the proxy contract
	proxyABI, err := abi.JSON(strings.NewReader(xenConfig.ProxyABI))
	if err != nil {
		return "0x00", big.NewFloat(0), fmt.Errorf("Cannot initialize Proxy ABI: %v", err)
	}

	// Pack the parameters for the execute method of the proxy contract
	xenParams := []interface{}{begin, count, xenConfig.XenAddress, data}
	fp, err := client.Utils.ConvertArguments(proxyABI.Methods["increaseAndExecute"].Inputs, xenParams)
	if err != nil {
		return "0x00", big.NewFloat(0), fmt.Errorf("Failed to convert arguments: %v", err)
	}
	fullData, err := proxyABI.Pack("increaseAndExecute", fp...)
	if err != nil {
		return "0x00", big.NewFloat(0), fmt.Errorf("Failed to pack values: %v", err)
	}

	// Estimate the gas limit for the transaction
	call := &w3types.CallMsg{
		From: client.Eth.Address(),
		To:   common.HexToAddress(xenConfig.ProxyAddress),
		Data: fullData,
		Gas:  w3types.NewCallMsgBigInt(big.NewInt(w3types.MAX_GAS_LIMIT)),
	}
	gasLimit, err := client.Eth.EstimateGas(call)
	if err != nil {
		log.Printf("Failed to estimate gas: %v", err)
		return "0x00", big.NewFloat(0), fmt.Errorf("Failed to estimate gas: %v", err)
	}

	balance, err := client.Eth.GetBalance(client.Eth.Address(), nil)
	if err != nil {
		return "0x00", big.NewFloat(0), fmt.Errorf("Failed to get balance: %v", err)
	}

	fee, err := client.Eth.EstimateFee()
	if err != nil {
		return "0x00", big.NewFloat(0), fmt.Errorf("Error estimate fee: %v", err)
	}
	util := utils.Utils{}

	baseFee := util.FromWeiWithUnit(fee.BaseFee, utils.EtherUnitGWei)
	// Convert baseFee to a float64 in GWei
	baseFeeInGWei, _ := baseFee.Float64()
	MaxFeePerGas := new(big.Int).Set(gasPrice)
	if fee.BaseFee.Cmp(gasPrice) > 0 {
		MaxFeePerGas.Set(fee.BaseFee)
	}

	log.Printf("BaseFee %v, PriorityFee %.3f Gwei\n", fee.BaseFee, util.FromWeiWithUnit(fee.MaxPriorityFeePerGas, utils.EtherUnitGWei))
	if nonce == 0 {
		nonce, err = client.Eth.GetNonce(client.Eth.Address(), nil)
		if err != nil {
			fmt.Printf("Failed to get nonce error: %v", err)
			return "0x00", big.NewFloat(0), fmt.Errorf("Failed to get nonce error: %v", err)
		}
	}
	// If the baseFee is greater than or equal to maxGas, print a message and return
	if baseFeeInGWei >= maxGas {
		fmt.Printf("Base fee is greater than or equal to %.4f GWei. Transaction cancelled.\n", maxGas)
		return "0x00", big.NewFloat(0), fmt.Errorf("Base Fee is too high")
	}

	// Calculate mint fees as gasLimit * MaxFeePerGas
	mintFeesInWei := new(big.Int).Mul(MaxFeePerGas, big.NewInt(int64(gasLimit)))
	// Convert mint fees from Wei to GWei
	mintFees := util.FromWeiWithUnit(mintFeesInWei, utils.EtherUnitGWei)

	log.Printf("Balance %v \n", util.FromWeiWithUnit(balance, utils.EtherUnitEther))
	log.Printf("Mint cbXEN Rank Prepare:%d, Team:%d, Count:%d, GasLimit:%v, Mint fees: %.3f GWei\n", begin, params[0], count, gasLimit, mintFees)

	tx, err := client.Eth.SendRawEIP1559Transaction(
		common.HexToAddress(xenConfig.ProxyAddress),
		big.NewInt(0),
		nonce,
		gasLimit,
		fee.MaxPriorityFeePerGas,
		MaxFeePerGas,
		fullData,
	)
	if err != nil {
		return "0x00", big.NewFloat(0), fmt.Errorf("error sending transaction: %v", err)
	}
	log.Printf("Mint cbXEN successfully, Transaction:{%s}, nonce: %d\n", tx.Hex(), nonce)
	return tx.Hex(), mintFees, nil
}

// GetXenConfig resolves the rpcUrl from the user specified options, or quits if an illegal combination or value is found.
func GetXenConfig(name string, testnet bool) Config {
	var xenconfig Config
	if testnet {
		if name != "" {
			w3common.FatalExit(fmt.Errorf("Cannot set both network %q and testnet", name))
		}
		name = "testnet"
	} else if name == "" {
		name = "goerli"
	}
	var ok bool
	xenconfig, ok = XenConfigs[name]
	if !ok {
		w3common.FatalExit(fmt.Errorf("Unrecognized network1 %q", name))
	}
	return xenconfig
}

// String returns the string representation of the XEN.
func (x *XENInfo) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "XENInfo:%s:%s%s", x.Solt, x.ID, x.Term)

	return buf.String()
}

func IncreaseAndMintSniper(ctx context.Context, name string, client *Web3, privateKey string, nonce uint64, begin, count int, gas, tips string, baseFeeThreshold float64, params ...interface{}) error {

	xenConfig := GetXenConfig(name, false)
	client.Eth.SetAccount(privateKey)

	gasPrice := new(big.Int)
	gasPrice, ok := gasPrice.SetString(gas, 10)
	if !ok {
		log.Fatalf("Cannot convert gas to big.Int  %v", gas)
	}

	tipsPrice := new(big.Int)
	tipsPrice, ok = tipsPrice.SetString(tips, 10)
	if !ok {
		log.Fatalf("Cannot convert tipsPrice to big.Int  %v", tipsPrice)
	}

	// Load the ABI for the Xen contract
	xenABI, err := abi.JSON(strings.NewReader(xenConfig.XenABI))
	if err != nil {
		return fmt.Errorf("Cannot initialize Xen ABI: %v", err)
	}

	// Get the function signature for the claimRank method
	fn := xenABI.Methods["claimRank"]

	// Convert the optional parameters to the correct Go types
	goParams, err := client.Utils.ConvertArguments(fn.Inputs, params)
	if err != nil {
		return fmt.Errorf("Failed to convert arguments: %v", err)
	}

	// Pack the arguments into a byte array for the contract call
	data, err := xenABI.Pack("claimRank", goParams...)
	if err != nil {
		return fmt.Errorf("Failed to pack values: %v", err)
	}
	fmt.Printf("Generate claimRank data successfully: 0x%x\n", data[:])

	// Load the ABI for the proxy contract
	proxyABI, err := abi.JSON(strings.NewReader(xenConfig.ProxyABI))
	if err != nil {
		return fmt.Errorf("Cannot initialize Proxy ABI: %v", err)
	}

	// Pack the parameters for the execute method of the proxy contract
	xenParams := []interface{}{begin, count, xenConfig.XenAddress, data}
	fp, err := client.Utils.ConvertArguments(proxyABI.Methods["increaseAndExecute"].Inputs, xenParams)
	if err != nil {
		return fmt.Errorf("Failed to convert arguments: %v", err)
	}
	fullData, err := proxyABI.Pack("increaseAndExecute", fp...)
	if err != nil {
		return fmt.Errorf("Failed to pack values: %v", err)
	}
	log.Printf("gen xen mint rank fulldata successfully: 0x%x\n", fullData[:])

	// Estimate the gas limit for the transaction
	call := &w3types.CallMsg{
		From: client.Eth.Address(),
		To:   common.HexToAddress(xenConfig.ProxyAddress),
		Data: fullData,
		Gas:  w3types.NewCallMsgBigInt(big.NewInt(w3types.MAX_GAS_LIMIT)),
	}
	gasLimit, err := client.Eth.EstimateGas(call)
	if err != nil {
		fmt.Printf("Failed to estimate gas: %v", err)
		return fmt.Errorf("Failed to estimate gas: %v", err)
	}

	fmt.Printf("Estimate gas limit %v\n", gasLimit)
	if nonce == 0 {
		nonce, err = client.Eth.GetNonce(client.Eth.Address(), nil)
		if err != nil {
			fmt.Printf("Failed to get nonce error: %v", err)
		}
	}
	fee, err := client.Eth.EstimateFee()
	if err != nil {
		w3common.FatalExit(fmt.Errorf("error estimate fee: %v", err))
	}
	util := utils.Utils{}

	balance, err := client.Eth.GetBalance(client.Eth.Address(), nil)
	fmt.Printf("base fee %v, %.3f Gwei\n", fee.BaseFee, util.FromWeiWithUnit(fee.BaseFee, utils.EtherUnitGWei))
	fmt.Printf("max priority fee per gas %v, %.3f Gwei\n", fee.MaxPriorityFeePerGas, util.FromWeiWithUnit(fee.MaxPriorityFeePerGas, utils.EtherUnitGWei))
	fmt.Printf("max fee per gas %v, %.3f Gwei\n", fee.MaxFeePerGas, util.FromWeiWithUnit(fee.MaxFeePerGas, utils.EtherUnitGWei))
	fmt.Printf("gas price manual set value %.3f Gwei, tips %.3f Gwei, Account Balance, %.3f Ether\n", util.FromWeiWithUnit(gasPrice, utils.EtherUnitGWei), util.FromWeiWithUnit(tipsPrice, utils.EtherUnitGWei), util.FromWei(balance))

	baseFee := util.FromWeiWithUnit(fee.BaseFee, utils.EtherUnitGWei)
	// Convert baseFee to a float64 in GWei
	baseFeeInGWei, _ := baseFee.Float64()

	totalFee := new(big.Int).Add(fee.BaseFee, fee.MaxPriorityFeePerGas)
	totalFee.Mul(totalFee, new(big.Int).SetUint64(gasLimit))

	optimizeFee := new(big.Int).Add(gasPrice, tipsPrice)
	optimizeFee.Mul(optimizeFee, new(big.Int).SetUint64(gasLimit))
	fmt.Printf("mint xen rank begin:%d,estimate gas Limit: %v, totalFee: %.3f Ether \n", begin, gasLimit, util.FromWei(optimizeFee))

	// If the baseFee is greater than or equal to 40 GWei, print a message and return
	if baseFeeInGWei >= baseFeeThreshold {
		fmt.Printf("Base fee is greater than or equal to %.0f GWei. Transaction cancelled.\n", baseFeeThreshold)
		return nil
	}
	receipt, err := client.Eth.SyncSendEIP1559RawTransaction(
		common.HexToAddress(xenConfig.ProxyAddress),
		big.NewInt(0),
		nonce,
		gasLimit,
		tipsPrice,
		gasPrice,
		fullData,
	)
	if err != nil {
		w3common.FatalExit(fmt.Errorf("error sending transaction: %v", err))
	}
	fmt.Printf("Mint xen successfully, gasLimit: %v, transaction: %s\n", gasLimit, receipt.TxHash.Hex())
	return nil
}
