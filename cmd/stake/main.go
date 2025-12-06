package main

import (
	"flag"
	"staking-interaction/adapter"
	"staking-interaction/common/logger"
	"staking-interaction/service"
)

func main() {
	log := logger.GetLogger().WithFields(map[string]interface{}{
		"module": "cmd/stake",
	})

	amountFlag := flag.Int64("amount", 0, "质押金额（最小单位，如wei）")
	periodFlag := flag.Uint("period", 0, "质押周期（天）")
	flag.Parse()

	// 验证必填参数
	if *amountFlag <= 0 {
		log.WithFields(map[string]interface{}{
			"action": "validate_input",
			"param":  "amount",
			"value":  *amountFlag,
			"detail": "amount should be greater than 0",
		}).Fatal("Invalid argument: -amount")
	}
	if *periodFlag < 0 {
		log.WithFields(map[string]interface{}{
			"action": "validate_input",
			"param":  "period",
			"value":  *periodFlag,
			"detail": "period is invalid",
		}).Fatal("Invalid argument: -period")
	}

	err := adapter.MysqlConn()
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "init_db",
			"error_code": "DB_CONN_FAIL",
			"detail":     err.Error(),
		}).Fatal("MySQL database connect failed")
		return
	}
	log.WithFields(map[string]interface{}{
		"action": "init_db",
		"detail": "MySQL database connected",
	}).Info("Database connected")

	defer func() {
		err := adapter.CloseConn()
		if err != nil {
			log.WithFields(map[string]interface{}{
				"action":     "close_db",
				"error_code": "DB_CLOSE_FAIL",
				"detail":     err.Error(),
			}).Error("Close database failed")
		} else {
			log.WithFields(map[string]interface{}{
				"action": "close_db",
				"detail": "Database connection closed",
			}).Info("Database connection closed")
		}
	}()

	clientInfo, err := adapter.NewInitEthClient()
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "init_client",
			"error_code": "CLIENT_INIT_FAIL",
			"detail":     err.Error(),
		}).Fatal("Init client failed")
	}
	log.WithFields(map[string]interface{}{
		"action": "init_client",
		"detail": "Client initialized",
	}).Info("Client initialized")
	defer func() {
		clientInfo.CloseEthClient()
		log.WithFields(map[string]interface{}{
			"action": "close_client",
			"detail": "Client closed",
		}).Info("Client closed")
	}()

	stakeService := service.NewStakeService(clientInfo)
	log.WithFields(map[string]interface{}{
		"action":  "init_service",
		"service": "StakeService",
		"detail":  "StakeService initialized",
	}).Info("StakeService initialized")

	response, err := stakeService.Stake(*amountFlag, uint8(*periodFlag))
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "stake",
			"error_code": "STAKE_FAIL",
			"detail":     err.Error(),
			"amount":     *amountFlag,
			"period":     *periodFlag,
		}).Fatal("Stake failed")
	}

	log.WithFields(map[string]interface{}{
		"action":        "stake",
		"result":        "success",
		"amount":        *amountFlag,
		"period":        *periodFlag,
		"tx_hash":       response.Hash,
		"contract_addr": response.ContractAddress,
		"from_addr":     response.FromAddress.Hex(),
		"method":        response.Method,
	}).Info("Stake succeeded")
}
