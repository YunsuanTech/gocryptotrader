package token

// GMGNResponse 表示GMGN API的响应结构
type GMGNResponse struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data GMGNRankData `json:"data"`
}

// GMGNRankData 表示GMGN排名数据
type GMGNRankData struct {
	Pumps        []interface{} `json:"pumps"`
	NewCreations []interface{} `json:"new_creations"`
	Completeds   []GMGNToken   `json:"completeds"`
}

// GMGNToken 表示GMGN代币信息
type GMGNToken struct {
	Symbol                   string      `json:"symbol"`
	Name                     string      `json:"name"`
	Logo                     string      `json:"logo"`
	TotalSupply              int64       `json:"total_supply"`
	Price                    string      `json:"price"`
	HolderCount              int         `json:"holder_count"`
	PriceChangePercent1m     string      `json:"price_change_percent1m"`
	PriceChangePercent5m     string      `json:"price_change_percent5m"`
	PriceChangePercent1h     string      `json:"price_change_percent1h"`
	BurnRatio                string      `json:"burn_ratio"`
	BurnStatus               string      `json:"burn_status"`
	IsShowAlert              bool        `json:"is_show_alert"`
	HotLevel                 int         `json:"hot_level"`
	Liquidity                string      `json:"liquidity"`
	Top10HolderRate          float64     `json:"top_10_holder_rate"`
	RenouncedMint            int         `json:"renounced_mint"`
	RenouncedFreezeAccount   int         `json:"renounced_freeze_account"`
	DexscrUpdateLink         int         `json:"dexscr_update_link,omitempty"`
	CtoFlag                  int         `json:"cto_flag,omitempty"`
	TwitterRenameCount       int         `json:"twitter_rename_count"`
	RugRatio                 float64     `json:"rug_ratio,omitempty"`
	SniperCount              int         `json:"sniper_count"`
	SmartDegenCount          int         `json:"smart_degen_count"`
	RenownedCount            int         `json:"renowned_count"`
	MarketCap                string      `json:"market_cap"`
	IsWashTrading            bool        `json:"is_wash_trading"`
	Creator                  string      `json:"creator"`
	CreatorCreatedInnerCount int         `json:"creator_created_inner_count"`
	CreatorCreatedOpenCount  int         `json:"creator_created_open_count"`
	CreatorBalanceRate       interface{} `json:"creator_balance_rate"`
	CreatorTokenStatus       string      `json:"creator_token_status"`
	RatTraderAmountRate      interface{} `json:"rat_trader_amount_rate"`
	BluechipOwnerPercentage  float64     `json:"bluechip_owner_percentage"`
	Volume                   string      `json:"volume"`
	Swaps                    int         `json:"swaps"`
	Buys                     int         `json:"buys"`
	Sells                    int         `json:"sells"`
	BuyTax                   interface{} `json:"buy_tax"`
	SellTax                  interface{} `json:"sell_tax"`
	IsHoneypot               interface{} `json:"is_honeypot"`
	Renounced                interface{} `json:"renounced"`
	DevTokenBurnAmount       interface{} `json:"dev_token_burn_amount"`
	DevTokenBurnRatio        interface{} `json:"dev_token_burn_ratio"`
	DexscrAd                 int         `json:"dexscr_ad,omitempty"`
	TwitterChangeFlag        int         `json:"twitter_change_flag,omitempty"`
	Address                  string      `json:"address"`
	Twitter                  string      `json:"twitter"`
	Website                  string      `json:"website"`
	Telegram                 string      `json:"telegram"`
	OpenTimestamp            int64       `json:"open_timestamp"`
	CreatedTimestamp         int64       `json:"created_timestamp"`
	UsdMarketCap             string      `json:"usd_market_cap"`
	Swaps1h                  int         `json:"swaps_1h"`
	Volume1h                 string      `json:"volume_1h"`
	Buys1h                   int         `json:"buys_1h"`
	Sells1h                  int         `json:"sells_1h"`
	BotDegenCount            string      `json:"bot_degen_count"`
}

// KlineResponse 表示K线数据的响应结构
type KlineResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data []KlineData `json:"data"`
}

// KlineData 表示单个K线数据点
type KlineData struct {
	Open   string `json:"open"`
	Close  string `json:"close"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Time   string `json:"time"`
	Volume string `json:"volume"`
}
