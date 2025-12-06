package service

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"staking-interaction/adapter"
	"staking-interaction/common/config"
	"staking-interaction/contracts/mtk"
	"staking-interaction/dto"
	"time"
)

type TransactionService struct {
	clientInfo *adapter.InitClient
}

func NewTransactionService(
	clientInfo *adapter.InitClient,
) *TransactionService {
	return &TransactionService{
		clientInfo: clientInfo,
	}
}

func (s *TransactionService) SendErc20(addr string, amount *big.Int) (res *dto.ERCRes, err error) {
	// 方案一：普通交易，与合约没关系，需要转账之后，等待交易是成功还是失败,
	// tx.wait();
	// 1. 实现转账，获取转账的hash（交易完成后才有hash）
	// 2. 通过hash查询交易是否成功
	// 创建转账交易
	cfg := config.Get()
	auth := s.clientInfo.Auth
	ethClient := s.clientInfo.Client
	toAddress := common.HexToAddress(addr)
	contractAddr := common.HexToAddress(cfg.BlockchainConfig.Contracts.Token)
	mtkContract, err := mtk.NewContracts(contractAddr, ethClient)

	if err != nil {
		return nil, fmt.Errorf("contract create failed: %v\n", err)
	}

	tx, err := mtkContract.Transfer(auth, toAddress, amount)
	if tx == nil || err != nil {
		return nil, fmt.Errorf("transfer failed: %v\n", err)
	}
	sym, _ := mtkContract.Symbol(&bind.CallOpts{})
	decimal, _ := mtkContract.Decimals(&bind.CallOpts{})
	res = &dto.ERCRes{Hash: tx.Hash().Hex(), Symbol: sym, Decimal: decimal, BlockNumber: nil}
	fmt.Println("TransactionService SendErc20:--- ", res.Hash)
	return res, nil
}

func (s *TransactionService) SendBNB(addr string, amount *big.Int) (res *dto.ERCRes, err error) {
	// 方案一：普通交易，与合约没关系，需要转账之后，等待交易是成功还是失败,
	// tx.wait();
	// 1. 实现转账，获取转账的hash（交易完成后才有hash）
	// 2. 通过hash查询交易是否成功
	ethClient := s.clientInfo.Client
	fromAddress := s.clientInfo.FromAddress
	toAddress := common.HexToAddress(addr)
	// 1. 准备交易参数
	// 1.1 动态获取nonce
	initialNonce, err := ethClient.PendingNonceAt(context.Background(), fromAddress)
	// 1.2 获取当前Gas价格（BSC网络）
	gasPrice, err := ethClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("retrieve gas price failed: %v", err)
	}
	// 1.3 BNB普通转账固定GasLimit为21000
	gasLimit := uint64(21000)
	// 2. 创建BNB转账交易（普通交易，不涉及合约）
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    initialNonce,
		To:       &toAddress,
		Value:    amount,   // 转账金额（wei）
		Gas:      gasLimit, // 固定值
		GasPrice: gasPrice,
		Data:     nil, // 无数据
	})
	// 3. 签名交易（使用BSC链ID）
	signedTx, err := types.SignTx(tx, types.NewLondonSigner(s.clientInfo.ChainID), s.clientInfo.PrivateKey)
	if signedTx == nil || err != nil {
		return nil, fmt.Errorf("signed Tx failed: %v", err)
	}
	// 4. 发送交易到BSC网络
	if err := ethClient.SendTransaction(context.Background(), signedTx); err != nil {
		return nil, fmt.Errorf("send transaction failed: %v", err)
	}
	receipt, err := checkTxStatus(ethClient, signedTx.Hash())
	if receipt == nil || receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("check transaction failed: %v", err)
	}
	res = &dto.ERCRes{Hash: tx.Hash().Hex(), Symbol: "BNB", BlockNumber: receipt.BlockNumber}
	fmt.Println("TransactionService SendBNB:--- ", res.Hash)
	return res, nil

	// 方案二：使用预签名生成hash的方式发送交易，转账之前提前生成了hash
	// 好处是：因为交易上链稳定是一个比较耗时的操作，大多数方式是发完交易区一直等待稳定，
	// 获取交易状态，一笔交易整体耗时可能是3~5s。假如现在有10000笔交易需要执行。
	// 需要有一个错误处理脚本去轮询失败的交易，重新发放，通过hash查询交易状态，如果失败，重新发放；getTransactionReceipt方法
}

func checkTxStatus(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("transfer timeout")
		case <-ticker.C:
			receipt, err := client.TransactionReceipt(ctx, txHash)
			if err == nil {
				return receipt, nil
			}
			// 忽略"交易未找到"的错误（可能还在pending）
			if err != ethereum.NotFound {
				return nil, err
			}
		}
	}
}
