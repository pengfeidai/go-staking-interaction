package service

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
	"math/big"
	"staking-interaction/adapter"
	"staking-interaction/common/config"
	"staking-interaction/contracts/airdrop"
	"staking-interaction/dto"
	"staking-interaction/utils"
	"sync"
	"time"
)

type AirdropService struct {
	clientInfo *adapter.InitClient
	log        *logrus.Logger
}

func NewAirdropService(
	clientInfo *adapter.InitClient,
	log *logrus.Logger,
) *AirdropService {
	return &AirdropService{
		clientInfo: clientInfo,
		log:        logrus.New(),
	}
}

var conf = config.Get()
var airdropContractAddr = common.HexToAddress(conf.BlockchainConfig.Contracts.Airdrop)

func (s *AirdropService) NewAirdropContract() (*airdrop.Contracts, error) {
	airdropContract, err := airdrop.NewContracts(airdropContractAddr, s.clientInfo.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create airdroperc contract")
	}

	if airdropContract == nil {
		return nil, fmt.Errorf("airdropContract should not be nil")
	}
	return airdropContract, nil
}

func (s *AirdropService) AirdropERC20(reqCount int, reqBatchSize int, reqAmount []*big.Int) (data *dto.AirdropResponse, err error) {
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex // 保护response切片的并发写入
		responses []dto.AirdropInfo
	)
	s.log.WithFields(map[string]interface{}{
		"service":      "airdrop",
		"from address": s.clientInfo.FromAddress,
	}).Info("airdropERC20 init success")

	contract, err := s.NewAirdropContract()
	if err != nil {
		return nil, fmt.Errorf("new contract failed: %v", err)
	}

	// generate multiple wallets
	walletAddresses, err := GetMultiWallets(reqCount)
	if err != nil || walletAddresses == nil || len(walletAddresses) == 0 {
		return nil, fmt.Errorf("generate wallet failed: %v", err)
	}

	addressLen := len(walletAddresses)
	fromAddr := s.clientInfo.FromAddress
	ethClient := s.clientInfo.Client
	auth := s.clientInfo.Auth

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 设置整体超时
	defer cancel()

	initialNonce, err := ethClient.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return nil, fmt.Errorf("retrieve nonce failed: %v", err)
	}

	batchNum := (addressLen + reqBatchSize - 1) / reqBatchSize
	startIndex := 0
	currentNonce := initialNonce
	for i := 0; i < batchNum && startIndex < addressLen; i++ {
		endIndex := startIndex + reqBatchSize
		if endIndex > addressLen {
			endIndex = addressLen
		}

		batchAddress := walletAddresses[startIndex:endIndex]
		batchAmounts := reqAmount[startIndex:endIndex]
		batchNonce := currentNonce
		currentNonce++

		wg.Add(1)
		go func(idx int, addresses []common.Address, amounts []*big.Int, nonce uint64) {
			defer wg.Done()
			batchAuth := *auth
			batchAuth.Nonce = big.NewInt(int64(nonce))
			batchAuth.Context = ctx

			res := s.processAirdropERC20(idx, addresses, amounts, &batchAuth, contract)
			// 线程安全地收集结果
			mu.Lock()
			responses = append(responses, res)
			mu.Unlock()
		}(i, batchAddress, batchAmounts, batchNonce)

		startIndex = endIndex
	}

	done := make(chan struct{})
	// 等待所有批次完成并返回结果
	go func() {
		wg.Wait()
		close(done)
	}()

	// 等待结果或超时
	select {
	case <-done:
		successCount := 0
		failedCount := 0
		for _, res := range responses {
			if res.Error == "" {
				successCount++
			} else {
				failedCount++
			}
		}
		resInfo := dto.AirdropResponse{
			Msg:              "airdrop erc success!",
			CompletedBatches: len(responses),
			SuccessBatches:   successCount,
			FailBatches:      failedCount,
			Data:             responses}
		return &resInfo, nil
	case <-ctx.Done():
		resInfo := dto.AirdropResponse{
			Msg:              "airdrop erc timeout",
			CompletedBatches: len(responses),
			Data:             responses,
			Error:            ctx.Err().Error(),
		}
		return &resInfo, nil
	}
}

func (s *AirdropService) processAirdropERC20(idx int, batchAddress []common.Address, batchAmounts []*big.Int, auth *bind.TransactOpts, contract *airdrop.Contracts) (response dto.AirdropInfo) {
	fromAddr := s.clientInfo.FromAddress
	trans, err := contract.AirdropERC20(auth, batchAddress, batchAmounts)
	if trans == nil || err != nil {
		return dto.AirdropInfo{
			BatchNum:        idx,
			Error:           fmt.Sprintf("airdroperc failed: %v", err),
			ContractAddress: airdropContractAddr.String(),
			FromAddress:     fromAddr,
			WalletAddress:   batchAddress,
		}
	} else {
		return dto.AirdropInfo{
			BatchNum:        idx,
			Hash:            trans.Hash().Hex(),
			ContractAddress: airdropContractAddr.String(),
			FromAddress:     fromAddr,
			WalletAddress:   batchAddress,
			Error:           "",
		}
	}

}

