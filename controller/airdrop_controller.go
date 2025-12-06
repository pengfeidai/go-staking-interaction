package controller

import (
	"github.com/gin-gonic/gin"
	"log"
	"math/big"
	"net/http"
	"staking-interaction/adapter"
	log2 "staking-interaction/common/logger"
	"staking-interaction/dto"
	"staking-interaction/service"
	"staking-interaction/utils"
)

func GenerateMultiWallets(c *gin.Context) {
	var request dto.AirdropRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(500, gin.H{"msg": "request body invalid", "err": err})
	}
	wallets, err := service.GetMultiWallets(request.Count)
	if err != nil {
		c.JSON(500, gin.H{"msg": "generate wallet failed", "err": err})
	}
	c.JSON(200, gin.H{"msg": "generate success!", "data": wallets})

}

func AirdropERC20(c *gin.Context) {
	logger := log2.GetLogger()
	logger.WithFields(map[string]interface{}{
		"module": "controller/airdroperc",
	})
	var request dto.AirdropRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"msg": "request body invalid", "err": err})
	}

	reqCount := request.Count
	reqBatchSize := request.BatchSize
	reqAmount := request.Amount

	if reqCount <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "count 必须大于 0"})
		return
	}
	if reqBatchSize <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "batchSize 必须大于 0"})
		return
	}
	if reqAmount == nil || reqAmount.Cmp(big.NewInt(0)) <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "amount 必须大于 0"})
		return
	}

	client, err := adapter.NewInitEthClient()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "client init failed", "err": err})
		return
	}
	amountArray, err := utils.GenerateRandomAmount(reqCount, reqAmount)
	if err != nil {
		log.Fatalf("failed to generate random amounts: %v", err)
	}

	airdropService := service.NewAirdropService(client, logger)
	responses, err := airdropService.AirdropERC20(reqCount, reqBatchSize, amountArray)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "Airdrop failed", "err": err.Error()})
		return
	}
	c.JSON(http.StatusOK, responses)
}

func AirdropBNB(c *gin.Context) {
	logger := log2.GetLogger()
	logger.WithFields(map[string]interface{}{
		"module": "controller/airdropbnb",
	})
	var (
		request dto.AirdropRequest
	)

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"msg": "request body invalid", "err": err})
	}

	reqCount := request.Count
	reqBatchSize := request.BatchSize
	reqAmount := request.Amount

	if reqCount <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "count 必须大于 0"})
		return
	}
	if reqBatchSize <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "batchSize 必须大于 0"})
		return
	}
	if reqAmount == nil || reqAmount.Cmp(big.NewInt(0)) <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "amount 必须大于 0"})
		return
	}
	amountArray, err := utils.GenerateRandomAmount(reqCount, reqAmount)
	if err != nil {
		log.Fatalf("failed to generate random amounts: %v", err)
	}

	client, err := adapter.NewInitEthClient()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "client init failed", "err": err})
		return
	}
	airdropService := service.NewAirdropService(client, logger)
	responses, err := airdropService.AirdropBNB(reqCount, reqBatchSize, amountArray)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "Airdrop failed", "err": err})
		return
	}
	c.JSON(http.StatusOK, responses)

}
