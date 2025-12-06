package listener

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"staking-interaction/common/config"
	"staking-interaction/common/redis"
	"staking-interaction/model"
	"staking-interaction/repository"
	"staking-interaction/service"
	"staking-interaction/utils"
	"sync/atomic"
	"time"
)

type WithdrawHandler struct {
	txService                *service.TransactionService
	isWithDrawHandlerRunning int32 // 原子操作控制同步状态
	lockManager              *redis.LockManager
	log                      *logrus.Logger
}

func NewWithdrawHandler(txService *service.TransactionService, lockManager *redis.LockManager, log *logrus.Logger) *WithdrawHandler {
	return &WithdrawHandler{
		txService:   txService,
		lockManager: lockManager,
		log:         log,
	}
}

func (w *WithdrawHandler) Start() {
	w.log.WithFields(logrus.Fields{
		"module": "withdraw_handler",
		"action": "start",
	}).Info("WithdrawHandler started")
	atomic.StoreInt32(&w.isWithDrawHandlerRunning, 1)

	for atomic.LoadInt32(&w.isWithDrawHandlerRunning) == 1 {
		w.processWithdrawals()
		time.Sleep(5 * time.Second)
	}
}

func (w *WithdrawHandler) Stop() {
	atomic.StoreInt32(&w.isWithDrawHandlerRunning, 0)
	w.log.WithFields(logrus.Fields{
		"module": "withdraw_handler",
		"action": "stop",
	}).Info("WithdrawHandler stopped")
}

func (w *WithdrawHandler) processWithdrawals() {
	withDrawList, err := repository.GetWithdrawalInfoByStatus(config.WithdrawStatusInit)
	if err != nil || len(withDrawList) == 0 {
		w.log.WithFields(logrus.Fields{
			"module":     "withdraw_handler",
			"action":     "get_withdrawals",
			"error_code": "GET_WITHDRAWALS_FAIL",
			"error":      err,
		}).Error("GetWithdrawInfo error or withdraw list is empty")
		return
	}

	for i, withdraw := range withDrawList {
		if atomic.LoadInt32(&w.isWithDrawHandlerRunning) != 1 {
			w.log.WithFields(logrus.Fields{
				"module":    "withdraw_handler",
				"action":    "service_stopping",
				"processed": i,
				"total":     len(withDrawList),
			}).Info("Service stopping, processed withdrawals")
			break
		}

		if err := w.executeHandlerWithLock(withdraw); err != nil {
			w.log.WithFields(logrus.Fields{
				"module":      "withdraw_handler",
				"action":      "process_withdrawal",
				"error_code":  "PROCESS_WITHDRAWAL_FAIL",
				"withdraw_id": withdraw.ID,
				"detail":      err.Error(),
			}).Error("Transaction failed")
		}
	}
}

func (w *WithdrawHandler) executeHandlerWithLock(withdraw model.Withdrawal) error {
	account, err := repository.GetAccount(withdraw.WalletAddress)
	if err != nil {
		w.log.WithFields(logrus.Fields{
			"module":         "withdraw_handler",
			"action":         "get_account",
			"wallet_address": withdraw.WalletAddress,
			"error_code":     "GET_ACCOUNT_FAIL",
			"detail":         err.Error(),
		}).Error("Get wallet account failed")
		return fmt.Errorf("get wallet account failed: %w", err)
	}
	accountId := account.AccountID

	//获取锁
	assetLock, err := w.lockManager.AcquireAssetLock(context.Background(), accountId, withdraw.TokenType)
	if err != nil {
		w.log.WithFields(logrus.Fields{
			"module":         "withdraw_handler",
			"action":         "acquire_lock",
			"account_id":     accountId,
			"token_type":     withdraw.TokenType,
			"wallet_address": withdraw.WalletAddress,
			"error_code":     "ACQUIRE_LOCK_FAIL",
			"detail":         err.Error(),
		}).Error("Acquire assetLock failed")
		return fmt.Errorf("acquire assetLock failed: %w ,wallet address:%d", err, withdraw.WalletAddress)
	}

	//释放锁
	defer func() {
		unlockCtx, unlockCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer unlockCancel()

		if err := assetLock.Unlock(unlockCtx); err != nil {
			w.log.WithFields(logrus.Fields{
				"module":         "withdraw_handler",
				"action":         "unlock_asset",
				"wallet_address": withdraw.WalletAddress,
				"error_code":     "UNLOCK_FAIL",
				"detail":         err.Error(),
			}).Error("Unlock assetLock failed")
		} else {
			w.log.WithFields(logrus.Fields{
				"module":         "withdraw_handler",
				"action":         "unlock_asset",
				"wallet_address": withdraw.WalletAddress,
				"result":         "success",
			}).Info("Asset lock released successfully")
		}
	}()

	return w.handleWithdrawTransaction(withdraw, accountId)
}

