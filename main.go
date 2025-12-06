package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	_ "math/big"
	"net/http"
	"os"
	"os/signal"
	"staking-interaction/adapter"
	"staking-interaction/common/config"
	"staking-interaction/common/logger"
	srouter "staking-interaction/router"
	"syscall"
	"time"
)

func main() {
	conf := config.Get()
	log := logger.GetLogger().WithFields(logrus.Fields{
		"module": "main",                     // 主模块名
		"env":    conf.AppConfig.Environment, // 环境
		"pid":    os.Getpid(),                // 进程号
	})

	err := adapter.MysqlConn()
	if err != nil {
		log.WithFields(logrus.Fields{
			"action":     "init_db",
			"error_code": "DB_CONN_FAIL",
			"detail":     err.Error(),
		}).Fatal("MySQL database connect failed")
		return
	}
	log.WithFields(logrus.Fields{
		"action": "init_db",
	}).Info("MySQL database connected.")

	defer func() {
		err := adapter.CloseConn()
		if err != nil {
			if err != nil {
				log.WithFields(logrus.Fields{
					"action":     "close_db",
					"error_code": "DB_CLOSE_FAIL",
					"detail":     err.Error(),
				}).Error("Close database failed")
			} else {
				log.WithFields(logrus.Fields{
					"action": "close_db",
				}).Info("Database connection closed")
			}
		}
	}()

	redis, err := adapter.NewRedisClientWithRetry()
	if err != nil {
		log.WithFields(map[string]interface{}{
			"action":     "init_redis",
			"error_code": "REDIS_CONN_FAIL",
			"detail":     err.Error(),
		}).Fatal("Redis connection failed")
	}
	defer redis.Close()

	router := srouter.InitRouter(redis)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", conf.AppConfig.Port),
		Handler: router,
	}
	log.WithFields(logrus.Fields{
		"action": "listen",
		"port":   conf.AppConfig.Port,
		"detail": "server ready",
	}).Info("Listening and serving HTTP.")

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting server :", err)
		}
	}()
	clientInfo, err := adapter.NewInitEthClient()
	if err != nil {
		log.WithFields(logrus.Fields{
			"action":     "serve_http",
			"error_code": "HTTP_SERVE_FAIL",
			"detail":     err.Error(),
		}).Fatal("Error starting server")
	}
	defer clientInfo.CloseEthClient()
	//listener.ListenToEvents()

	// 创建系统信号接收器
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Println("shutdown server...")

	//listener.CloseListener()
	//logger.Println("stake event listener closed")

	// 创建5s的超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.WithFields(logrus.Fields{
			"action":     "shutdown_http",
			"error_code": "HTTP_SHUTDOWN_FAIL",
			"detail":     err.Error(),
		}).Error("Server shutdown error")
	}
	log.WithFields(logrus.Fields{
		"action": "exit",
		"detail": "Server exiting",
	}).Info("Server exiting")
}
