package repository

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"staking-interaction/adapter"
	"staking-interaction/common/config"
	"staking-interaction/model"
	"time"
)

func GetAccount(fromAddress string) (*model.Account, error) {
	account := model.Account{}
	if err := adapter.DB.Model(&model.Account{}).Where("wallet_address = ?", fromAddress).First(&account).Error; err != nil {
		return nil, fmt.Errorf("repo: get account asset failed: %w", err)
	}

	return &account, nil
}

func GetAccountAsset(accountId int) (*model.AccountAsset, error) {
	asset := model.AccountAsset{}
	err := adapter.DB.Model(&model.AccountAsset{}).Where("account_id = ?", accountId).First(&asset).Error
	if err != nil {
		return nil, fmt.Errorf("repo: get account asset failed: %w", err)
	}

	return &asset, nil
}

func UpdateAsset(asset *model.AccountAsset) error {
	// 注意：需确保asset包含主键ID或唯一索引字段（如AccountID+TokenAddr）
	return adapter.DB.Save(asset).Error
}

// GetAssetByAccountIdWithLock 添加行锁的资产查询
func (t *TxRepository) GetAssetByAccountIdWithLock(accountID int) (*model.AccountAsset, error) {
	var asset model.AccountAsset

	// 使用 FOR UPDATE 添加行锁，防止并发修改
	err := t.db.Set("gorm:query_option", "FOR UPDATE").
		Where("account_id = ?", accountID).
		First(&asset).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("account asset not found for account_id: %d", accountID)
		}
		return nil, fmt.Errorf("get asset with lock failed: %w", err)
	}

	return &asset, nil
}

func (w *SwRepo) GetAssetByAccountIdWithLock(accountID int) (*model.AccountAsset, error) {
	var asset model.AccountAsset

	// 使用 FOR UPDATE 添加行锁，防止并发修改
	err := w.db.Set("gorm:query_option", "FOR UPDATE").
		Where("account_id = ?", accountID).
		First(&asset).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("account asset not found for account_id: %d", accountID)
		}
		return nil, fmt.Errorf("get asset with lock failed: %w", err)
	}

	return &asset, nil
}

// GetAssetByAccountIdWithLock 添加行锁的资产查询
func (w *WdRepo) GetAssetByAccountIdWithLock(accountID int) (*model.AccountAsset, error) {
	var asset model.AccountAsset

	// 使用 FOR UPDATE 添加行锁，防止并发修改
	err := w.db.Set("gorm:query_option", "FOR UPDATE").
		Where("account_id = ?", accountID).
		First(&asset).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("account asset not found for account_id: %d", accountID)
		}
		return nil, fmt.Errorf("get asset with lock failed: %w", err)
	}

	return &asset, nil
}

// UpdateAsset 事务内更新用户资产
func (t *TxRepository) UpdateAsset(asset *model.AccountAsset) error {
	if err := t.db.Save(asset).Error; err != nil {
		return fmt.Errorf("tx update asset failed: %w", err)
	}
	return nil
}

func (w *SwRepo) UpdateAssetWithOptimisticLock(
	asset *model.AccountAsset,
	newBalance string,
	tokenType int,
) error {
	// 记录当前版本号
	currentVersion := asset.Version

	// 更新余额和版本号
	switch tokenType {
	case config.TokenTypeMTK:
		asset.MtkBalance = newBalance
	case config.TokenTypeBNB:
		asset.BnbBalance = newBalance
	default:
		return fmt.Errorf("unsupported token type: %d", tokenType)
	}

	asset.Version = currentVersion + 1
	asset.UpdatedAt = time.Now()

	// 使用乐观锁更新（WHERE version = current_version）
	result := w.db.Model(asset).
		Where("account_id = ? AND version = ?", asset.AccountID, currentVersion).
		Updates(map[string]interface{}{
			"bnb_balance": asset.BnbBalance,
			"mtk_balance": asset.MtkBalance,
			"status":      config.WithdrawStatusSuccess,
			"version":     asset.Version,
			"updated_at":  asset.UpdatedAt,
		})

	if result.Error != nil {
		return fmt.Errorf("updateasset with optimistic lock failed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("optimistic lock failed: asset was modified by another process")
	}

	return nil
}

func (t *TxRepository) UpdateAssetWithOptimisticLock(
	asset *model.AccountAsset,
	newBalance string,
	tokenType int,
) error {
	// 记录当前版本号
	currentVersion := asset.Version

	// 更新余额和版本号
	switch tokenType {
	case config.TokenTypeMTK:
		asset.MtkBalance = newBalance
	case config.TokenTypeBNB:
		asset.BnbBalance = newBalance
	default:
		return fmt.Errorf("unsupported token type: %d", tokenType)
	}

	asset.Version = currentVersion + 1
	asset.UpdatedAt = time.Now()

	// 使用乐观锁更新（WHERE version = current_version）
	res := t.db.Model(asset).
		Where("account_id = ? AND version = ?", asset.AccountID, currentVersion).
		Updates(map[string]interface{}{
			"bnb_balance": asset.BnbBalance,
			"mtk_balance": asset.MtkBalance,
			"version":     asset.Version,
			"updated_at":  asset.UpdatedAt,
		})

	if res.Error != nil {
		return fmt.Errorf("updateasset with optimistic lock failed: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return fmt.Errorf("optimistic lock failed: asset was modified by another process")
	}

	return nil
}