func (w *WithdrawHandler) handleWithdrawTransaction(withdraw model.Withdrawal, accountId int) error {
	return repository.WdWithTransaction(func(wdRepo *repository.WdRepo) error {
		var newWithdraw model.Withdrawal
		asset, err := wdRepo.GetAssetByAccountIdWithLock(accountId)
		if err != nil {
			w.log.WithFields(logrus.Fields{
				"module":     "withdraw_handler",
				"action":     "get_asset",
				"account_id": accountId,
				"error_code": "GET_ASSET_FAIL",
				"detail":     err.Error(),
			}).Error("Get asset by address failed")
			return fmt.Errorf("get asset by address failed: %w", err)
		}
		switch withdraw.TokenType {
		case config.TokenTypeBNB:
			res, err := w.transactionBNB(withdraw, asset.BnbBalance)
			if err != nil {
				w.log.WithFields(logrus.Fields{
					"module":      "withdraw_handler",
					"action":      "transaction_bnb",
					"withdraw_id": withdraw.ID,
					"error_code":  "TRANSACTION_BNB_FAIL",
					"detail":      err.Error(),
				}).Error("transactionBNB failed")
				return fmt.Errorf("transactionBNB failed: %w, withdrawid: %d", err, withdraw.ID)
			}
			newWithdraw = *res
		case config.TokenTypeMTK:
			res, err := w.transactionERC20(withdraw, asset.MtkBalance)
			if err != nil {
				w.log.WithFields(logrus.Fields{
					"module":      "withdraw_handler",
					"action":      "transaction_erc20",
					"withdraw_id": withdraw.ID,
					"error_code":  "TRANSACTION_ERC20_FAIL",
					"detail":      err.Error(),
				}).Error("transactionERC20 failed")
				return fmt.Errorf("transactionERC20 failed:  %w", err)
			}
			newWithdraw = *res
		}
		if err := wdRepo.UpdateWithdrawalInfo(newWithdraw); err != nil {
			w.log.WithFields(logrus.Fields{
				"module":      "withdraw_handler",
				"action":      "update_withdraw_info",
				"withdraw_id": newWithdraw.ID,
				"error_code":  "UPDATE_WITHDRAW_INFO_FAIL",
				"detail":      err.Error(),
			}).Error("Update withdraw info failed")
			return fmt.Errorf("update withdraw info failed: %w", err)
		}
		w.log.WithFields(logrus.Fields{
			"module":      "withdraw_handler",
			"action":      "update_withdraw_info",
			"withdraw_id": newWithdraw.ID,
			"result":      "success",
		}).Info("Update withdraw info success")
		return nil
	})
}

