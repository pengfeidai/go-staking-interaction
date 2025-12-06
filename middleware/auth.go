package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"staking-interaction/common/config"
	"staking-interaction/dto"
	"strings"
)

type Auth struct {
	redis  *redis.Client
	config *config.Config
}

func NewAuthMiddleware(client *redis.Client) *Auth {
	conf := config.Get()
	return &Auth{
		redis:  client,
		config: conf,
	}
}

func (a *Auth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := a.extractToken(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "extract token failed", "error": err.Error()})
			return
		}
		//校验 JWT
		address, err := a.VerifyJWTToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "invalid or expired token", "error": err.Error()})
			return
		}

		//从 Redis 查找 Token 是否存在
		key := fmt.Sprintf("token_bsc:%s", address)
		exists, err := a.redis.Exists(context.Background(), key).Result()
		if err != nil || exists == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "token is not existed", "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

func (a *Auth) extractToken(req *http.Request) (string, error) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("auth header is empty")
	}

	if !strings.HasPrefix(authHeader, a.config.AuthConfig.Prefix) {
		return "", fmt.Errorf("auth header format is invalid")
	}
	return strings.TrimPrefix(authHeader, a.config.AuthConfig.Prefix), nil
}

func (a *Auth) VerifyJWTToken(tokenStr string) (string, error) {
	claims := &dto.CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.config.AuthConfig.EcdsaPublicKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", fmt.Errorf("token is expired, please login, ExpiresAt:%v", claims.ExpiresAt)
		}
		return "", fmt.Errorf("token parse failed: %v", err)
	}
	if !token.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	return claims.WalletAddress, nil
}

func (a *Auth) VerifyJWTTokens(tokenStr string) (string, error) {
	claims := &dto.CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		switch token.Method.(type) {
		case *jwt.SigningMethodECDSA:
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return a.config.AuthConfig.EcdsaPublicKey, nil
		case *jwt.SigningMethodEd25519:
			if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return *a.config.AuthConfig.Ed25519PublicKey, nil
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", fmt.Errorf("token is expired, please login, ExpiresAt:%v", claims.ExpiresAt)
		}
		return "", fmt.Errorf("token parse failed: %v", err)
	}
	if !token.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	return claims.WalletAddress, nil
}

//func (a *AuthBSCService) Verify(token string, accountId int) bool {
//	//从 Redis 查找 Token 是否存在
//	key := fmt.Sprintf("token_bsc:%d", accountId)
//	exists, err := a.redis.Exists(context.Background(), key).Result()
//	if err != nil || exists == 0 {
//		fmt.Println("token is not existed", "error:", err)
//		return false
//	}
//	//校验 JWT
//	if err := a.VerifyJWTToken(token); err != nil {
//		fmt.Println("invalid or expired token", "error:", err)
//		return false
//	}
//	return true
//}

// ======
//
// func (a *AuthSolanaService) VerifyJWTToken(tokenStr string) error {
//
//		claims := &dto.CustomClaims{}
//
//		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
//			if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
//				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
//			}
//			return *a.config.AuthConfig.Ed25519PublicKey, nil
//		})
//		if err != nil {
//			if errors.Is(err, jwt.ErrTokenExpired) {
//				return fmt.Errorf("token is expired, please login, ExpiresAt:%v", claims.ExpiresAt)
//			}
//			return fmt.Errorf("token parse failed: %v", err)
//		}
//		if !token.Valid {
//			return fmt.Errorf("token is not valid")
//		}
//
//		return nil
//	}
//func (a *Auth) Verify(token string, accountId int) bool {
//	////从 Redis 查找 Token 是否存在
//	//key := fmt.Sprintf("token_solana:%d", accountId)
//	//exists, err := a.redis.Exists(context.Background(), key).Result()
//	//if err != nil || exists == 0 {
//	//	fmt.Println("token is not existed", "error:", err)
//	//	return false
//	//}
//	////校验 JWT
//	//if err := a.VerifyJWTToken(token); err != nil {
//	//	fmt.Println("invalid or expired token", "error:", err)
//	//	return false
//	//}
//	//return true
//}
