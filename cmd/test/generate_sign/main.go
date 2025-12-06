package main

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/rand"
	"staking-interaction/common/config"
	"time"
)

// 构造挑战消息
func ChallengeMessage(nonce int, timestamp int64) string {
	return fmt.Sprintf("Welcome to DApp! Please sign this message to login.\nNonce: %d\nTimestamp: %d\n", nonce, timestamp)
}

func main() {
	conf := config.Get()
	key, _ := crypto.GenerateKey()
	fmt.Println("PrivKey hex:", hex.EncodeToString(key.D.Bytes()))
	// 本地测试参数
	privateKeyHex := conf.BlockchainConfig.PrivateKey
	walletAddr := conf.BlockchainConfig.Owners[1]
	nonce := rand.Intn(999999)
	timestamp := time.Now().Unix()

	// 构造消息
	msg := ChallengeMessage(nonce, timestamp)
	fmt.Println("Challenge Message:")
	fmt.Println(msg)

	// 按 EIP-191 进行消息哈希（个人签名标准）
	msgPrefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg)
	msgHash := crypto.Keccak256([]byte(msgPrefix))

	// 用私钥签名
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("私钥格式错误: %v", err)
	}
	signature, err := crypto.Sign(msgHash, privateKey)
	if err != nil {
		log.Fatalf("签名失败: %v", err)
	}

	// V 值处理：crypto.Sign 输出的 V 已经是 0/1，无需处理
	// 如需兼容部分前端钱包，可以将 V 改为 27 或 28
	// signature[64] += 27

	// 输出结果
	fmt.Printf("钱包地址: %s\n", walletAddr)
	fmt.Printf("nonce: %d\n", nonce)
	fmt.Printf("timestamp: %d\n", timestamp)
	fmt.Printf("signature: 0x%s\n", hex.EncodeToString(signature))
}
