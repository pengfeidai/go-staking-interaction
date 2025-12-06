package repository

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"staking-interaction/adapter"
)

type TxRepository struct {
	db    *gorm.DB // 事务对应的 GORM DB 实例（已绑定事务）
	redis *redis.Client
}

// TxWithTransaction 核心事务封装函数
// 参数：fn 是“需要在事务内执行的业务逻辑”（接收 TxRepository，返回 error）
func TxWithTransaction(fn func(txRepo *TxRepository) error) error {
	return adapter.DB.Transaction(func(tx *gorm.DB) error {
		txRepo := &TxRepository{db: tx} // 所有操作共享同一个事务

		if err := fn(txRepo); err != nil {
			// GORM 会自动回滚事务
			return fmt.Errorf("sync block with transaction failed: %w", err)
		}

		return nil // GORM 会自动提交事务
	})
}
