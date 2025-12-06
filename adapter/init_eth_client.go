package adapter

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"staking-interaction/common/config"
)

type InitClient struct {
	Auth        *bind.TransactOpts
	FromAddress common.Address
	Client      *ethclient.Client
	PrivateKey  *ecdsa.PrivateKey
	ChainID     *big.Int
}

func NewInitEthClient() (*InitClient, error) {
	cfg := config.Get()
	// 初始化客户端
	ethClient, err := ethclient.Dial(cfg.BlockchainConfig.RawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the stake contract: %v", err)
	}

	// 加载私钥
	privateKey, err := crypto.HexToECDSA(cfg.BlockchainConfig.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 获取链ID
	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	// create a transaction signer from a single private key.
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorized transactor: %v", err)
	}

	return &InitClient{
		Auth:        auth,
		Client:      ethClient,
		FromAddress: fromAddress,
		PrivateKey:  privateKey,
		ChainID:     chainID,
	}, nil
}

func (c *InitClient) CloseEthClient() {
	c.Client.Close()
}

func NewSyncEthClient() (*InitClient, error) {
	cfg := config.Get()
	// 初始化客户端
	ethClient, err := ethclient.Dial(cfg.BlockchainConfig.RpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the stake contract: %v", err)
	}

	// 获取链ID
	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	return &InitClient{
		Client:  ethClient,
		ChainID: chainID,
	}, nil
}

func (c *InitClient) CloseSyncEthClient() {
	c.Client.Close()
}
