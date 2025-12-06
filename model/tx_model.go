package model

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type Receipt struct {
	Type              uint8          // 交易类型（如 0x0 表示普通交易）
	PostState         []byte         // 交易执行后的账户状态根（已废弃，用 Status 替代）
	Status            uint64         // 交易状态：1=成功（0x1），0=失败（0x0）
	CumulativeGasUsed uint64         // 从区块开始到该交易的累计燃气消耗
	GasUsed           uint64         // 该交易实际消耗的燃气
	ContractAddress   common.Address // 若为创建合约交易，此处是新合约地址（否则为空）
	Logs              []*types.Log   // 交易触发的日志（Events，如 ERC20 的 Transfer 事件）
	LogsBloom         types.Bloom    // 日志布隆过滤器（用于快速查询日志）
	TransactionHash   common.Hash    // 对应的交易哈希
	BlockHash         common.Hash    // 交易所在区块的哈希
	BlockNumber       *big.Int       // 交易所在区块的高度
	TransactionIndex  uint           // 交易在区块中的索引（从 0 开始）
}

type Log struct {
	// Consensus fields:
	// address of the contract that generated the event
	Address common.Address `json:"address" gencodec:"required"`
	// list of topics provided by the contract.
	// Topics[0]：事件签名哈希（即事件类型的唯一标识），Topics[1] ~ Topics[3]：事件定义里 indexed 的参数
	Topics []common.Hash `json:"topics" gencodec:"required"`
	// supplied by the contract, usually ABI-encoded
	Data []byte `json:"data" gencodec:"required"`

	// Derived fields. These fields are filled in by the node
	// but not secured by consensus.
	// block in which the transaction was included
	BlockNumber uint64 `json:"blockNumber" rlp:"-"`
	// hash of the transaction
	TxHash common.Hash `json:"transactionHash" gencodec:"required" rlp:"-"`
	// index of the transaction in the block
	TxIndex uint `json:"transactionIndex" rlp:"-"`
	// hash of the block in which the transaction was included
	BlockHash common.Hash `json:"blockHash" rlp:"-"`
	// timestamp of the block in which the transaction was included
	BlockTimestamp uint64 `json:"blockTimestamp" rlp:"-"`
	// index of the logger in the block
	Index uint `json:"logIndex" rlp:"-"`

	// The Removed field is true if this logger was reverted due to a chain reorganisation.
	// You must pay attention to this field if you receive logs through a filter query.
	Removed bool `json:"removed" rlp:"-"`
}
type Transaction struct {
	data txdata // 核心数据字段（包含以下信息）
}

// txdata 是 Transaction 的底层数据结构
type txdata struct {
	AccountNonce uint64          // 发送者的交易序号（防止重放攻击）
	Price        *big.Int        // 燃气价格（单位：Wei）
	GasLimit     uint64          // 最大燃气限制
	Recipient    *common.Address // 接收地址（普通转账/合约调用的目标地址，nil 表示创建合约）
	Amount       *big.Int        // 转账金额（仅普通 ETH/BNB 转账有效，合约转账金额在 Data 中）
	Data         []byte          // 核心：合约调用数据（ABI 编码）或空（普通转账）
	V, R, S      *big.Int        // 签名信息（用于验证发送者）
}
