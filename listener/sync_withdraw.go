package listener

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"math/big"
	"staking-interaction/common/config"
	"staking-interaction/common/redis"
	"staking-interaction/model"
	"staking-interaction/repository"
	"staking-interaction/utils"
	"strconv"
	"sync/atomic"
	"time"
)

type SyncWithdrawHandler struct {
	client                   *ethclient.Client
	isWithDrawHandlerRunning int32
	lockManager              *redis.LockManager
	log                      *logrus.Logger
}

var withdrawConf = config.Get().BlockchainConfig

func NewSyncWithdrawHandler(client *ethclient.Client, lockManager *redis.LockManager, log *logrus.Logger) *SyncWithdrawHandler {
	return &SyncWithdrawHandler{
		client:      client,
		lockManager: lockManager,
		log:         log,
	}
}

func (s *SyncWithdrawHandler) Start() {
	atomic.StoreInt32(&s.isWithDrawHandlerRunning, 1)
	s.syncWithdraw()
}

func (s *SyncWithdrawHandler) Stop() {
	atomic.StoreInt32(&s.isWithDrawHandlerRunning, 0)
	s.log.WithFields(logrus.Fields{
		"module": "sync_withdraw",
		"action": "service_stopping",
	}).Info("Service stopping, SyncWithdrawHandler")
}

func (s *SyncWithdrawHandler) syncWithdraw() {
	for atomic.LoadInt32(&s.isWithDrawHandlerRunning) == 1 {
		withDrawList, err := repository.GetWithdrawalInfoByStatus(config.WithdrawStatusPending)
		if err != nil || len(withDrawList) == 0 {
			s.log.WithFields(logrus.Fields{
				"module":     "sync_withdraw",
				"action":     "GetWithdrawInfo",
				"error_code": "GET_WITHDRAW_FAIL",
				"error":      err,
			}).Error("GetWithdrawInfo error or withdraw list is empty")
			time.Sleep(5 * time.Second)
			continue
		}
		for i, withdrawInfo := range withDrawList {
			if atomic.LoadInt32(&s.isWithDrawHandlerRunning) != 1 {
				s.log.WithFields(logrus.Fields{
					"module":    "withdraw_handler",
					"action":    "service_stopping",
					"processed": i,
					"total":     len(withDrawList),
				}).Info("Service stopping, processed withdrawals")
				break
			}

			err := s.processWithdraw(withdrawInfo)
			if err != nil {
				s.log.WithFields(logrus.Fields{
					"module":      "sync_withdraw",
					"action":      "processWithdraw",
					"withdraw_id": withdrawInfo.ID,
					"error_code":  "PARSE_WITHDRAW_FAIL",
					"detail":      err.Error(),
				}).Error("Process withdraw error")
			}
			s.log.WithFields(logrus.Fields{
				"module":         "sync_withdraw",
				"action":         "processWithdraw",
				"wallet_address": withdrawInfo.WalletAddress,
				"result":         "success",
			}).Info("processWithdraw successfully")
		}
		time.Sleep(5 * time.Second)
	}
}

func (s *SyncWithdrawHandler) processWithdraw(withdrawInfo model.Withdrawal) error {
	hash := common.HexToHash(withdrawInfo.Hash)
	ctx, cancel := context.WithTimeout(context.Background(), withdrawConf.Sync.SyncInterval)
	defer cancel()

	receipt, err := s.client.TransactionReceipt(ctx, hash)
	if err != nil {
		return fmt.Errorf("SyncWithdrawHandler: GetTransactionReceipt failed: %v\n", err)

	}
	tx, isPending, err := s.client.TransactionByHash(ctx, hash)
	if err != nil || isPending {
		return fmt.Errorf("SyncWithdrawHandler: GetTransactionByHash failed: %v\n", err)
	}

	//更新withdraw信息
	gasUsed := utils.Uint64ToBigInt(receipt.GasUsed)
	gasPrice := tx.GasPrice()
	fee := new(big.Int).Mul(gasPrice, gasUsed) // 手续费 = gasPrice × gasUsed
	value, err := utils.StringToBigInt(withdrawInfo.Value)
	if err != nil {
		return fmt.Errorf("SyncWithdrawHandler: StringToBigInt failed: %v\n", err)
	}
	amount := new(big.Int).Add(fee, value)
	withdrawInfo.Fee = fee.String()
	withdrawInfo.GasPrice = gasPrice.String()
	withdrawInfo.Amount = amount.String()

	return s.executeWithdrawWithLock(ctx, withdrawInfo, receipt, amount)
}

