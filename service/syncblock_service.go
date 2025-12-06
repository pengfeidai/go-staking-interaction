package service

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"staking-interaction/common/config"
	"staking-interaction/dto"
	"staking-interaction/model"
	"staking-interaction/repository"
	"staking-interaction/utils"
	"strconv"
	"strings"
	"time"
)

var (
	isSyncRunning bool
)

type SyncBlockService struct {
	client *ethclient.Client
}

func NewSyncBlockService(client *ethclient.Client) *SyncBlockService {
	return &SyncBlockService{
		client: client,
	}
}

func (s *SyncBlockService) SyncBlocks() {
	isSyncRunning = true
	chainID, err := s.client.ChainID(context.Background())
	if err != nil {
		log.Fatalf("SyncBlockInfo: Failed to get ChainID: %v", err)
	}
	doneBlock, err := s.client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("SyncBlockInfo: Failed to get start block number: %v", err)
	}
	fmt.Println("startBlock:", doneBlock)

	for isSyncRunning {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		currentBlock, err := s.client.BlockNumber(ctx)
		if err != nil {
			fmt.Println("SyncBlockInfo: Failed to get current block number: %v", err)
			cancel()
			time.Sleep(10 * time.Second)
			continue
		}
		if currentBlock < doneBlock+30 {
			fmt.Println("currentBlock: --", currentBlock)
			cancel()
			time.Sleep(10 * time.Second)
			continue
		}
		block, err := s.client.BlockByNumber(ctx, big.NewInt(int64(doneBlock)))
		if err != nil {
			fmt.Println("SyncBlockInfo: Failed to get block: %v", err)
			cancel()
			time.Sleep(10 * time.Second)
			continue
		}
		for _, tx := range block.Transactions() {
			receipt, err := s.client.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				fmt.Println("SyncBlockInfo: Failed to get transaction receipt: ", err)
				continue
			}
			if tx.To() == nil {
				fmt.Println("SyncBlockInfo: it's contract creation, skip")
				continue
			}
			// to 是否是平台账号
			toAddr := *tx.To()
			if isToAddrValid := isToAddrValid(toAddr); !isToAddrValid {
				fmt.Println("SyncBlockInfo: not to our portfolio")
				continue
			}
			fmt.Println("Block Transaction Hash: ", tx.Hash().Hex())
			if receipt.Status == types.ReceiptStatusSuccessful {
				//receipt 状态成功，且非创建合约的tx
				isContract, err := isContractTx(ctx, s.client, toAddr)
				if err != nil {
					fmt.Println("SyncBlockInfo: Failed to get if it is contract tx: ", err)
					continue
				}

				//get from address
				signer := types.NewLondonSigner(chainID)
				fromAddr, err := types.Sender(signer, tx)
				if err != nil {
					fmt.Println("SyncBlockInfo: Failed to get sender address: ", err)
					continue
				}
				fmt.Println("fromAddr:", fromAddr.Hex())

				//判断from address是否是本平台用户
				isValidAccount, accountId := isValidAccount(fromAddr)
				if !isValidAccount {
					fmt.Println("SyncBlockInfo: not our customer")
					continue
				}

				if isContract {
					handleERC20Tx(tx, receipt, *accountId)
				} else if len(tx.Data()) == 0 {
					//普通交易
					handleBnbTx(tx, fromAddr, receipt, *accountId)
				}
			}

			fmt.Println("Block Transaction Hash: %v", tx.Hash().Hex())
			fmt.Println("Block Transaction ContractAddress: %v", receipt.ContractAddress.Hex())
			fmt.Println("Block Transaction To: %v", toAddr.Hex())
			fmt.Println("Block Transaction Value: %v", tx.Value())
			fmt.Println("Block Transaction Data: %v", common.Bytes2Hex(tx.Data()))

		}
		doneBlock++
		fmt.Println("doneBlock: --", doneBlock)
		cancel()
	}
}

func isContractTx(ctx context.Context, client *ethclient.Client, to common.Address) (bool, error) {
	// 调用eth_getCode：合约地址返回代码，外部账户返回空字符串
	code, err := client.CodeAt(ctx, to, nil)
	fmt.Println("Contract Code:", code)
	if err != nil {
		fmt.Println("isContractTx error:", err)
		return false, err
	}

	return len(code) > 0, nil

}

