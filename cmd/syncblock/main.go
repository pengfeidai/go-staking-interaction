package main

import (
	"os"
	"os/signal"
	"staking-interaction/adapter"
	"staking-interaction/common/config"
	"staking-interaction/common/logger"
	redisClient "staking-interaction/common/redis"
	"staking-interaction/listener"
	"staking-interaction/service"
	"syscall"
)

func main() {
	log := logger.GetLogger()
	log.WithFields(map[string]interface{}{
		"module": "cmd/syncblock",
	})
	conf := config.Get()

	// 1. 初始化数据库
	err := adapter.MysqlConn()
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "init_db",
			"error_code": "DB_CONN_FAIL",
			"detail":     err.Error(),
		}).Fatal("MySQL database connect failed")
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

	// 2. 初始化 Redis 连接
	redis, err := adapter.NewRedisClientWithRetry()
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "init_redis",
			"error_code": "REDIS_CONN_FAIL",
			"detail":     err.Error(),
		}).Fatal("Redis connection failed")
	}
	defer redis.Close()

	// 3. 创建 LockManager 实例
	lockManager := redisClient.NewLockManager(redis)
	log.WithFields(map[string]interface{}{
		"action": "init_lock_manager",
		"detail": "LockManager initialized successfully",
	}).Info("LockManager initialized")

	// 4. 初始化区块链客户端
	clientInfo, err := adapter.NewSyncEthClient()
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "init_client",
			"error_code": "CLIENT_INIT_FAIL",
			"detail":     err.Error(),
		}).Fatal("Init client failed")
	}
	defer clientInfo.CloseSyncEthClient()

	// 5. syncBlock
	log.WithFields(map[string]interface{}{
		"action": "init_sync_block",
		"detail": "Initializing sync block",
	}).Info("Initializing sync block...")
	syncBlock := listener.NewSyncBlockInfo(clientInfo, conf.BlockchainConfig, lockManager, log)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.WithFields(map[string]interface{}{
					"action": "sync_block_panic",
					"detail": r,
				}).Error("SyncBlockInfo panic")
			}
		}()
		log.WithFields(map[string]interface{}{
			"action": "sync_block_start",
			"detail": "SyncBlockInfo goroutine started",
		}).Info("SyncBlockInfo goroutine started")
		syncBlock.Start()
	}()
	defer func() {
		log.WithFields(map[string]interface{}{
			"action": "sync_block_stop",
			"detail": "Stopping SyncBlockInfo",
		}).Info("Stopping SyncBlockInfo...")
		syncBlock.Stop()
	}()

	// 6. withdraw handler
	log.WithFields(map[string]interface{}{
		"action": "init_withdraw_handler",
		"detail": "Initializing withdraw handler",
	}).Info("Initializing withdraw handler...")
	txService := service.NewTransactionService(clientInfo)
	withdrawHd := listener.NewWithdrawHandler(txService, lockManager, log)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.WithFields(map[string]interface{}{
					"action": "withdraw_handler_panic",
					"detail": r,
				}).Error("WithdrawHandler panic")
			}
		}()
		log.WithFields(map[string]interface{}{
			"action": "withdraw_handler_start",
			"detail": "WithdrawHandler goroutine started",
		}).Info("WithdrawHandler goroutine started")
		withdrawHd.Start()
	}()
	defer func() {
		log.WithFields(map[string]interface{}{
			"action": "withdraw_handler_stop",
			"detail": "Stopping WithdrawHandler",
		}).Info("Stopping WithdrawHandler...")
		withdrawHd.Stop()
	}()

	// 7. sync withdraw
	log.WithFields(map[string]interface{}{
		"action": "init_sync_withdraw",
		"detail": "Initializing sync withdraw",
	}).Info("Initializing sync withdraw...")
	syncWithdraw := listener.NewSyncWithdrawHandler(clientInfo.Client, lockManager, log)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.WithFields(map[string]interface{}{
					"action": "sync_withdraw_panic",
					"detail": r,
				}).Error("SyncWithdrawHandler panic")
			}
		}()
		log.WithFields(map[string]interface{}{
			"action": "sync_withdraw_start",
			"detail": "SyncWithdrawHandler goroutine started",
		}).Info("SyncWithdrawHandler goroutine started")
		syncWithdraw.Start()
	}()
	defer func() {
		log.WithFields(map[string]interface{}{
			"action": "sync_withdraw_stop",
			"detail": "Stopping SyncWithdrawHandler",
		}).Info("Stopping SyncWithdrawHandler...")
		syncWithdraw.Stop()
	}()

	log.WithFields(map[string]interface{}{
		"action": "service_initialized",
		"detail": "SyncWithdrawHandler launched, all services initialized",
	}).Info("Services initialized")

	// 8. 等待关闭信号
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.WithFields(map[string]interface{}{
		"action": "wait_signal",
		"detail": "Waiting for shutdown signal",
	}).Info("Waiting for shutdown signal...")
	<-signalChan
	log.WithFields(map[string]interface{}{
		"action": "shutdown",
		"detail": "Shutdown signal received, exiting",
	}).Info("Shutdown signal received, exiting...")
}
