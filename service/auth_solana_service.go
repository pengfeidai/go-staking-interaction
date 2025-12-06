package service

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/mr-tron/base58"
	"github.com/sirupsen/logrus"
	"os"
	"staking-interaction/common/config"
	"staking-interaction/common/logger"
	"staking-interaction/dto"
	"staking-interaction/repository"
	"time"
)

type AuthSolanaService struct {
	redis  *redis.Client
	log    *logrus.Logger
	config *config.Config
}

func NewAuthSolanaService(redis *redis.Client) *AuthSolanaService {
	conf := config.Get()
	log := logger.GetLogger()
	log.WithFields(logrus.Fields{
		"module": "solana_auth_service",      // 主模块名
		"env":    conf.AppConfig.Environment, // 环境
		"pid":    os.Getpid(),                // 进程号
	})

	return &AuthSolanaService{
		redis:  redis,
		log:    log,
		config: conf,
	}
}

func (a *AuthSolanaService) GenerateSolanaMessage(nonce string, timestamp int64) string {
	return fmt.Sprintf("SOLANA LOGIN: %s - %d", nonce, timestamp)
}

func (a *AuthSolanaService) Login(signature string, address string, nonce string, timestamp int64) (string, error) {
	if a.isNonceUsed(nonce) {
		return "", fmt.Errorf("nonce is already used, nonce:%s", nonce)
	}
	if !a.isTimestampValid(timestamp) {
		return "", fmt.Errorf("timestamp is invalid, timestamp:%d, now:%d, delta:%d", timestamp, time.Now().Unix(), a.config.AuthConfig.MaxDelta)
	}
	msg := a.GenerateSolanaMessage(nonce, timestamp)

	isValid, err := a.verifySolanaToken(msg, signature, address)
	if err != nil {
		return "", fmt.Errorf("verify solana token failed, signature:%s, msg:%s, address:%s, err:%s", signature, msg, address, err)
	}
	if !isValid {
		return "", fmt.Errorf("invalid signature, signature:%s, msg:%s, address:%s", signature, msg, address)
	}

	account, err := repository.GetAccount(address)
	if err != nil || account.AccountID == 0 {
		return "", fmt.Errorf("invalid account, error: %v, address:%s", err, address)
	}

	publicKey, jwtToken, err := a.generateJWTToken(address)
	config.SetEd25519PublicKey(a.config, publicKey)

	if err != nil {
		return "", fmt.Errorf("generate jwtToken error: %v", err)
	}

	if err := a.storeTokenToRedis(jwtToken, address); err != nil {
		return "", fmt.Errorf("store token to redis error: %v", err)
	}

	a.storeNonceToRedis(nonce)
	return jwtToken, nil
}

func (a *AuthSolanaService) verifySolanaToken(msg string, signatureB58 string, addressB58 string) (bool, error) {
	// 解码公钥（base58转[]byte），
	addressBytes, err := base58.Decode(addressB58)
	if err != nil {
		return false, fmt.Errorf("decode public key failed, error: %v", err)
	}
	// 解码签名（base58转[]byte）
	sigBytes, err := base58.Decode(signatureB58)
	if err != nil {
		return false, fmt.Errorf("decode signature failed, error: %v", err)
	}
	// ed25519验签
	return ed25519.Verify(addressBytes, []byte(msg), sigBytes), nil
}

func (a *AuthSolanaService) generateJWTToken(walletAddress string) (*ed25519.PublicKey, string, error) {
	// 生成的ed25519密钥
	seedBytes, err := hex.DecodeString(a.config.AuthConfig.Ed25519Seed)
	if err != nil {
		return nil, "", fmt.Errorf("decode seed failed, error: %v", err)
	}
	private := ed25519.NewKeyFromSeed(seedBytes)      // 64字节私钥
	publicKey := private.Public().(ed25519.PublicKey) // 直接获取公钥

	claims := dto.CustomClaims{
		WalletAddress: walletAddress,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.config.AuthConfig.JwtExpiration)), //过期时间（当前时间+24小时）
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                        //签发时间（当前时间）
		},
	}
	// 创建一个使用 EdDSA 算法签名的 JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	// 用配置中的密钥进行签名，
	jwtToken, err := token.SignedString(private)
	//返回token字符串
	return &publicKey, jwtToken, err
}

func (a *AuthSolanaService) storeTokenToRedis(token string, address string) error {
	key := fmt.Sprintf("token_solana:%s", address)
	a.redis.Set(context.Background(), key, token, a.config.AuthConfig.JwtExpiration)
	return nil
}

func (a *AuthSolanaService) isNonceUsed(nonce string) bool {
	key := fmt.Sprintf("nonce:%s", nonce)
	exist, _ := a.redis.Exists(context.Background(), key).Result()
	return exist > 0
}

func (a *AuthSolanaService) storeNonceToRedis(nonce string) {
	key := fmt.Sprintf("nonce:%s", nonce)
	a.redis.Set(context.Background(), key, 1, a.config.AuthConfig.JwtExpiration)
}

// isTimestampValid 允许的时间误差窗口, 防止因为客户端和服务端之间时间不同步造成误判
func (a *AuthSolanaService) isTimestampValid(ts int64) bool {
	now := time.Now().Unix()
	delta := int64(a.config.AuthConfig.MaxDelta * 60)
	return ts <= now+delta && ts >= now-delta
}
