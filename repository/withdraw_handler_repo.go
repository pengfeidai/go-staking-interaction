package repository

import (
	"fmt"
	"gorm.io/gorm"
	"staking-interaction/adapter"
)

type WdRepo struct {
	db *gorm.DB // 事务对应的 GORM DB 实例（已绑定事务）
}

func WdWithTransaction(fn func(wd *WdRepo) error) error {
	db := adapter.DB

	return db.Transaction(func(wd *gorm.DB) error {
		wdRepo := &WdRepo{db: wd}

		if err := fn(wdRepo); err != nil {
			return fmt.Errorf("WithTransaction failed: %w", err)
		}

		return nil
	})
}
