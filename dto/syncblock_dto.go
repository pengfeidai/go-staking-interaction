package dto

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type TransferEvent struct {
	FromAddress common.Address `json:"from_address"`
	ToAddress   common.Address `json:"to_address"`
	Value       *big.Int       `json:"value"`
}
