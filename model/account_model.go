package model

import "time"

// Account 对应 account 表
type Account struct {
	AccountID     int    `gorm:"column:account_id;type:int;primary_key;AUTO_INCREMENT" json:"account_id"`
	AccountName   string `gorm:"column:account_name;type:varchar(128);not null" json:"account_name"`
	Email         string `gorm:"column:email;type:varchar(64);unique_index:email_index" json:"email"`
	WalletAddress string `gorm:"column:wallet_address;type:varchar(64)" json:"wallet_address"`
}

// AccountAsset 对应 account_asset 表
type AccountAsset struct {
	AssetID    int       `gorm:"column:asset_id;type:int;primary_key;AUTO_INCREMENT" json:"asset_id"`
	AccountID  int       `gorm:"column:account_id;type:int;not null" json:"account_id"`
	BnbBalance string    `gorm:"column:bnb_balance;type:varchar(64);default:''" json:"bnb_balance"`
	MtkBalance string    `gorm:"column:mtk_balance;type:varchar(64);default:''" json:"mtk_balance"`
	Version    int       `gorm:"default:0" json:"version"`
	Account    Account   `gorm:"foreignKey:AccountID;references:AccountID"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
