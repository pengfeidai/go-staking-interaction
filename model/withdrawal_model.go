package model

import "time"

type Withdrawal struct {
	ID            int       `gorm:"column:id;type:bigint unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	TokenType     int       `gorm:"column:token_type;type:int" json:"token_type"` // 1.BNB 2.MTK
	WalletAddress string    `gorm:"column:wallet_address;type:varchar(64)" json:"wallet_address"`
	Amount        string    `gorm:"column:amount;type:varchar(64)" json:"amount"` // 数量, 10bnb
	Value         string    `gorm:"column:value;type:varchar(64)" json:"value"`   // 实际到账数量, 9bnb
	Fee           string    `gorm:"column:fee;type:varchar(64)" json:"fee"`       // 手续费, 1bnb
	GasPrice      string    `gorm:"column:gas_price;type:varchar(64);default:'0'" json:"gas_price"`
	Status        int8      `gorm:"column:status;type:tinyint" json:"status"` //1.INIT 2.PENDING 3.SUCCESS 4.FAILED  5.PASS(初审通过)-备用 6.REJECT(人工审核驳回)-备用 6.ABNORMAL(提现异常，需人工处理)-备用
	Hash          string    `gorm:"column:hash;type:varchar(120)" json:"hash"`
	Nonce         int       `gorm:"column:nonce;type:int" json:"nonce"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;" comment:"记录创建时间"`
}