func handleERC20Tx(tx *types.Transaction, receipt *types.Receipt, accountId int) {
	transEvent, err := parseERC20TxByReceipt(receipt)
	if err != nil {
		fmt.Println("handleERC20Tx: Failed to get amount err:", err)
		return
	}
	amount := transEvent.Value
	to := transEvent.ToAddress // *tx.To() 是token合约地址
	from := transEvent.FromAddress

	asset, err := repository.GetAccountAsset(accountId)
	if err != nil {
		fmt.Println("SyncBlockInfo: Failed to get asset: ", err)
		return
	}
	preBalance, err := utils.StringToBigInt(asset.MtkBalance)
	if err != nil {
		fmt.Println("SyncBlockInfo: Failed to parse pre balance: ", asset.MtkBalance, " Error: ", err)
	}
	nextBalance := new(big.Int).Add(preBalance, amount)

	bill := model.Bill{
		AccountID:   accountId,
		TokenType:   config.TokenTypeMTK,
		BillType:    config.BillTypeRecharge,
		Amount:      amount.String(),
		Fee:         strconv.FormatUint(receipt.GasUsed, 10),
		PreBalance:  asset.MtkBalance,
		NextBalance: nextBalance.String(),
	}
	if err := repository.AddBill(&bill); err != nil {
		fmt.Println("SyncBlockInfo: Failed to add bill: ", err)
	}

	transLog := model.TransactionLog{
		AccountID:   accountId,
		TokenType:   config.TokenTypeMTK,
		Hash:        tx.Hash().Hex(),
		Amount:      amount.String(),
		FromAddress: from.Hex(),
		ToAddress:   to.Hex(),
		BlockNumber: receipt.BlockNumber.String(),
	}
	if err := repository.AddTransactionLog(&transLog); err != nil {
		fmt.Println("SyncBlockInfo: Failed to add transaction logger: ", err)
		return
	}

	asset.MtkBalance = nextBalance.String()
	if err := repository.UpdateAsset(asset); err != nil {
		fmt.Println("SyncBlockInfo: Failed to update asset: ", err)
	}

	isSyncRunning = false
	fmt.Println("SyncBlockInfo: add logger and bill success ", transLog)
}

func parseERC20TxByReceipt(receipt *types.Receipt) (*dto.TransferEvent, error) {
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
	conf := config.Get()
	var tokenAddr = common.HexToAddress(conf.BlockchainConfig.Contracts.Token)

	erc20ABI, err := abi.JSON(strings.NewReader(erc20TransferEventABIJson))
	if err != nil {
		return nil, fmt.Errorf("parseERC20TxByReceipt: Failed to parse ERC20 Transfer Event ABI: %v", err)
	}

	transferEventSig := erc20ABI.Events["Transfer"].ID.Hex()

	for _, log := range receipt.Logs {
		// mtk token
		if log.Address != tokenAddr {
			continue
		}
		// confirm event sig
		if common.BytesToHash(log.Topics[0].Bytes()).Hex() != transferEventSig {
			continue
		}

		// Topics[0]是事件签名
		tran.FromAddress = common.HexToAddress(log.Topics[1].String())
		tran.ToAddress = common.HexToAddress(log.Topics[2].String())

		if err := erc20ABI.UnpackIntoInterface(&tran, "Transfer", log.Data); err != nil {
			return nil, fmt.Errorf("parseERC20TxByReceipt: Failed to unpack Transfer: %v", err)
		}
		return &tran, nil
	}
	return nil, fmt.Errorf("parseERC20TxByReceipt: no ERC-20 Transfer event found in receipt")
}

func isValidAccount(fromAddr common.Address) (bool, *int) {
	account, err := repository.GetAccount(fromAddr.Hex())
	if err != nil || account.AccountID <= 0 {
		return false, nil
	}
	return true, &account.AccountID
}

