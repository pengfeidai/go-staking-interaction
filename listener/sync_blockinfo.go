package listener

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"math/big"
	"staking-interaction/adapter"
	"staking-interaction/common/config"
	"staking-interaction/common/redis"
	"staking-interaction/dto"
	"staking-interaction/model"
	"staking-interaction/repository"
	"staking-interaction/utils"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// SyncBlock 区块同步服务
type SyncBlock struct {
	client        *adapter.InitClient
	isSyncRunning int32         // 原子操作控制同步状态
	mu            sync.RWMutex  // 保护doneBlock的读写
	doneBlock     uint64        // 当前已同步的区块号
	workerPool    chan struct{} // 工作池控制并发数量
	blockManager  *repository.BlockSyncManager
	workerWg      sync.WaitGroup // 等待所有交易处理Goroutine退出
	lockManager   *redis.LockManager
	config        config.BlockchainConfig
	log           *logrus.Logger
}

// NewSyncBlockInfo 创建新的区块同步服务
func NewSyncBlockInfo(clientInfo *adapter.InitClient, config config.BlockchainConfig, lockManager *redis.LockManager, log *logrus.Logger) *SyncBlock {
	return &SyncBlock{
		client:       clientInfo,
		blockManager: repository.NewBlockSyncManager("last_synced_block.txt"),
		workerPool:   make(chan struct{}, config.Sync.Workers), // 限制并发处理数量
		config:       config,
		lockManager:  lockManager,
		log:          log,
	}
}

// Start 启动区块同步
func (s *SyncBlock) Start() {
	if !atomic.CompareAndSwapInt32(&s.isSyncRunning, 0, 1) {
		s.log.WithFields(logrus.Fields{
			"module": "sync_block",
			"action": "start",
			"result": "already_running",
		}).Warn("Sync service is already running")
		return
	}
	s.log.WithFields(logrus.Fields{
		"module": "sync_block",
		"action": "start",
		"result": "running",
	}).Info("Sync service is running...")

	// 初始化起始区块
	if err := s.initializeStartBlock(); err != nil {
		s.log.WithFields(logrus.Fields{
			"module":     "sync_block",
			"action":     "init_start_block",
			"error_code": "INIT_START_BLOCK_FAIL",
			"detail":     err.Error(),
		}).Error("Initialize start block failed")
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.log.WithFields(logrus.Fields{
					"module":     "sync_block",
					"action":     "init_start_block",
					"error_code": "PANIC",
					"detail":     r,
				}).Error("Recovered from panic during start block initialization")
				return
			}
		}()
		// 处理扫块逻辑
		s.syncLoop()
	}()

	s.log.WithFields(logrus.Fields{
		"module":             "sync_block",
		"action":             "start",
		"result":             "success",
		"initial_done_block": s.getDoneBlock(),
	}).Info("Sync service started successfully")
}

func (s *SyncBlock) initializeStartBlock() error {
	//if err := s.loadLastSyncedBlock(); err != nil || s.getDoneBlock() == 0 {
	//fmt.Println("Failed to load last synced block, starting from current", "error", err)
	if s.getDoneBlock() == 0 {
		ctx, cancel := context.WithTimeout(context.Background(), s.config.Sync.SyncInterval)
		defer cancel()

		currentBlock, e := s.client.Client.BlockNumber(ctx)
		if e != nil {
			return fmt.Errorf("failed to get current block: %w", e)
		}
		s.setDoneBlock(currentBlock)
	}
	return nil
}

// Stop 停止区块同步
func (s *SyncBlock) Stop() {
	atomic.StoreInt32(&s.isSyncRunning, 0)
	s.log.WithFields(logrus.Fields{
		"module": "sync_block",
		"action": "stop",
	}).Info("Sync service stopping")

	s.workerWg.Wait()
	s.log.WithFields(logrus.Fields{
		"module": "sync_block",
		"action": "stop",
	}).Info("Sync service stopped")
}

