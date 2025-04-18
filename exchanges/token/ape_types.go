package token

// BaseAsset 对应 Rust 的 BaseAsset 结构
type BaseAsset struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Symbol       string   `json:"symbol"`
	Icon         *string  `json:"icon,omitempty"`
	Decimals     uint8    `json:"decimals"`
	Website      *string  `json:"website,omitempty"`
	Dev          *string  `json:"dev,omitempty"`
	USDPrice     *float64 `json:"usdPrice,omitempty"`
	NativePrice  *float64 `json:"nativePrice,omitempty"`
	PoolAmount   *float64 `json:"poolAmount,omitempty"`
	CircSupply   *float64 `json:"circSupply,omitempty"`
	TotalSupply  *float64 `json:"totalSupply,omitempty"`
	FDV          *float64 `json:"fdv,omitempty"`
	MCap         *float64 `json:"mcap,omitempty"`
	Launchpad    *string  `json:"launchpad,omitempty"`
	TokenProgram *string  `json:"tokenProgram,omitempty"`
	DevMintCount *uint32  `json:"devMintCount,omitempty"`
}

// QuoteAsset 对应 Rust 的 QuoteAsset 结构
type QuoteAsset struct {
	ID         string   `json:"id"`
	Symbol     string   `json:"symbol"`
	Decimals   uint8    `json:"decimals"`
	PoolAmount *float64 `json:"poolAmount,omitempty"`
}

// Audit 对应 Rust 的 Audit 结构
type Audit struct {
	MintAuthorityDisabled   bool    `json:"mintAuthorityDisabled"`
	FreezeAuthorityDisabled bool    `json:"freezeAuthorityDisabled"`
	TopHoldersPercentage    float64 `json:"topHoldersPercentage"`
	LPBurnedPercentage      float64 `json:"lpBurnedPercentage"`
}

// Stats 对应 Rust 的 Stats 结构
type Stats struct {
	PriceChange float64 `json:"priceChange"`
	BuyVolume   float64 `json:"buyVolume"`
	SellVolume  float64 `json:"sellVolume"`
	NumBuys     uint32  `json:"numBuys"`
	NumSells    uint32  `json:"numSells"`
	NumTraders  uint32  `json:"numTraders"`
	NumBuyers   uint32  `json:"numBuyers"`
	NumSellers  uint32  `json:"numSellers"`
}

// Pool 表示单个代币池数据
type Pool struct {
	ID           string     `json:"id"`
	Chain        string     `json:"chain"`
	Dex          string     `json:"dex"`
	Type         string     `json:"type"`
	BaseAsset    BaseAsset  `json:"baseAsset"`
	QuoteAsset   QuoteAsset `json:"quoteAsset"`
	Audit        Audit      `json:"audit"`
	CreatedAt    string     `json:"createdAt"`
	IsUnreliable bool       `json:"isUnreliable,omitempty"`
	UpdatedAt    string     `json:"updatedAt"`
	Liquidity    float64    `json:"liquidity,omitempty"`
	Stats5m      Stats      `json:"stats5m,omitempty"`
	Stats1h      Stats      `json:"stats1h,omitempty"`
	Stats6h      Stats      `json:"stats6h,omitempty"`
	Stats24h     Stats      `json:"stats24h,omitempty"`
	BondingCurve float64    `json:"bondingCurve,omitempty"`
}

// PoolCategory 表示分类的代币池数据
type PoolCategory struct {
	Pools []Pool `json:"pools"`
}

// APIResponse 表示完整的API响应结构
type APIResponse struct {
	New             PoolCategory `json:"new"`
	AboutToGraduate PoolCategory `json:"aboutToGraduate"`
	Graduated       PoolCategory `json:"graduated"`
}
