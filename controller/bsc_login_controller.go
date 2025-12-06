package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"net/http"
	"staking-interaction/dto"
	"staking-interaction/service"
)

func LoginBSC(c *gin.Context, redis *redis.Client) {
	var req dto.LoginBSCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "request body invalid", "error": err.Error()})
		return
	}
	bscService := service.NewAuthBSCService(redis)
	token, err := bscService.Login(req.Signature, req.WalletAddress, req.Nonce, req.Timestamp)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "login failed", "error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