func (s *SyncBlock) syncLoop() {
	chainID := s.client.ChainID
	for atomic.LoadInt32(&s.isSyncRunning) == 1 {
		blockCtx, cancel := context.WithTimeout(context.Background(), s.config.Sync.SyncInterval)

		currentBlock, err := s.client.Client.BlockNumber(blockCtx)
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"module":            "sync_block",
				"action":            "get_current_block",
				"error_code":        "BLOCK_NUMBER_FAIL",
				"last_synced_block": s.getDoneBlock(),
				"detail":            err.Error(),
			}).Warn("Failed to get current block number")
			cancel()
			time.Sleep(10 * time.Second)
			continue
		}

		doneBlock := s.getDoneBlock()
		// 控制同步距离，避免超前太多
		if currentBlock < doneBlock+s.config.Sync.BlockBuffer {
			s.log.WithFields(logrus.Fields{
				"module":        "sync_block",
				"action":        "wait_new_block",
				"current_block": currentBlock,
				"done_block":    doneBlock,
			}).Info("Waiting for new blocks")
			cancel()
			time.Sleep(10 * time.Second)
			continue
		}

		// 处理当前区块
		if err := s.processBlock(blockCtx, chainID, doneBlock); err != nil {
			s.log.WithFields(logrus.Fields{
				"module":     "sync_block",
				"action":     "process_block",
				"block":      doneBlock,
				"error_code": "PROCESS_BLOCK_FAIL",
				"detail":     err.Error(),
			}).Error("Failed to process block")
			cancel()
			// 处理区块错误时重试延迟加倍
			time.Sleep(20 * time.Second)
			continue
		}

		// 更新已完成区块号并持久化
		newDoneBlock := doneBlock + 1
		s.setDoneBlock(newDoneBlock)
		// 保存区块到文件中
		if err := s.saveSyncedBlock(newDoneBlock); err != nil {
			s.log.WithFields(logrus.Fields{
				"module":     "sync_block",
				"action":     "save_synced_block",
				"block":      newDoneBlock,
				"error_code": "SAVE_BLOCK_FAIL",
				"detail":     err.Error(),
			}).Error("Failed to save synced block")
		} else {
			s.log.WithFields(logrus.Fields{
				"module":         "sync_block",
				"action":         "save_synced_block",
				"block":          doneBlock,
				"new_done_block": newDoneBlock,
				"result":         "success",
			}).Info("Processed block")
		}
		cancel()
	}
}

// 处理单个区块
func (s *SyncBlock) processBlock(blockCtx context.Context, chainID *big.Int, blockNumber uint64) error {
	// 获取区块详情
	block, err := s.client.Client.BlockByNumber(blockCtx, big.NewInt(int64(blockNumber)))
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			s.log.WithFields(logrus.Fields{
				"module":       "sync_block",
				"action":       "get_block",
				"block_number": blockNumber,
				"error_code":   "BLOCK_NOT_FOUND",
				"detail":       err.Error(),
			}).Error("Block not found")
			return fmt.Errorf("block %d not found: %w", blockNumber, err)
		}
		s.log.WithFields(logrus.Fields{
			"module":       "sync_block",
			"action":       "get_block",
			"block_number": blockNumber,
			"error_code":   "GET_BLOCK_FAIL",
			"detail":       err.Error(),
		}).Error("Get block failed")
		return fmt.Errorf("get block failed: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"module":            "sync_block",
		"action":            "process_block_start",
		"block_number":      blockNumber,
		"transaction_count": len(block.Transactions()),
	}).Info("Start processing block")

	var txWg sync.WaitGroup
	for _, tx := range block.Transactions() {
		txWg.Add(1)
		s.workerPool <- struct{}{} // 使用工作池控制并发
		s.workerWg.Add(1)          // 绑定全局交易WaitGroup

		go func(tx *types.Transaction, blockNum uint64) {
			txCtx, txCancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer func() {
				txCancel()
				txWg.Done()
				<-s.workerPool
				s.workerWg.Done()

				// 捕获panic，防止单个交易处理崩溃整个服务
				if r := recover(); r != nil {
					s.log.WithFields(logrus.Fields{
						"module":  "sync_block",
						"action":  "process_transaction_panic",
						"tx_hash": tx.Hash().Hex(),
						"panic":   r,
					}).Error("Recovered from panic in transaction processing")
				}
			}()

			if err := s.processTransaction(txCtx, chainID, tx); err != nil {
				s.log.WithFields(logrus.Fields{
					"module":       "sync_block",
					"action":       "process_transaction",
					"tx_hash":      tx.Hash().Hex(),
					"block_number": blockNum,
					"error_code":   "PROCESS_TRANSACTION_FAIL",
					"detail":       err.Error(),
				}).Error("Failed to process transaction")
			}
		}(tx, blockNumber)
	}

	txWg.Wait()
	s.log.WithFields(logrus.Fields{
		"module":            "sync_block",
		"action":            "process_block_end",
		"block_number":      blockNumber,
		"transaction_count": len(block.Transactions()),
		"result":            "success",
	}).Info("Processed block success")
	return nil
}

