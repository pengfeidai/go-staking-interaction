package repository

import (
	"fmt"
	"staking-interaction/adapter"
	"staking-interaction/model"
)

func (w *SwRepo) UpdateWithdrawalInfo(withdraw model.Withdrawal) error {
	if err := w.db.Save(withdraw).Error; err != nil {
		return fmt.Errorf("SwRepo UpdateWithdrawalInfo failed: %w", err)
	}
	return nil
}

func (w *WdRepo) UpdateWithdrawalInfo(withdraw model.Withdrawal) error {
	if err := w.db.Save(withdraw).Error; err != nil {
		return fmt.Errorf("WdRepo UpdateWithdrawalInfo failed: %w", err)
	}
	return nil
}

func GetWithdrawalInfoByStatus(status int) ([]model.Withdrawal, error) {
	var withdrawalInfo []model.Withdrawal
	if err := adapter.DB.Where("status=?", status).Find(&withdrawalInfo).Error; err != nil {
		return nil, fmt.Errorf("WdRepo GetWithdrawalInfoByStatus failed: %w", err)
	}
	return withdrawalInfo, nil
}
