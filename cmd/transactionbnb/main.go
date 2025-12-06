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
		"module": "cmd/transactionbnb",
	})

	toAddressFlag := flag.String("toAddress", "", "BNB收款地址")
	amountFlag := flag.String("amount", "", "BNB数量（最小单位，如wei）")
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

	defer func() {
		err := adapter.CloseConn()
		if err != nil {
			log.WithFields(map[string]interface{}{
				"action":     "close_db",
				"error_code": "DB_CLOSE_FAIL",
				"detail":     err.Error(),
			}).Error("Close database failed")
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
	defer clientInfo.CloseEthClient()

	transactionService := service.NewTransactionService(clientInfo)
	response, err := transactionService.SendBNB(*toAddressFlag, amount)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "send_bnb",
			"error_code": "SEND_BNB_FAIL",
			"to_address": *toAddressFlag,
			"amount":     amount.String(),
			"detail":     err.Error(),
		}).Error("Transaction BNB failed")
		return
	}

	log.WithFields(map[string]interface{}{
		"action":     "send_bnb",
		"result":     "success",
		"to_address": *toAddressFlag,
		"amount":     amount.String(),
		"tx_hash":    response.Hash,
		"symbol":     response.Symbol,
	}).Info("Transaction BNB succeeded")
}
