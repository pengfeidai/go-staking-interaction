package main

import (
	"flag"
	"staking-interaction/adapter"
	"staking-interaction/common/logger"
	"staking-interaction/service"
	"staking-interaction/utils"
)

func main() {
	log := logger.GetLogger().WithFields(map[string]interface{}{
		"module": "cmd/withdraw",
	})

	indexFlag := flag.String("index", "", "质押索引（最小单位，如bigint）")
	flag.Parse()

	index, e := utils.StringToBigInt(*indexFlag)
	if e != nil {
		log.WithFields(map[string]interface{}{
			"action": "parse_index",
			"param":  "index",
			"value":  *indexFlag,
			"detail": e.Error(),
		}).Error("Invalid index format")
		return
	}

	err := adapter.MysqlConn()
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "init_db",
			"error_code": "DB_CONN_FAIL",
			"detail":     err.Error(),
		}).Error("MySQL database connect failed")
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
		}).Error("Init client failed")
		return
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

	response, err := stakeService.Withdraw(index)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "withdraw",
			"error_code": "WITHDRAW_FAIL",
			"index":      index.String(),
			"detail":     err.Error(),
		}).Error("Withdraw failed")
		return
	}

	log.WithFields(map[string]interface{}{
		"action":        "withdraw",
		"result":        "success",
		"index":         index.String(),
		"tx_hash":       response.Hash,
		"contract_addr": response.ContractAddress,
		"from_addr":     response.FromAddress.Hex(),
		"method":        response.Method,
	}).Info("Withdraw succeeded")
}
