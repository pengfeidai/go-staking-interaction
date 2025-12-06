package repository

import (
	"fmt"
	"staking-interaction/adapter"
	"staking-interaction/model"
)

func AddBill(bill *model.Bill) error {
	if bill == nil {
		return fmt.Errorf("bill 不能为 nil")
	}
	err := adapter.DB.Create(bill).Error
	if err != nil {
		return fmt.Errorf("创建账单失败: %w", err)
	}
	return nil
}

// AddBill 事务内创建账单
func (t *TxRepository) AddBill(bill *model.Bill) error {
	if err := t.db.Create(bill).Error; err != nil {
		return fmt.Errorf("tx add bill failed: %w", err)
	}
	return nil
}

// AddBill 事务内创建账单
func (w *SwRepo) AddBill(bill *model.Bill) error {
	if err := w.db.Create(bill).Error; err != nil {
		return fmt.Errorf("tx add bill failed: %w", err)
	}
	return nil
}