// 处理单个交易
func (s *SyncBlock) processTransaction(txCtx context.Context, chainID *big.Int, tx *types.Transaction) error {
	// 获取交易回执
	receipt, err := s.client.Client.TransactionReceipt(txCtx, tx.Hash())
	if err != nil {
		return fmt.Errorf("get transaction receipt: %w", err)
	}

	// 跳过合约创建交易
	if tx.To() == nil {
		s.log.WithFields(logrus.Fields{
			"module":  "sync_block",
			"action":  "skip_contract_creation",
			"tx_hash": tx.Hash().Hex(),
		}).Info("Skipping contract creation transaction")
		return nil
	}

	toAddr := *tx.To()
	// 检查接收地址是否为平台地址
	if !s.isToAddrValid(toAddr) {
		s.log.WithFields(logrus.Fields{
			"module":     "sync_block",
			"action":     "skip_not_to_our_address",
			"tx_hash":    tx.Hash().Hex(),
			"to_address": toAddr.Hex(),
		}).Info("Transaction not to our address")
		return nil
	}

	// 只处理成功的交易
	if receipt.Status != types.ReceiptStatusSuccessful {
		s.log.WithFields(logrus.Fields{
			"module":  "sync_block",
			"action":  "skip_failed_transaction",
			"tx_hash": tx.Hash().Hex(),
		}).Info("Skipping failed transaction")
		return nil
	}

	// 获取发送者地址
	signer := types.NewLondonSigner(chainID)
	fromAddr, err := types.Sender(signer, tx)
	if err != nil {
		return fmt.Errorf("get sender address: %w", err)
	}

	// 检查发送者是否为平台用户
	IsPlatformAccount, accountId := isFromAddrValid(fromAddr)
	if !IsPlatformAccount {
		s.log.WithFields(logrus.Fields{
			"module":       "sync_block",
			"action":       "skip_not_customer",
			"tx_hash":      tx.Hash().Hex(),
			"from_address": fromAddr.Hex(),
		}).Info("Sender is not our customer")
		return nil
	}

	// 判断是否为合约交易
	isContract, err := s.isContractTx(txCtx, toAddr)
	if err != nil {
		return fmt.Errorf("check contract tx: %w", err)
	}

	// 处理不同类型的交易
	if isContract {
		return s.handleERC20Tx(tx, receipt, *accountId)
	} else if len(tx.Data()) == 0 {
		// 处理普通BNB交易
		return s.handleTokenTransaction(
			tx,
			receipt,
			*accountId,
			fromAddr,
			*tx.To(),
			tx.Value(),
			config.TokenTypeBNB,
		)
	}

	s.log.WithFields(logrus.Fields{
		"module":  "sync_block",
		"action":  "unsupported_tx_type",
		"tx_hash": tx.Hash().Hex(),
	}).Warn("Unsupported transaction type")
	return nil
}

