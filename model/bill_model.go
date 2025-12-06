package model

import "time"

// Bill 对应 bill 表
type Bill struct {
	ID          uint64    `gorm:"column:id;type:bigint unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	AccountID   int       `gorm:"column:account_id;type:int" json:"account_id"`
	TokenType   int       `gorm:"column:token_type;type:tinyint" json:"token_type"` //  1.BNB 2.MTK
	BillType    int       `gorm:"column:bill_type;type:tinyint" json:"bill_type"`   // 1.充值 2.提现
	Amount      string    `gorm:"column:amount;type:varchar(64);default:'0'" json:"amount"`
	Fee         string    `gorm:"column:fee;type:varchar(64);default:'0'" json:"fee"`
	PreBalance  string    `gorm:"column:pre_balance;type:varchar(64);default:'0'" json:"pre_balance"`
	NextBalance string    `gorm:"column:next_balance;type:varchar(64);default:'0'" json:"next_balance"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;" comment:"记录创建时间"`
}
