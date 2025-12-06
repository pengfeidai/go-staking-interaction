package repository

import (
	"fmt"
	"staking-interaction/adapter"
	"staking-interaction/model"
)

func AddTransactionLog(log *model.TransactionLog) error {
	if log == nil {
		return fmt.Errorf("logger 不能为 nil")
	}
	err := adapter.DB.Create(log).Error
	if err != nil {
		return fmt.Errorf("创建trancation log失败: %w", err)
	}
	return nil
}

// AddTransactionLog 事务内创建交易日志
func (t *TxRepository) AddTransactionLog(log *model.TransactionLog) error {
	if log == nil {
		return fmt.Errorf("logger 不能为 nil")
	}

	if err := t.db.Create(log).Error; err != nil {
		return fmt.Errorf("add trancation logger failed: %w", err)
	}
	return nil
}

// TransactionExists 事务内查询交易记录
func (t *TxRepository) TransactionExists(hash string) (bool, error) {
	var count int64
	err := t.db.Model(&model.TransactionLog{}).Where("hash = ?", hash).Select("count(*)").Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