// 通用代币交易处理逻辑
func (s *SyncBlock) handleTokenTransaction(
	tx *types.Transaction,
	receipt *types.Receipt,
	accountId int,
	fromAddr, toAddr common.Address,
	amount *big.Int,
	tokenType int,
) error {
	//获取锁
	assetLock, err := s.lockManager.AcquireAssetLock(context.Background(), accountId, tokenType)
	if err != nil {
		return fmt.Errorf("acquire assetLock failed: %w ,accountid:%d, tx_hash:%s", err, accountId, tx.Hash().Hex())
	}

	//释放锁
	defer func() {
		unlockCtx, unlockCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer unlockCancel()

		if err := assetLock.Unlock(unlockCtx); err != nil {
			s.log.WithFields(logrus.Fields{
				"module":     "sync_block",
				"action":     "unlock_asset",
				"tx_hash":    tx.Hash().Hex(),
				"account_id": accountId,
				"error_code": "UNLOCK_FAIL",
				"detail":     err.Error(),
			}).Error("Unlock assetLock failed")
		} else {
			s.log.WithFields(logrus.Fields{
				"module":       "sync_block",
				"action":       "unlock_asset",
				"tx_hash":      tx.Hash().Hex(),
				"block_number": receipt.BlockNumber.Int64(),
				"account_id":   accountId,
				"result":       "success",
			}).Info("Asset lock released successfully")
		}
	}()

	return s.executeTransactionWithLock(receipt, accountId, fromAddr, toAddr, amount, tokenType)
}

func (s *SyncBlock) executeTransactionWithLock(
	receipt *types.Receipt,
	accountId int,
	fromAddr, toAddr common.Address,
	amount *big.Int,
	tokenType int) error {
	return repository.TxWithTransaction(func(txRepo *repository.TxRepository) error {
		hash := receipt.TxHash.String()
		// 确认交易记录是否已存在
		isExistTx, err := txRepo.TransactionExists(hash)
		if err != nil {
			return fmt.Errorf("get TransactionExists failed: %w, hash:%s", err, hash)
		}
		if isExistTx {
			return fmt.Errorf("this transaction is existed, hash:%s", hash)
		}

		// 获取账户资产
		asset, err := txRepo.GetAssetByAccountIdWithLock(accountId)
		if err != nil {
			return fmt.Errorf("get account asset: %w", err)
		}

		// 计算new balance
		preBalance, nextBalance, err := s.calculateBalance(tokenType, asset, amount)
		if err != nil {
			return fmt.Errorf("unsupported token type: %d", tokenType)
		}

		// 创建账单记录
		bill := model.Bill{
			AccountID:   accountId,
			TokenType:   tokenType,
			BillType:    config.BillTypeRecharge,
			Amount:      amount.String(),
			Fee:         strconv.FormatUint(receipt.GasUsed, 10),
			PreBalance:  preBalance,
			NextBalance: nextBalance,
			CreatedAt:   time.Now(),
		}
		if err := txRepo.AddBill(&bill); err != nil {
			return fmt.Errorf("add bill: %w", err)
		}

		// 创建交易日志
		transLog := model.TransactionLog{
			AccountID:   accountId,
			TokenType:   tokenType,
			Hash:        hash,
			Amount:      amount.String(),
			FromAddress: fromAddr.Hex(),
			ToAddress:   toAddr.Hex(),
			BlockNumber: receipt.BlockNumber.String(),
			CreatedAt:   time.Now(),
		}
		if err := txRepo.AddTransactionLog(&transLog); err != nil {
			return fmt.Errorf("add transaction logger: %w", err)
		}

		// 更新资产余额（使用乐观锁）
		if err := txRepo.UpdateAssetWithOptimisticLock(asset, nextBalance, tokenType); err != nil {
			return fmt.Errorf("update asset: %w", err)
		}

		s.log.WithFields(logrus.Fields{
			"module":     "sync_block",
			"action":     "process_transaction_success",
			"hash":       hash,
			"account_id": accountId,
			"token_type": tokenType,
			"amount":     amount.String(),
		}).Info("Successfully processed transaction")
		return nil
	})
}
func (s *SyncBlock) calculateBalance(tokenType int, asset *model.AccountAsset, amount *big.Int) (string, string, error) {
	var preBalanceStr string
	var err error
	var preBalance *big.Int

	switch tokenType {
	case config.TokenTypeMTK:
		preBalanceStr = asset.MtkBalance
		preBalance, err = utils.StringToBigInt(preBalanceStr)
		if err != nil {
			return "", "", fmt.Errorf("parse pre balance: %w", err)
		}

	case config.TokenTypeBNB:
		preBalanceStr = asset.BnbBalance
		preBalance, err = utils.StringToBigInt(preBalanceStr)
		if err != nil {
			return "", "", fmt.Errorf("parse pre balance: %w", err)
		}

	default:
		return "", "", fmt.Errorf("unsupported token type: %d", tokenType)
	}

	nextBalance := new(big.Int).Add(preBalance, amount)
	return preBalanceStr, nextBalance.String(), nil
}