func (w *WithdrawHandler) transactionBNB(withdraw model.Withdrawal, bnbBalance string) (*model.Withdrawal, error) {
	value, err := utils.StringToBigInt(withdraw.Value)
	if err != nil {
		w.log.WithFields(logrus.Fields{
			"module":         "withdraw_handler",
			"action":         "parse_amount",
			"withdraw_id":    withdraw.ID,
			"withdraw_Value": withdraw.Value,
			"error_code":     "PARSE_AMOUNT_FAIL",
			"detail":         err.Error(),
		}).Error("Parse withdraw value error")
		return nil, fmt.Errorf("transactionBNB: parse withdraw value error: %w", err)
	}

	balance, err := utils.StringToBigInt(bnbBalance)
	if err != nil {
		w.log.WithFields(logrus.Fields{
			"module":      "withdraw_handler",
			"action":      "parse_balance",
			"withdraw_id": withdraw.ID,
			"bnbBalance":  bnbBalance,
			"error_code":  "PARSE_BALANCE_FAIL",
			"detail":      err.Error(),
		}).Error("Parse withdraw BNB balance error")
		return nil, fmt.Errorf("transactionBNB: parse withdraw bnb balance error: %w", err)
	}

	// 判断是否余额充足 balance>=value
	if value.Cmp(balance) >= 0 {
		w.log.WithFields(logrus.Fields{
			"module":      "withdraw_handler",
			"action":      "balance_check",
			"withdraw_id": withdraw.ID,
			"value":       value.String(),
			"balance":     balance.String(),
			"error_code":  "INSUFFICIENT_BALANCE",
		}).Error("Withdraw BNB balance is not enough")
		return nil, fmt.Errorf("transactionBNB: balance is not enough, value: %w, balance:%w", value, balance)
	}

	// 调用transfer BNB， 发送BNB
	tx, err := w.txService.SendBNB(withdraw.WalletAddress, value)
	if err != nil {
		w.log.WithFields(logrus.Fields{
			"module":         "withdraw_handler",
			"action":         "send_bnb",
			"withdraw_id":    withdraw.ID,
			"wallet_address": withdraw.WalletAddress,
			"value":          value.String(),
			"error_code":     "SEND_BNB_FAIL",
			"detail":         err.Error(),
		}).Error("Send withdraw BNB error")
		return nil, fmt.Errorf("transactionBNB: send withdraw value error: %w", err)
	}

	withdraw.Hash = tx.Hash
	withdraw.Status = config.WithdrawStatusPending

	w.log.WithFields(logrus.Fields{
		"module":         "withdraw_handler",
		"action":         "send_bnb",
		"withdraw_id":    withdraw.ID,
		"wallet_address": withdraw.WalletAddress,
		"value":          value.String(),
		"tx_hash":        tx.Hash,
		"result":         "pending",
	}).Info("BNB withdrawal transaction sent")

	return &withdraw, nil
}

func (w *WithdrawHandler) transactionERC20(withdraw model.Withdrawal, mtkBalance string) (*model.Withdrawal, error) {
	value, err := utils.StringToBigInt(withdraw.Value)
	if err != nil {
		w.log.WithFields(logrus.Fields{
			"module":      "withdraw_handler",
			"action":      "parse_amount",
			"withdraw_id": withdraw.ID,
			"error_code":  "PARSE_AMOUNT_FAIL",
			"detail":      err.Error(),
		}).Error("Parse withdraw value error")
		return nil, fmt.Errorf("transactionERC20: parse withdraw value error: %w, withdraw.Amount:%s", err, withdraw.Amount)
	}
	balance, err := utils.StringToBigInt(mtkBalance)
	if err != nil {
		w.log.WithFields(logrus.Fields{
			"module":      "withdraw_handler",
			"action":      "parse_balance",
			"withdraw_id": withdraw.ID,
			"error_code":  "PARSE_BALANCE_FAIL",
			"detail":      err.Error(),
		}).Error("Parse withdraw MTK balance error")
		return nil, fmt.Errorf("transactionERC20: parse withdraw mtk balance error: %w", err)
	}

	// 判断是否余额充足 balance>=value
	if value.Cmp(balance) >= 0 {
		w.log.WithFields(logrus.Fields{
			"module":      "withdraw_handler",
			"action":      "balance_check",
			"withdraw_id": withdraw.ID,
			"value":       value.String(),
			"balance":     balance.String(),
			"error_code":  "INSUFFICIENT_BALANCE",
		}).Error("Withdraw MTK balance is not enough")
		return nil, fmt.Errorf("transactionERC20: balance is not enough, value: %w, balance:%w", value, balance)
	}

	// 调用transfer Erc20， 发送MTK
	tx, err := w.txService.SendErc20(withdraw.WalletAddress, value)
	if err != nil {
		w.log.WithFields(logrus.Fields{
			"module":         "withdraw_handler",
			"action":         "send_erc20",
			"withdraw_id":    withdraw.ID,
			"wallet_address": withdraw.WalletAddress,
			"value":          value.String(),
			"error_code":     "SEND_ERC20_FAIL",
			"detail":         err.Error(),
		}).Error("Send withdraw ERC20 error")
		return nil, fmt.Errorf("transactionERC20:send SendErc20error: %w, walletAddress: %s, withdrawid:%d", err, withdraw.WalletAddress, withdraw.ID)
	}

	withdraw.Hash = tx.Hash
	withdraw.Status = config.WithdrawStatusPending

	w.log.WithFields(logrus.Fields{
		"module":         "withdraw_handler",
		"action":         "send_erc20",
		"withdraw_id":    withdraw.ID,
		"wallet_address": withdraw.WalletAddress,
		"value":          value.String(),
		"tx_hash":        tx.Hash,
		"result":         "pending",
	}).Info("ERC20 withdrawal transaction sent")

	return &withdraw, nil
}