func isToAddrValid(toAddr common.Address) bool {
	cfg := config.Get()
	// to address是否是本平台账户
	for _, addr := range cfg.BlockchainConfig.Owners {
		if common.HexToAddress(addr) == toAddr {
			fmt.Println("SyncBlockInfo: to is our portfolio", addr, " toAddr: ", toAddr)
			return true
		}
	}
	return false
}

func handleBnbTx(tx *types.Transaction, from common.Address, receipt *types.Receipt, accountId int) {
	amount := tx.Value()

	asset, err := repository.GetAccountAsset(accountId)
	if err != nil {
		fmt.Println("handleBnbTx: Failed to get asset: ", err)
		return
	}
	preBalance, err := utils.StringToBigInt(asset.BnbBalance)
	if err != nil {
		fmt.Println("handleBnbTx: Failed to parse pre balance: ", err)
		return
	}
	nextBalance := new(big.Int).Add(preBalance, amount)

	bill := model.Bill{
		AccountID:   accountId,
		TokenType:   config.TokenTypeBNB,
		BillType:    config.BillTypeRecharge,
		Amount:      amount.String(),
		Fee:         strconv.FormatUint(receipt.GasUsed, 10),
		PreBalance:  asset.BnbBalance,
		NextBalance: nextBalance.String(),
	}
	if err := repository.AddBill(&bill); err != nil {
		fmt.Println("handleBnbTx: Failed to add bill: ", err)
	}

	transLog := model.TransactionLog{
		AccountID:   accountId,
		TokenType:   config.TokenTypeBNB,
		Hash:        tx.Hash().Hex(),
		Amount:      amount.String(),
		FromAddress: from.Hex(),
		ToAddress:   tx.To().Hex(),
		BlockNumber: receipt.BlockNumber.String(),
	}
	if err := repository.AddTransactionLog(&transLog); err != nil {
		fmt.Println("handleBnbTx: Failed to add transaction logger: ", err)
		return
	}

	asset.BnbBalance = nextBalance.String()
	if err := repository.UpdateAsset(asset); err != nil {
		fmt.Println("handleBnbTx: Failed to update asset: ", err)
	}
	fmt.Println("handleBnbTx: add logger and bill success ")
}

func parseERC20TxByData(tx *types.Transaction) (*big.Int, error) {
	var erc20ABIJson = `[
    {
        "constant": false,
        "inputs": [
            {"name": "to", "type": "address"},
            {"name": "value", "type": "uint256"}
        ],
        "name": "transfer",
        "outputs": [{"name": "", "type": "bool"}],
        "type": "function"
    },
    {
        "constant": false,
        "inputs": [
            {"name": "from", "type": "address"},
            {"name": "to", "type": "address"},
            {"name": "value", "type": "uint256"}
        ],
        "name": "transferFrom",
        "outputs": [{"name": "", "type": "bool"}],
        "type": "function"
    }
]`
	data := tx.Data()
	if len(data) < 4 { // 至少包含4字节方法签名
		return nil, fmt.Errorf("invalid data length")
	}

	// 2. 解析ABI
	erc20ABI, err := abi.JSON(strings.NewReader(erc20ABIJson))
	if err != nil {
		return nil, err
	}

	// 3. 提取方法签名（前4字节），匹配transfer或transferFrom
	methodSig := common.BytesToHash(data[:4]).Hex()
	switch methodSig {
	case "0xa9059cbb": // transfer方法
		// 解码参数：(address to, uint256 value)
		var params struct {
			To    common.Address
			Value *big.Int
		}
		// 从第4字节开始解码参数（跳过方法签名）
		if err := erc20ABI.UnpackIntoInterface(&params, "transfer", data[4:]); err != nil {
			return nil, err
		}
		return params.Value, nil

	case "0x23b872dd": // transferFrom方法
		// 解码参数：(address from, address to, uint256 value)
		var params struct {
			From  common.Address
			To    common.Address
			Value *big.Int
		}
		if err := erc20ABI.UnpackIntoInterface(&params, "transferFrom", data[4:]); err != nil {
			return nil, err
		}
		return params.Value, nil

	default:
		return nil, fmt.Errorf("not an ERC-20 transfer transaction")
	}
}
