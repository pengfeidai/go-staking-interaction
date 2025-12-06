package model

import "time"

// TransactionLog 对应 transaction_log 表
type TransactionLog struct {
	LogID       uint64    `gorm:"column:log_id;type:bigint unsigned;primary_key;AUTO_INCREMENT" json:"log_id"`
	AccountID   int       `gorm:"column:account_id;type:int" json:"account_id"`
	TokenType   int       `gorm:"column:token_type;type:tinyint" json:"token_type"`      // 1.BNB 2.ROM 3.WCN
	Hash        string    `gorm:"column:hash;uniqueIndex;type:varchar(228)" json:"hash"` //唯一索引怕重复
	Amount      string    `gorm:"column:amount;type:varchar(64)" json:"amount"`
	FromAddress string    `gorm:"column:from_address;type:varchar(228)" json:"from_address"`
	ToAddress   string    `gorm:"column:to_address;type:varchar(228)" json:"to_address"`
	BlockNumber string    `gorm:"column:block_number;varchar(64)" json:"block_number"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;" comment:"记录创建时间"`
}
