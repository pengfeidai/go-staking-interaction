package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"staking-interaction/controller"
	"staking-interaction/middleware"
)

func InitRouter(redis *redis.Client) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	group := router.Group("/v1")

	staking := group.Group("/staking")
	{
		staking.POST("/stake", controller.Stake)
		staking.POST("/withdraw", controller.Withdraw)
		staking.GET("stake/:address", controller.GetAllStakesByFromAddress)
	}

	airdrop := group.Group("/airdropping")
	{
		airdrop.POST("/airdroperc20", controller.AirdropERC20)
		airdrop.POST("/airdropbnb", controller.AirdropBNB)
		airdrop.POST("/generateWallet", controller.GenerateMultiWallets)
	}

	authMid := middleware.NewAuthMiddleware(redis)
	transfer := group.Group("/transfer")
	transfer.Use(authMid.AuthMiddleware())
	{
		transfer.POST("/transferERC20", controller.SendErc20)
		transfer.POST("/transferBNB", controller.SendBNB)
	}

	auth := group.Group("/login")
	{
		auth.POST("/bsc", func(c *gin.Context) {
			controller.LoginBSC(c, redis)
		})
		auth.POST("/solana", func(c *gin.Context) {
			controller.LoginSolana(c, redis)
		})
	}

	return router
}
