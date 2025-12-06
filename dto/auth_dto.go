package dto

import (
	"github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	WalletAddress string `json:"wallet_address"`
	jwt.RegisteredClaims
}
