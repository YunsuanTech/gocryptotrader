package fill

import (
	"time"

	"gocryptotrader/currency"
	"gocryptotrader/exchanges/asset"
	"gocryptotrader/exchanges/order"
)

// Fills is used to hold data and methods related to fill dissemination
type Fills struct {
	dataHandler      chan interface{}
	fillsFeedEnabled bool
}

// Data defines fill data
type Data struct {
	ID            string
	Timestamp     time.Time
	Exchange      string
	AssetType     asset.Item
	CurrencyPair  currency.Pair
	Side          order.Side
	OrderID       string
	ClientOrderID string
	TradeID       string
	Price         float64
	Amount        float64
}