func (s *AirdropService) AirdropBNB(reqCount int, reqBatchSize int, reqAmount []*big.Int) (data *dto.AirdropResponse, err error) {
	var (
		responses []dto.AirdropInfo
		wg        sync.WaitGroup
		mu        sync.Mutex // 保护response切片的并发写入
	)

	s.log.WithFields(map[string]interface{}{
		"service":      "airdrop",
		"from address": s.clientInfo.FromAddress,
	}).Info("AirdropBNB init success")

	contract, err := s.NewAirdropContract()
	if err != nil {
		return nil, fmt.Errorf("new contract failed: %v", err)
	}

	// generate multiple wallets
	walletAddresses, err := GetMultiWallets(reqCount)
	if err != nil || walletAddresses == nil || len(walletAddresses) == 0 {
		return nil, fmt.Errorf("generate wallet failed: %v", err)
	}

	addressLen := len(walletAddresses)
	fromAddr := s.clientInfo.FromAddress
	ethClient := s.clientInfo.Client
	auth := s.clientInfo.Auth

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 设置整体超时
	defer cancel()

	initialNonce, err := ethClient.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return nil, fmt.Errorf("retrieve nonce failed: %v", err)
	}

	batchNum := (addressLen + reqBatchSize - 1) / reqBatchSize
	startIndex := 0
	currentNonce := initialNonce
	for i := 0; i < batchNum && startIndex < addressLen; i++ {
		endIndex := startIndex + reqBatchSize
		if endIndex > addressLen {
			endIndex = addressLen
		}

		batchAddress := walletAddresses[startIndex:endIndex]
		batchAmounts := reqAmount[startIndex:endIndex]
		batchNonce := currentNonce
		currentNonce++

		wg.Add(1)
		go func(idx int, addresses []common.Address, amounts []*big.Int, nonce uint64) {
			defer wg.Done()
			val := utils.CalculateSumOfAmounts(amounts)

			batchAuth := *auth
			batchAuth.Nonce = big.NewInt(int64(nonce))
			batchAuth.Context = ctx
			batchAuth.Value = val

			res := s.processAirdropBNB(idx, addresses, amounts, &batchAuth, contract)
			// 线程安全地收集结果
			mu.Lock()
			responses = append(responses, res)
			mu.Unlock()
		}(i, batchAddress, batchAmounts, batchNonce)

		startIndex = endIndex
	}

	done := make(chan struct{})
	// 等待所有批次完成并返回结果
	go func() {
		wg.Wait()
		close(done)
	}()

	// 等待结果或超时
	select {
	case <-done:
		successCount := 0
		failedCount := 0
		for _, res := range responses {
			if res.Error == "" {
				successCount++
			} else {
				failedCount++
			}
		}
		resInfo := dto.AirdropResponse{
			Msg:              "airdrop bnb success!",
			CompletedBatches: len(responses),
			SuccessBatches:   successCount,
			FailBatches:      failedCount,
			Data:             responses}
		return &resInfo, nil
	case <-ctx.Done():
		resInfo := dto.AirdropResponse{
			Msg:              "airdrop bnb timeout",
			CompletedBatches: len(responses),
			Data:             responses,
			Error:            ctx.Err().Error(),
		}
		return &resInfo, nil
	}

}

func (s *AirdropService) processAirdropBNB(idx int, batchAddress []common.Address, batchAmounts []*big.Int, auth *bind.TransactOpts, contract *airdrop.Contracts) (response dto.AirdropInfo) {
	fromAddr := s.clientInfo.FromAddress
	trans, err := contract.AirdropBNB(auth, batchAddress, batchAmounts)
	if trans == nil || err != nil {
		return dto.AirdropInfo{
			BatchNum:        idx,
			Error:           fmt.Sprintf("airdrop bnb failed: %v", err),
			ContractAddress: airdropContractAddr.String(),
			FromAddress:     fromAddr,
			WalletAddress:   batchAddress,
		}
	} else {
		return dto.AirdropInfo{
			BatchNum:        idx,
			Hash:            trans.Hash().Hex(),
			ContractAddress: airdropContractAddr.String(),
			FromAddress:     fromAddr,
			WalletAddress:   batchAddress,
			Error:           "",
		}
	}
}

func GetMultiWallets(count int) (walletAddresses []common.Address, err error) {
	for i := 0; i < count; i++ {
		// 1. 生成随机私钥（secp256k1曲线）
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			return nil, err
		}

		// 2. 从私钥推导公钥（未压缩格式）
		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, err
		}

		// 3. 从公钥推导以太坊地址（0x前缀）
		address := crypto.PubkeyToAddress(*publicKeyECDSA)

		walletAddresses = append(walletAddresses, address)
	}
	return walletAddresses, nil
}
