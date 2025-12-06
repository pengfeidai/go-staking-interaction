package main

import (
	"crypto/ed25519"
	rander "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/mr-tron/base58"
)

// 测试数据结构
type TestUser struct {
	Nonce     int // base58 公钥
	Wallet    string
	Signature string // base58 签名
	Timestamp int64
}

func main() {
	rand.Seed(time.Now().UnixNano())
	var users []TestUser
	publicKey, privateKey, _ := ed25519.GenerateKey(rander.Reader)

	fmt.Printf("privateKey (len=%d): %s\n", len(privateKey), hex.EncodeToString(privateKey)) // 64字节
	fmt.Printf("publicKey (len=%d): %s\n", len(publicKey), hex.EncodeToString(publicKey))    // 32字节

	for i := 0; i < 5; i++ {
		pub, priv, _ := ed25519.GenerateKey(nil)
		wallet := base58.Encode(pub)
		timestamp := time.Now().Unix()
		nonce := rand.Intn(999999)
		challenge := fmt.Sprintf("SOLANA LOGIN: %d - %d", nonce, timestamp)

		signature := ed25519.Sign(priv, []byte(challenge))
		signatureB58 := base58.Encode(signature)

		users = append(users, TestUser{
			Nonce:     nonce,
			Wallet:    wallet,
			Signature: signatureB58,
			Timestamp: timestamp,
		})
	}

	// 打印测试数据（可存到文件或数据库）
	for _, u := range users {
		fmt.Printf("Nonce: %d\nWallet: %s\nSignature: %s\nTimestamp: %d\n\n", u.Nonce, u.Wallet, u.Signature, u.Timestamp)
	}
}
