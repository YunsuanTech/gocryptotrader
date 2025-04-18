package utils

import (
	"encoding/hex"
	"math/big"
	"strings"
	"sync"
)

var (
	decimalCache = make(map[string]uint8)
	mu           sync.Mutex
)

func SortToken(tokenA, tokenB string) (token0 string, token1 string) {
	tokenABytes, _ := hex.DecodeString(strings.Replace(tokenA, "0x", "", -1))
	tokenBBytes, _ := hex.DecodeString(strings.Replace(tokenB, "0x", "", -1))
	if new(big.Int).SetBytes(tokenABytes).Cmp(new(big.Int).SetBytes(tokenBBytes)) < 0 {
		token0, token1 = tokenA, tokenB
	} else {
		token0, token1 = tokenB, tokenA
	}
	return
}
