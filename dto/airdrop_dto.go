package dto

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type Wallet struct {
	PrivateKey string // 私钥（十六进制）
	Address    string // 以太坊地址（0x前缀）
}

type AirdropRequest struct {
	Count     int      `json:"count"`
	BatchSize int      `json:"batchSize"`
	Amount    *big.Int `json:"amount"`
}

type AirdropInfo struct {
	BatchNum        int              `json:"batchNum"`
	Hash            string           `json:"hash"`
	ContractAddress string           `json:"contractAddress"`
	FromAddress     common.Address   `json:"fromAddress"`
	WalletAddress   []common.Address `json:"walletAddress"`
	Error           string           `json:"error"`
}

type AirdropResponse struct {
	Msg              string        `json:"msg"`
	CompletedBatches int           `json:"completedBatches"`
	SuccessBatches   int           `json:"successBatches"`
	FailBatches      int           `json:"failBatches"`
	Data             []AirdropInfo `json:"data"`
	Error            string        `json:"error"`
}
