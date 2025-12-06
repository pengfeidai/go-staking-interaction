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
		"module": "cmd/transactionerc20",
	})

	toAddressFlag := flag.String("toAddress", "", "ERC20收款地址")
	amountFlag := flag.String("amount", "", "ERC20数量（最小单位，如wei）")
	flag.Parse()

	amount, e := utils.StringToBigInt(*amountFlag)
	if e != nil {
		log.WithFields(map[string]interface{}{
			"action": "parse_amount",
			"param":  "amount",
			"value":  *amountFlag,
			"detail": e.Error(),
		}).Error("Invalid amount format")
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

	transactionService := service.NewTransactionService(clientInfo)
	log.WithFields(map[string]interface{}{
		"action":  "init_service",
		"service": "TransactionService",
		"detail":  "TransactionService initialized",
	}).Info("TransactionService initialized")

	response, err := transactionService.SendErc20(*toAddressFlag, amount)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "send_erc20",
			"error_code": "SEND_ERC20_FAIL",
			"to_address": *toAddressFlag,
			"amount":     amount.String(),
			"detail":     err.Error(),
		}).Error("Transaction ERC20 failed")
		return
	}

	log.WithFields(map[string]interface{}{
		"action":     "send_erc20",
		"result":     "success",
		"to_address": *toAddressFlag,
		"amount":     amount.String(),
		"tx_hash":    response.Hash,
		"symbol":     response.Symbol,
		"decimal":    response.Decimal,
	}).Info("Transaction ERC20 succeeded")
}
