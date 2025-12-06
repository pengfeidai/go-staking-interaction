package repository

import (
	"log"
	"staking-interaction/adapter"
	"staking-interaction/model"
)

func AddStakeInfo(stake model.Stake) {
	adapter.DB.Create(&stake)

	log.Printf("Add " + stake.Method + " info to database success, indexId: " + stake.IndexNum)
}

func GetAllStakesByFromAddress(fromAddress string) model.Stake {
	var stake model.Stake
	adapter.DB.Where("from_address = ?", fromAddress).First(&stake)
	return stake
}
