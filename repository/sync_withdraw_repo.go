package repository

import (
	"fmt"
	"gorm.io/gorm"
	"staking-interaction/adapter"
)

type SwRepo struct {
	db *gorm.DB // 事务对应的 GORM DB 实例（已绑定事务）
}

func SwWithTransaction(fn func(wd *SwRepo) error) error {
	db := adapter.DB

	return db.Transaction(func(wd *gorm.DB) error {
		SwRepo := &SwRepo{db: wd}

		if err := fn(SwRepo); err != nil {
			return fmt.Errorf("WithTransaction failed: %w", err)
		}

		return nil
	})
}