// 处理ERC20代币交易
func (s *SyncBlock) handleERC20Tx(tx *types.Transaction, receipt *types.Receipt, accountId int) error {
	transEvent, err := s.parseERC20TxByReceipt(receipt)
	if err != nil {
		return fmt.Errorf("parse erc20 tx: %w", err)
	}

	return s.handleTokenTransaction(
		tx,
		receipt,
		accountId,
		transEvent.FromAddress,
		transEvent.ToAddress,
		transEvent.Value,
		config.TokenTypeMTK,
	)
}

// 解析ERC20交易事件
func (s *SyncBlock) parseERC20TxByReceipt(receipt *types.Receipt) (*dto.TransferEvent, error) {
	tran := dto.TransferEvent{}
	var erc20TransferEventABIJson = `[
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "from", "type": "address"},
				{"indexed": true, "name": "to", "type": "address"},
				{"indexed": false, "name": "value", "type": "uint256"}
			],
			"name": "Transfer",
			"type": "event"
		}
	]`

	tokenAddr := common.HexToAddress(s.config.Contracts.Token)
	erc20ABI, err := abi.JSON(strings.NewReader(erc20TransferEventABIJson))
	if err != nil {
		return nil, fmt.Errorf("parse erc20 abi: %w", err)
	}

	transferEventSig := erc20ABI.Events["Transfer"].ID.Hex()

	for _, log := range receipt.Logs {
		if log.Address != tokenAddr {
			continue
		}

		if log.Topics[0].Hex() != transferEventSig {
			continue
		}

		tran.FromAddress = common.HexToAddress(log.Topics[1].String())
		tran.ToAddress = common.HexToAddress(log.Topics[2].String())

		if err := erc20ABI.UnpackIntoInterface(&tran, "Transfer", log.Data); err != nil {
			return nil, fmt.Errorf("unpack transfer data: %w", err)
		}
		return &tran, nil
	}
	return nil, fmt.Errorf("no erc20 transfer event found in receipt")
}

// 检查是否为合约地址
func (s *SyncBlock) isContractTx(blockCtx context.Context, to common.Address) (bool, error) {
	// 合约地址有字节码，普通地址无
	code, err := s.client.Client.CodeAt(blockCtx, to, nil)
	if err != nil {
		return false, fmt.Errorf("get code at address: %w", err)
	}
	return len(code) > 0, nil
}

// 检查接收地址是否有效
func (s *SyncBlock) isToAddrValid(toAddr common.Address) bool {
	for _, addr := range s.config.Owners {
		if common.HexToAddress(addr) == toAddr {
			return true
		}
	}
	return false
}

// 检查账户是否有效
func isFromAddrValid(fromAddr common.Address) (bool, *int) {
	account, err := repository.GetAccount(fromAddr.Hex())
	if err != nil {
		return false, nil
	}
	if account.AccountID > 0 {
		return true, &account.AccountID
	}
	return false, nil
}

func getTokenTypeName(tokenType int) string {
	switch tokenType {
	case config.TokenTypeMTK:
		return "MTK"
	case config.TokenTypeBNB:
		return "BNB"
	default:
		return strconv.Itoa(tokenType)
	}
}

// 加载上次同步的区块号
func (s *SyncBlock) loadLastSyncedBlock() error {
	block, err := s.blockManager.GetLastSyncedBlock()
	if err != nil {
		return err
	}
	s.setDoneBlock(block)
	return nil
}

// 保存已同步的区块号
func (s *SyncBlock) saveSyncedBlock(block uint64) error {
	return s.blockManager.SaveSyncedBlock(block)
}

// 获取当前已同步的区块号
func (s *SyncBlock) getDoneBlock() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.doneBlock
}

// 设置已同步的区块号
func (s *SyncBlock) setDoneBlock(block uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.doneBlock = block
}