func (s *SyncWithdrawHandler) executeWithdrawWithLock(ctx context.Context, withdrawInfo model.Withdrawal, receipt *types.Receipt, amount *big.Int) error {
	account, err := repository.GetAccount(withdrawInfo.WalletAddress)
	if err != nil {
		return fmt.Errorf("get wallet account failed: %w", err)
	}

	//获取锁
	assetLock, err := s.lockManager.AcquireAssetLock(context.Background(), account.AccountID, withdrawInfo.TokenType)
	if err != nil {
		return fmt.Errorf("acquire assetLock failed: %w ,accountid:%d, tx_hash:%s", err, account.AccountID, withdrawInfo.Hash)
	}
	withdrawLock, err := s.lockManager.AcquireWithdrawLock(context.Background(), withdrawInfo.ID)
	if err != nil {
		return fmt.Errorf("acquire assetLock failed: %w ,withdrawid:%d, tx_hash:%s", err, withdrawInfo.ID, withdrawInfo.Hash)
	}

	//释放锁
	defer func() {
		unlockCtx, unlockCancel := context.WithTimeout(context.Background(), withdrawConf.Sync.RetryDelay)
		defer unlockCancel()

		if err := assetLock.Unlock(unlockCtx); err != nil {
			fmt.Printf("unlock assetLock failed: %w ,accountid:%d, tx_hash:%s\n", err, account.AccountID, withdrawInfo.Hash)
		}
		if err := withdrawLock.Unlock(unlockCtx); err != nil {
			fmt.Printf("unlock withdrawLock failed: %w ,withdrawid:%d, tx_hash:%s\n", err, withdrawInfo.ID, withdrawInfo.Hash)
		}
		fmt.Printf("lock  releasing: blocknumber:%s, tx_hash:%s\n", receipt.BlockNumber.String(), withdrawInfo.Hash)
	}()

	return s.handleWithdrawTransaction(ctx, receipt, withdrawInfo, account.AccountID, amount)
}

func (s *SyncWithdrawHandler) handleWithdrawTransaction(ctx context.Context, receipt *types.Receipt, withdrawInfo model.Withdrawal, accountID int, amount *big.Int) error {
	return repository.SwWithTransaction(func(wd *repository.SwRepo) error {
		// 检查交易是否成功
		if receipt.Status != types.ReceiptStatusSuccessful {
			// 更新提现状态为"失败"
			withdrawInfo.Status = config.WithdrawStatusFailed
			if err := wd.UpdateWithdrawalInfo(withdrawInfo); err != nil {
				fmt.Printf("SyncWithdrawHandler: UpdateWithdrawInfo failed: %v\n", err)
			}
			return fmt.Errorf("SyncWithdrawHandler: GetTransactionReceipt failed id: %d, hash:%s\n", withdrawInfo.ID, withdrawInfo.Hash)
		}
		withdrawInfo.Status = config.WithdrawStatusSuccess
		if err := wd.UpdateWithdrawalInfo(withdrawInfo); err != nil {
			fmt.Printf("SyncWithdrawHandler: UpdateWithdrawInfo failed: %v\n", err)
		}

		// 检查区块确认数是否足够
		blockNumber := receipt.BlockNumber
		currentHeight, err := s.client.BlockNumber(ctx)
		if err != nil {
			return fmt.Errorf("SyncWithdrawHandler: GetBlockNumber failed: %v\n", err)
		}
		expectBlockNumber := new(big.Int).Add(blockNumber, big.NewInt(int64(withdrawConf.Sync.BlockBuffer)))
		if big.NewInt(int64(currentHeight)).Cmp(expectBlockNumber) < 0 { //current height< block number+30
			return fmt.Errorf("SyncWithdrawHandler: current block height is not enough, currentHeight: %d, receipt block number:%d \n", currentHeight, blockNumber)
		}

		// 处理资产扣减和账单记录
		asset, err := wd.GetAssetByAccountIdWithLock(accountID)
		if err != nil {
			return fmt.Errorf("SyncWithdrawHandler: GetAccountAsset failed: %v, accountid:%d\n", err, accountID)
		}

		// 检查余额是否充足
		var preBalanceStr string
		switch withdrawInfo.TokenType {
		case config.TokenTypeBNB:
			preBalanceStr = asset.BnbBalance
		case config.TokenTypeMTK:
			preBalanceStr = asset.MtkBalance
		}
		preBalance, err := utils.StringToBigInt(preBalanceStr)
		if err != nil {
			return fmt.Errorf("SyncWithdrawHandler: preBalance parse failed: %v, preBalance:%s\n", err, preBalanceStr)
		}
		if preBalance.Cmp(amount) == -1 {
			return fmt.Errorf("SyncWithdrawHandler: balance is not enough, preBalance:%s, txamount:%s\n", preBalance, withdrawInfo.Amount)
		}

		// 计算nextBalance
		nextBalance := new(big.Int).Sub(preBalance, amount)

		// 创建提现账单记录
		bill := model.Bill{
			AccountID:   accountID,
			TokenType:   withdrawInfo.TokenType,
			BillType:    config.BillTypeWithdrawal,
			Amount:      amount.String(),
			Fee:         strconv.FormatUint(receipt.GasUsed, 10),
			PreBalance:  preBalanceStr,
			NextBalance: nextBalance.String(),
		}
		if err := wd.AddBill(&bill); err != nil {
			return fmt.Errorf("AddBill failed: %w", err)
		}
		s.log.WithFields(logrus.Fields{
			"module": "sync_withdraw",
			"action": "handleWithdrawTransaction",
			"result": "success",
		}).Info("update bill successfully")

		// 使用乐观锁更新账户资产余额
		if err := wd.UpdateAssetWithOptimisticLock(asset, nextBalance.String(), withdrawInfo.TokenType); err != nil {
			return fmt.Errorf("update asset: %w", err)
		}
		return nil
	})
}
