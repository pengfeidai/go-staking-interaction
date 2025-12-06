package dto

import "math/big"

type ERCRequest struct {
	ToAddress string   `json:"toAddress"`
	Amount    *big.Int `json:"amount"`
}
type ERCRes struct {
	Hash        string   `json:"hash"`
	Symbol      string   `json:"symbol"`
	Decimal     uint8    `json:"decimal"`
	BlockNumber *big.Int `json:"blockNumber"`
}
