package service

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"os"
	"staking-interaction/common/config"
	"staking-interaction/common/logger"
	"staking-interaction/dto"
	"staking-interaction/repository"
	"strings"
	"time"
)

type AuthBSCService struct {
	redis  *redis.Client
	log    *logrus.Logger
	config *config.Config
}

func NewAuthBSCService(redis *redis.Client) *AuthBSCService {
	conf := config.Get()
	log := logger.GetLogger()
	log.WithFields(logrus.Fields{
		"module": "bsc_auth_service",         // 主模块名
		"env":    conf.AppConfig.Environment, // 环境
		"pid":    os.Getpid(),                // 进程号
	})

	return &AuthBSCService{
		redis:  redis,
		log:    log,
		config: conf,
	}
}

func (a *AuthBSCService) GenerateChallengeMessage(nonce string, timestamp int64) string {
	return fmt.Sprintf(
		"Welcome to DApp! Please sign this message to login.\nNonce: %s\nTimestamp: %d\n",
		nonce, timestamp)
}

func (a *AuthBSCService) Login(signature string, address string, nonce string, timestamp int64) (string, error) {
	// 判断nonce是否重复
	if a.isNonceUsed(nonce) {
		return "", fmt.Errorf("nonce is already used, nonce:%s", nonce)
	}

	// 判断timestamp是否有效
	if !a.isTimestampValid(timestamp) {
		return "", fmt.Errorf("timestamp is invalid, timestamp:%d, now:%d, delta:%d", timestamp, time.Now().Unix(), a.config.AuthConfig.MaxDelta)
	}

	// 生成挑战消息
	msg := a.GenerateChallengeMessage(nonce, timestamp)

	//确认BSC token
	isValid, err := a.verifyBSCToken(signature, msg, address)
	if err != nil || !isValid {
		return "", fmt.Errorf("invalid signature, error: %v", err)
	}

	account, err := repository.GetAccount(address)
	if err != nil || account.AccountID == 0 {
		return "", fmt.Errorf("invalid account, error: %v", err)
	}

	// 生成publicKey, JWT token
	publicKey, jwtToken, err := a.generateJWTToken(account.WalletAddress)
	if err != nil {
		return "", fmt.Errorf("generate jwtToken error: %v", err)
	}

	// 保存publicKey到config里
	config.SetEcdsaPublicKey(a.config, publicKey)

	// 保存jwtToken到redis里
	if err := a.storeTokenToRedis(jwtToken, account.WalletAddress); err != nil {
		return "", fmt.Errorf("store token to redis error: %v", err)
	}

	// 更新nonce到redis里
	a.storeNonceToRedis(nonce)

	return jwtToken, nil
}

func (a *AuthBSCService) verifyBSCToken(signature string, msg string, address string) (bool, error) {
	recoveredAddress, err := a.getVerifyAddress(signature, msg)
	if err != nil {
		return false, fmt.Errorf("get recovered address failed: %v", err)
	}
	return recoveredAddress == common.HexToAddress(address), nil
}

func (a *AuthBSCService) getVerifyAddress(signature string, msg string) (common.Address, error) {
	// 把要签名的消息，包装成以太坊标准格式（EIP-191），再用 Keccak256 哈希。
	msgHash := accounts.TextHash([]byte(msg))

	// 签名解码成原始字节格式
	sigHex := strings.TrimPrefix(signature, "0x")
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return common.Address{}, fmt.Errorf("decode signature error: %v", err)
	}

	// 恢复公钥
	// 以太坊签名采用 ECDSA，可以通过签名和消息哈希反向推导出公钥（并不是所有椭圆曲线算法都能这样，但 SECP256k1 可以）
	publicKeyBytes, err := crypto.Ecrecover(msgHash, sigBytes)
	if err != nil {
		return common.Address{}, fmt.Errorf("VerifyBSCToken: failed to recover public key: %v", err)
	}

	// 转ecdsa公钥
	publicKey, err := crypto.UnmarshalPubkey(publicKeyBytes)
	if err != nil {
		return common.Address{}, fmt.Errorf("VerifyBSCToken: failed to unmarshal public key: %v", err)
	}

	// 公钥转地址
	addr := crypto.PubkeyToAddress(*publicKey)
	return addr, nil
}

func (a *AuthBSCService) generateJWTToken(walletAddress string) (*ecdsa.PublicKey, string, error) {
	// 生成密钥对
	private, err := crypto.HexToECDSA(a.config.AuthConfig.EcdsaPrivateKey)
	if err != nil {
		return nil, "", fmt.Errorf("HexToECDSA error: %v", err)
	}

	// 构造claims
	claims := dto.CustomClaims{
		WalletAddress: walletAddress,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.config.AuthConfig.JwtExpiration)), //过期时间（当前时间+24小时）
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                        //签发时间（当前时间）
		},
	}
	// 用 ES256 算法（ECDSA-P256）创建 token
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	// 用私钥签名，得到 token 字符串
	jwtToken, err := token.SignedString(private)

	return &private.PublicKey, jwtToken, err
}

func (a *AuthBSCService) storeTokenToRedis(token string, address string) error {
	key := fmt.Sprintf("token_bsc:%s", address)
	a.redis.Set(context.Background(), key, token, a.config.AuthConfig.JwtExpiration)
	return nil
}

func (a *AuthBSCService) isNonceUsed(nonce string) bool {
	key := fmt.Sprintf("nonce:%s", nonce)
	exist, _ := a.redis.Exists(context.Background(), key).Result()
	return exist > 0
}

func (a *AuthBSCService) storeNonceToRedis(nonce string) {
	key := fmt.Sprintf("nonce:%s", nonce)
	a.redis.Set(context.Background(), key, 1, a.config.AuthConfig.JwtExpiration)
}

// isTimestampValid 允许的时间误差窗口, 防止因为客户端和服务端之间时间不同步造成误判
func (a *AuthBSCService) isTimestampValid(ts int64) bool {
	now := time.Now().Unix()
	delta := int64(a.config.MaxDelta * 60)
	return ts <= now+delta && ts >= now-delta
}
