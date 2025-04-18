package common

import (
	"fmt"
	"math/big"
)

const (
	bscExplorerURL = "https://bscscan.com/"
	//etherRPC       = "https://go.getblock.io/ac12af33fe2a4ea39fff1fe92d2c0054"
	etherRPC = "https://eth-mainnet.g.alchemy.com/v2/i4_pCkMl1chn_XT0EXxnYynnabEj-PM8"
	//etherRPC       = "https://eth-mainnet.g.alchemy.com/v2/TfUsGyA_ATs3JCVRRUsiro4N-mS0i22P"
	bscURL   = "https://binance.rpc.thirdweb.com/"
	maticURL = "https://polygon-mainnet.g.alchemy.com/v2/i4_pCkMl1chn_XT0EXxnYynnabEj-PM8"
	//maticURL  = "https://polygon-rpc.com"
	gnosisURL = "https://rpc.gnosischain.com"
	goerliURL = "https://goerli.infura.io/v3/c3c7195ff645422396aaef78a10e1a8f"
	//opt = "https://opt-mainnet.g.alchemy.com/v2/UU5jTyd5du9THVtrzN-ASCuV0hRuuJ-I"
	opt       = "https://opt-mainnet.g.alchemy.com/v2/i4_pCkMl1chn_XT0EXxnYynnabEj-PM8"
	Arbitrum  = "https://arb-mainnet.g.alchemy.com/v2/6FBye5s_dwrlwrBT6V_5mzvD3Z_0Jy4O"
	flashbots = "https://rpc.flashbots.net"
	pulse     = "https://rpc.pulsechain.com"
	x1dev     = "https://x1-devnet.xen.network"
	x1fast    = "https://x1-fastnet.infrafc.org"
	x1test    = "https://x1-testnet.xen.network"
	sepolia   = "https://eth-sepolia.g.alchemy.com/v2/QNf4aV81pRHRevFc08t60OJZYDVss90t"
	oktest    = "https://testrpc.x1.tech"
	zkSync    = "https://zksync2-mainnet.zksync.io"
	zkEVM     = "https://polygon-zkevm.rpc.thirdweb.com"
	mantleRPC = "https://rpc.mantle.xyz"
	// baseRPC     = "https://mainnet.base.org"
	baseRPC = "https://base-mainnet.g.alchemy.com/v2/i4_pCkMl1chn_XT0EXxnYynnabEj-PM8"

	FilecoinRPC = "https://api.node.glif.io/rpc/v1"
)

var Networks = map[string]Network{
	"bsc": {
		Name:        "bsc",
		URL:         bscURL,
		ChainID:     big.NewInt(56),
		Unit:        "BNB",
		ExplorerURL: bscExplorerURL,
	},
	"localhost": {
		Name:    "localhost",
		ChainID: big.NewInt(8545),
		URL:     "http://localhost:8545",
		Unit:    "ETH",
	},
	"Ethereum": {
		Name:        "Ethereum",
		URL:         etherRPC,
		ChainID:     big.NewInt(1),
		Unit:        "ETH",
		ExplorerURL: "https://etherscan.io",
	},
	"Matic": {
		Name:        "Matic",
		URL:         maticURL,
		ChainID:     big.NewInt(137),
		Unit:        "MATIC",
		ExplorerURL: "https://polygonscan.com",
	},
	"Gnosis": {
		Name:        "Gnosis",
		URL:         gnosisURL,
		ChainID:     big.NewInt(100),
		Unit:        "xDAI",
		ExplorerURL: "https://rpc.gnosischain.com",
	},
	"sepolia": {
		Name:    "sepolia",
		URL:     sepolia,
		ChainID: big.NewInt(11155111),
		Unit:    "ETH",
	},
	"oktest": {
		Name:    "oktest",
		URL:     oktest,
		ChainID: big.NewInt(195),
		Unit:    "OKB",
	},
	"Optimism": {
		Name:    "Optimism",
		URL:     opt,
		ChainID: big.NewInt(10),
		Unit:    "OETH",
	},
	"Arbitrum": {
		Name:    "Arbitrum",
		URL:     Arbitrum,
		ChainID: big.NewInt(42161),
		Unit:    "AETH",
	}, "pulse": {
		Name:    "pulse",
		URL:     pulse,
		ChainID: big.NewInt(369),
		Unit:    "PLS",
	}, "zkSync": {
		Name:    "zkSync",
		URL:     zkSync,
		ChainID: big.NewInt(324),
		Unit:    "ZETH",
	}, "base": {
		Name:    "base",
		URL:     baseRPC,
		ChainID: big.NewInt(8453),
		Unit:    "BETH",
	}, "filecoin": {
		Name:    "filecoin",
		URL:     FilecoinRPC,
		ChainID: big.NewInt(314),
		Unit:    "FIL",
	},
}

type Network struct {
	Name        string
	URL         string
	ExplorerURL string
	Unit        string
	ChainID     *big.Int
}

// GetNetwork resolves the rpcUrl from the user specified options, or quits if an illegal combination or value is found.
func GetNetwork(name, rpcURL string, testnet bool) Network {

	var network Network
	if rpcURL != "" {
		if name != "" {
			FatalExit(fmt.Errorf("Cannot set both rpcURL %q and network %q", rpcURL, network))
		}
		if testnet {
			FatalExit(fmt.Errorf("Cannot set both rpcURL %q and testnet", rpcURL))
		}
		network.URL = rpcURL
		network.Unit = "ARB"
	} else {
		if testnet {
			if name != "" {
				FatalExit(fmt.Errorf("Cannot set both network %q and testnet", name))
			}
			name = "testnet"
		} else if name == "" {
			name = "goerli"
		}
		var ok bool
		network, ok = Networks[name]
		if !ok {
			FatalExit(fmt.Errorf("Unrecognized network %q", name))
		}
	}
	return network
}
