package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"staking-interaction/adapter"
	"staking-interaction/dto"
	"staking-interaction/service"
)

func Stake(c *gin.Context) {
	var request dto.StakeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "request body invalid", "error": err.Error()})
		return
	}
	client, err := adapter.NewInitEthClient()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "client init failed", "err": err})
		return
	}
	stakeService := service.NewStakeService(client)
	response, err := stakeService.Stake(request.Amount, request.Period)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "stake transaction error", "error": err})
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Stake success!", "data": response})
}

func Withdraw(c *gin.Context) {
	var request dto.WithDrawnRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "request body invalid", "error": err.Error()})
		return
	}
	client, err := adapter.NewInitEthClient()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"msg": "client init failed", "err": err})
		return
	}
	stakeService := service.NewStakeService(client)
	response, err := stakeService.Withdraw(&request.Index)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"msg": "withdrawn transaction error", "error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Withdrawn success!", "data": response})
}

func GetAllStakesByFromAddress(c *gin.Context) {
	id := c.Param("id")
	res := service.GetAllStakesByFromAddress(id)
	c.JSON(http.StatusOK, gin.H{"msg": "success", "data": res})
}
