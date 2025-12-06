package listener

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"staking-interaction/adapter"
	"staking-interaction/common/config"
	stakeContract "staking-interaction/contracts/stake"
	"staking-interaction/dto"
	"staking-interaction/model"
	"staking-interaction/repository"
	"strings"
	"time"
)

var (
	client    *ethclient.Client // 保存客户端实例，用于关闭
	isRunning bool              // 标记扫块循环是否运行
)

func ListenToEvents() {
	conf := config.Get()
	clientInfo, err := adapter.NewSyncEthClient()
	if err != nil {
		log.Fatalf("ListenToStakedEvent: Failed to connect to the BSC network: %v", err)
	}
	defer clientInfo.CloseSyncEthClient()
	client := clientInfo.Client

	contractAddress := common.HexToAddress(conf.BlockchainConfig.Contracts.Stake)

	contractABI, err := abi.JSON(strings.NewReader(stakeContract.ContractsMetaData.ABI))
	if err != nil {
		log.Fatalf("ListenToStakedEvent: Failed to parse contract ABI: %v", err)
	}
	isRunning = true
	stakedEventId := contractABI.Events[config.StakedEventName].ID
	withdrawnEventId := contractABI.Events[config.WithdrawnEventName].ID

	startBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalf("ListenToStakedEvent: Failed to get block number: %v", err)
	}
	fmt.Println("start block number:", startBlock)

	for isRunning {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		currentBlock, err := client.BlockNumber(ctx)
		if err != nil {
			log.Printf("ListenToStakedEvent: Failed to get block number, will retry after 10 sec: %v", err)
			cancel()
			time.Sleep(10 * time.Second)
			continue
		}
		endBlock := startBlock + uint64(conf.BlockchainConfig.Sync.BatchSize)
		if endBlock > currentBlock {
			endBlock = currentBlock
		}

		if endBlock > startBlock {
			log.Printf("Scanning blocks: %d ~ %d", startBlock, endBlock)
			logs, err := client.FilterLogs(ctx, ethereum.FilterQuery{
				FromBlock: big.NewInt(int64(startBlock)),
				ToBlock:   big.NewInt(int64(endBlock)),
				Addresses: []common.Address{contractAddress},
				Topics:    [][]common.Hash{{stakedEventId, withdrawnEventId}},
			})
			if err != nil {
				log.Printf("ListenToStakedEvent:  Failed to filter logs, will retry after 10 sec: BlockNumber is: %d, err is: %v", startBlock, err)
				cancel()
				time.Sleep(10 * time.Second)
				continue
			}

			for _, l := range logs {
				handleLog(l, dto.StakeEventListener{
					StakedEventId:    stakedEventId,
					WithdrawnEventId: withdrawnEventId,
					ContractAddress:  contractAddress,
					ContractABI:      contractABI,
				})
			}

			startBlock = endBlock + 1
			cancel()
		} else {
			cancel()
			time.Sleep(10 * time.Second)
		}

	}

}

func handleLog(l types.Log, listener dto.StakeEventListener) {
	switch l.Topics[0] {
	case listener.StakedEventId:
		var event stakeContract.ContractsStaked
		if err := listener.ContractABI.UnpackIntoInterface(&event, config.StakedEventName, l.Data); err != nil {
			log.Printf("ListenToStakedEvent: Staked failed to unpack event: %v", err)
			return
		}

		if len(l.Topics) >= 2 {
			event.User = common.HexToAddress(l.Topics[1].Hex())
		} else {
			log.Printf("ListenToStakedEvent: Staked missing user topic (tx: %s)", l.TxHash.Hex())
			return
		}
		fmt.Printf("Staked: block number: %d, transaction hash:%v, user:%v, amount:%s, timestamp=%d,  StakedIndex=%s\n",
			l.BlockNumber,
			l.TxHash.Hex(),
			event.User,
			event.Amount.String(),
			event.Timestamp.Int64(),
			event.StakeIndex.String(),
		)

		stake := model.Stake{
			IndexNum:        event.StakeIndex.String(),
			Hash:            l.TxHash.Hex(),
			ContractAddress: listener.ContractAddress.Hex(),
			FromAddress:     l.Address.Hex(),
			Method:          config.StakedEventName,
			Amount:          event.Amount.String(),
			BlockNumber:     int64(l.BlockNumber),
			Status:          0,
			Timestamp:       time.Now(),
		}

		storeStakeInfo(stake)

	case listener.WithdrawnEventId:
		var event stakeContract.ContractsWithdrawn
		if err := listener.ContractABI.UnpackIntoInterface(&event, config.WithdrawnEventName, l.Data); err != nil {
			log.Printf("ListenToStakedEvent: Withdrawn failed to unpack event: %v", err)
			return
		}

		if len(l.Topics) >= 2 {
			event.User = common.HexToAddress(l.Topics[1].Hex())
		} else {
			log.Printf("ListenToStakedEvent: Withdrawn missing user topic (tx: %s)", l.TxHash.Hex())
			return
		}

		fmt.Printf("Withdrawn: block number: %d, transaction hash:%v, user:%v, amount:%s, StakedIndex=%s\n",
			l.BlockNumber,
			l.TxHash.Hex(),
			event.User,
			event.TotalAmount.String(),
			event.StakeIndex.String(),
		)

		stake := model.Stake{
			IndexNum:        event.StakeIndex.String(),
			Hash:            l.TxHash.Hex(),
			ContractAddress: listener.ContractAddress.Hex(),
			FromAddress:     l.Address.Hex(),
			Method:          config.WithdrawnEventName,
			Amount:          event.TotalAmount.String(),
			BlockNumber:     int64(l.BlockNumber),
			Status:          1,
			Timestamp:       time.Now(),
		}

		storeStakeInfo(stake)
	}

}

func CloseListener() {
	if !isRunning {
		return
	}
	isRunning = false
	if client != nil {
		client.Close()
	}
	log.Printf("ListenToStakedEvent: Listener closed")
}

func storeStakeInfo(stake model.Stake) {
	repository.AddStakeInfo(stake)
}
