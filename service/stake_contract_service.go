package service

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"staking-interaction/adapter"
	"staking-interaction/common/config"
	"staking-interaction/contracts/stake"
	"staking-interaction/dto"
	"staking-interaction/model"
	"staking-interaction/repository"
)

type StakeService struct {
	clientInfo *adapter.InitClient
}

func NewStakeService(
	clientInfo *adapter.InitClient,
) *StakeService {
	return &StakeService{
		clientInfo: clientInfo,
	}
}

var cfg = config.Get()
var stakeContractAddr = common.HexToAddress(cfg.BlockchainConfig.Contracts.Stake)

func (s *StakeService) NewStakeContract() (*stake.Contracts, error) {
	//creates a new instance of Contracts, bound to a specific deployed contract
	stakingContract, err := stake.NewContracts(stakeContractAddr, s.clientInfo.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create staking contract")
	}

	return stakingContract, nil
}

func (s *StakeService) Stake(amount int64, period uint8) (response *dto.StakeResponse, err error) {
	stakingContract, err := s.NewStakeContract()
	if err != nil {
		return nil, fmt.Errorf("failed to create staking contract: %v", err)
	}
	auth := s.clientInfo.Auth

	trans, err := stakingContract.Stake(
		auth,
		big.NewInt(amount),
		period,
	)

	if trans == nil || err != nil {
		return nil, fmt.Errorf("Stake transaction error: %w", err)
	}

	response = &dto.StakeResponse{
		Hash:            trans.Hash().String(),
		ContractAddress: stakeContractAddr.String(),
		FromAddress:     s.clientInfo.FromAddress,
		Method:          "stake",
	}

	return response, nil
}

func (s *StakeService) Withdraw(index *big.Int) (response *dto.StakeResponse, err error) {
	stakingContract, err := s.NewStakeContract()
	if err != nil {
		return nil, fmt.Errorf("failed to create staking contract: %v", err)
	}
	auth := s.clientInfo.Auth

	trans, err := stakingContract.Withdraw(auth, index)

	if trans == nil || err != nil {
		return nil, fmt.Errorf("withdraw transaction error: %w", err)
	}

	response = &dto.StakeResponse{
		Hash:            trans.Hash().String(),
		ContractAddress: stakeContractAddr.String(),
		FromAddress:     s.clientInfo.FromAddress,
		Method:          "withdraw",
	}

	return response, nil
}

func GetAllStakesByFromAddress(id string) (response model.Stake) {
	stake := repository.GetAllStakesByFromAddress(id)
	return stake
}
