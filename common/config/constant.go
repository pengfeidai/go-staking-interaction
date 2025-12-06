package config

const (
	TIME_FORMAT = "2006-01-02 15:04:05"
)

// BillType 账单类型
const (
	BillTypeRecharge   = 1
	BillTypeWithdrawal = 2
)

// WithdrawStatus 提现状态
const (
	WithdrawStatusInit    = 1
	WithdrawStatusPending = 2
	WithdrawStatusSuccess = 3
	WithdrawStatusFailed  = 4
)

// Token 类型
const (
	TokenTypeBNB = 1
	TokenTypeMTK = 2
)

// 事件名称
const (
	StakedEventName    = "Staked"
	WithdrawnEventName = "Withdrawn"
)
