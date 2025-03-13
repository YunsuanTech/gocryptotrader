package account

import (
	"time"
)

type Account struct {
	ID                int    `gorm:"id"`
	Name              string `gorm:"name"`
	Address           string `gorm:"address"`
	ExchangeAddressID string `gorm:"exchange_address_id"`
	ZkAddressID       string `gorm:"zk_address_id"`
	F4AddressID       string `gorm:"f4_address_id"`
	OTAddressID       string `gorm:"ot_address_id"`
	Cipher            string `gorm:"cipher"`
	Layer             int    `gorm:"layer"`
	Owner             string `gorm:"owner"`
	ChainName         string `gorm:"chain_name"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
