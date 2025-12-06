// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package stake

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// StakeContractsStake is an auto generated low-level Go binding around an user-defined struct.
type StakeContractsStake struct {
	Amount     *big.Int
	StartTime  *big.Int
	EndTime    *big.Int
	RewardRate *big.Int
	IsActive   bool
	StakeIndex *big.Int
}

// ContractsMetaData contains all meta data concerning the Contracts contract.
var ContractsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"_mtkToken\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"enumStakeContracts.StakingPeriod\",\"name\":\"period\",\"type\":\"uint8\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"stakeIndex\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"Staked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"stakeIndex\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"totalAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"standAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"reward\",\"type\":\"uint256\"}],\"name\":\"Withdrawn\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"enumStakeContracts.StakingPeriod\",\"name\":\"\",\"type\":\"uint8\"}],\"name\":\"apy\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"}],\"name\":\"getUserActiveStakes\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardRate\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isActive\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"stakeIndex\",\"type\":\"uint256\"}],\"internalType\":\"structStakeContracts.Stake[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"enumStakeContracts.StakingPeriod\",\"name\":\"period\",\"type\":\"uint8\"}],\"name\":\"stake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"stakeOwnerMapping\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stakingToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"userStakeIndexes\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"userStakes\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"startTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endTime\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardRate\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isActive\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"stakeIndex\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"stakeIndex\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ContractsABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractsMetaData.ABI instead.
var ContractsABI = ContractsMetaData.ABI

// Contracts is an auto generated Go binding around an Ethereum contract.
type Contracts struct {
	ContractsCaller     // Read-only binding to the contract
	ContractsTransactor // Write-only binding to the contract
	ContractsFilterer   // Log filterer for contract events
}

// ContractsCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractsFilterer is an auto generated logger filtering Go binding around an Ethereum contract events.
type ContractsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractsSession struct {
	Contract     *Contracts        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ContractsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractsCallerSession struct {
	Contract *ContractsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// ContractsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractsTransactorSession struct {
	Contract     *ContractsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// ContractsRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractsRaw struct {
	Contract *Contracts // Generic contract binding to access the raw methods on
}

// ContractsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractsCallerRaw struct {
	Contract *ContractsCaller // Generic read-only contract binding to access the raw methods on
}

// ContractsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractsTransactorRaw struct {
	Contract *ContractsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContracts creates a new instance of Contracts, bound to a specific deployed contract.
func NewContracts(address common.Address, backend bind.ContractBackend) (*Contracts, error) {
	contract, err := bindContracts(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Contracts{ContractsCaller: ContractsCaller{contract: contract}, ContractsTransactor: ContractsTransactor{contract: contract}, ContractsFilterer: ContractsFilterer{contract: contract}}, nil
}

// NewContractsCaller creates a new read-only instance of Contracts, bound to a specific deployed contract.
func NewContractsCaller(address common.Address, caller bind.ContractCaller) (*ContractsCaller, error) {
	contract, err := bindContracts(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractsCaller{contract: contract}, nil
}

// NewContractsTransactor creates a new write-only instance of Contracts, bound to a specific deployed contract.
func NewContractsTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractsTransactor, error) {
	contract, err := bindContracts(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractsTransactor{contract: contract}, nil
}

// NewContractsFilterer creates a new logger filterer instance of Contracts, bound to a specific deployed contract.
func NewContractsFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractsFilterer, error) {
	contract, err := bindContracts(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractsFilterer{contract: contract}, nil
}

// bindContracts binds a generic wrapper to an already deployed contract.
func bindContracts(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractsMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contracts *ContractsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contracts.Contract.ContractsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contracts *ContractsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contracts.Contract.ContractsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contracts *ContractsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contracts.Contract.ContractsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Contracts *ContractsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Contracts.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Contracts *ContractsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Contracts.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Contracts *ContractsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Contracts.Contract.contract.Transact(opts, method, params...)
}

// Apy is a free data retrieval call binding the contract method 0x1f1accb2.
//
// Solidity: function apy(uint8 ) view returns(uint256)
func (_Contracts *ContractsCaller) Apy(opts *bind.CallOpts, arg0 uint8) (*big.Int, error) {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "apy", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Apy is a free data retrieval call binding the contract method 0x1f1accb2.
//
// Solidity: function apy(uint8 ) view returns(uint256)
func (_Contracts *ContractsSession) Apy(arg0 uint8) (*big.Int, error) {
	return _Contracts.Contract.Apy(&_Contracts.CallOpts, arg0)
}

// Apy is a free data retrieval call binding the contract method 0x1f1accb2.
//
// Solidity: function apy(uint8 ) view returns(uint256)
func (_Contracts *ContractsCallerSession) Apy(arg0 uint8) (*big.Int, error) {
	return _Contracts.Contract.Apy(&_Contracts.CallOpts, arg0)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_Contracts *ContractsCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_Contracts *ContractsSession) Decimals() (uint8, error) {
	return _Contracts.Contract.Decimals(&_Contracts.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_Contracts *ContractsCallerSession) Decimals() (uint8, error) {
	return _Contracts.Contract.Decimals(&_Contracts.CallOpts)
}

// GetUserActiveStakes is a free data retrieval call binding the contract method 0xa262ab35.
//
// Solidity: function getUserActiveStakes(address user) view returns((uint256,uint256,uint256,uint256,bool,uint256)[])
func (_Contracts *ContractsCaller) GetUserActiveStakes(opts *bind.CallOpts, user common.Address) ([]StakeContractsStake, error) {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "getUserActiveStakes", user)

	if err != nil {
		return *new([]StakeContractsStake), err
	}

	out0 := *abi.ConvertType(out[0], new([]StakeContractsStake)).(*[]StakeContractsStake)

	return out0, err

}

// GetUserActiveStakes is a free data retrieval call binding the contract method 0xa262ab35.
//
// Solidity: function getUserActiveStakes(address user) view returns((uint256,uint256,uint256,uint256,bool,uint256)[])
func (_Contracts *ContractsSession) GetUserActiveStakes(user common.Address) ([]StakeContractsStake, error) {
	return _Contracts.Contract.GetUserActiveStakes(&_Contracts.CallOpts, user)
}

// GetUserActiveStakes is a free data retrieval call binding the contract method 0xa262ab35.
//
// Solidity: function getUserActiveStakes(address user) view returns((uint256,uint256,uint256,uint256,bool,uint256)[])
func (_Contracts *ContractsCallerSession) GetUserActiveStakes(user common.Address) ([]StakeContractsStake, error) {
	return _Contracts.Contract.GetUserActiveStakes(&_Contracts.CallOpts, user)
}

// StakeOwnerMapping is a free data retrieval call binding the contract method 0xa57a5304.
//
// Solidity: function stakeOwnerMapping(uint256 ) view returns(address)
func (_Contracts *ContractsCaller) StakeOwnerMapping(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "stakeOwnerMapping", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakeOwnerMapping is a free data retrieval call binding the contract method 0xa57a5304.
//
// Solidity: function stakeOwnerMapping(uint256 ) view returns(address)
func (_Contracts *ContractsSession) StakeOwnerMapping(arg0 *big.Int) (common.Address, error) {
	return _Contracts.Contract.StakeOwnerMapping(&_Contracts.CallOpts, arg0)
}

// StakeOwnerMapping is a free data retrieval call binding the contract method 0xa57a5304.
//
// Solidity: function stakeOwnerMapping(uint256 ) view returns(address)
func (_Contracts *ContractsCallerSession) StakeOwnerMapping(arg0 *big.Int) (common.Address, error) {
	return _Contracts.Contract.StakeOwnerMapping(&_Contracts.CallOpts, arg0)
}

// StakingToken is a free data retrieval call binding the contract method 0x72f702f3.
//
// Solidity: function stakingToken() view returns(address)
func (_Contracts *ContractsCaller) StakingToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "stakingToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakingToken is a free data retrieval call binding the contract method 0x72f702f3.
//
// Solidity: function stakingToken() view returns(address)
func (_Contracts *ContractsSession) StakingToken() (common.Address, error) {
	return _Contracts.Contract.StakingToken(&_Contracts.CallOpts)
}

// StakingToken is a free data retrieval call binding the contract method 0x72f702f3.
//
// Solidity: function stakingToken() view returns(address)
func (_Contracts *ContractsCallerSession) StakingToken() (common.Address, error) {
	return _Contracts.Contract.StakingToken(&_Contracts.CallOpts)
}

// UserStakeIndexes is a free data retrieval call binding the contract method 0xcc4a3eed.
//
// Solidity: function userStakeIndexes(address , uint256 ) view returns(uint256)
func (_Contracts *ContractsCaller) UserStakeIndexes(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "userStakeIndexes", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// UserStakeIndexes is a free data retrieval call binding the contract method 0xcc4a3eed.
//
// Solidity: function userStakeIndexes(address , uint256 ) view returns(uint256)
func (_Contracts *ContractsSession) UserStakeIndexes(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Contracts.Contract.UserStakeIndexes(&_Contracts.CallOpts, arg0, arg1)
}

// UserStakeIndexes is a free data retrieval call binding the contract method 0xcc4a3eed.
//
// Solidity: function userStakeIndexes(address , uint256 ) view returns(uint256)
func (_Contracts *ContractsCallerSession) UserStakeIndexes(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Contracts.Contract.UserStakeIndexes(&_Contracts.CallOpts, arg0, arg1)
}

// UserStakes is a free data retrieval call binding the contract method 0xb5d5b5fa.
//
// Solidity: function userStakes(address , uint256 ) view returns(uint256 amount, uint256 startTime, uint256 endTime, uint256 rewardRate, bool isActive, uint256 stakeIndex)
func (_Contracts *ContractsCaller) UserStakes(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (struct {
	Amount     *big.Int
	StartTime  *big.Int
	EndTime    *big.Int
	RewardRate *big.Int
	IsActive   bool
	StakeIndex *big.Int
}, error) {
	var out []interface{}
	err := _Contracts.contract.Call(opts, &out, "userStakes", arg0, arg1)

	outstruct := new(struct {
		Amount     *big.Int
		StartTime  *big.Int
		EndTime    *big.Int
		RewardRate *big.Int
		IsActive   bool
		StakeIndex *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Amount = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.StartTime = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.EndTime = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.RewardRate = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.IsActive = *abi.ConvertType(out[4], new(bool)).(*bool)
	outstruct.StakeIndex = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// UserStakes is a free data retrieval call binding the contract method 0xb5d5b5fa.
//
// Solidity: function userStakes(address , uint256 ) view returns(uint256 amount, uint256 startTime, uint256 endTime, uint256 rewardRate, bool isActive, uint256 stakeIndex)
func (_Contracts *ContractsSession) UserStakes(arg0 common.Address, arg1 *big.Int) (struct {
	Amount     *big.Int
	StartTime  *big.Int
	EndTime    *big.Int
	RewardRate *big.Int
	IsActive   bool
	StakeIndex *big.Int
}, error) {
	return _Contracts.Contract.UserStakes(&_Contracts.CallOpts, arg0, arg1)
}

// UserStakes is a free data retrieval call binding the contract method 0xb5d5b5fa.
//
// Solidity: function userStakes(address , uint256 ) view returns(uint256 amount, uint256 startTime, uint256 endTime, uint256 rewardRate, bool isActive, uint256 stakeIndex)
func (_Contracts *ContractsCallerSession) UserStakes(arg0 common.Address, arg1 *big.Int) (struct {
	Amount     *big.Int
	StartTime  *big.Int
	EndTime    *big.Int
	RewardRate *big.Int
	IsActive   bool
	StakeIndex *big.Int
}, error) {
	return _Contracts.Contract.UserStakes(&_Contracts.CallOpts, arg0, arg1)
}

// Stake is a paid mutator transaction binding the contract method 0x10087fb1.
//
// Solidity: function stake(uint256 amount, uint8 period) returns()
func (_Contracts *ContractsTransactor) Stake(opts *bind.TransactOpts, amount *big.Int, period uint8) (*types.Transaction, error) {
	return _Contracts.contract.Transact(opts, "stake", amount, period)
}

// Stake is a paid mutator transaction binding the contract method 0x10087fb1.
//
// Solidity: function stake(uint256 amount, uint8 period) returns()
func (_Contracts *ContractsSession) Stake(amount *big.Int, period uint8) (*types.Transaction, error) {
	return _Contracts.Contract.Stake(&_Contracts.TransactOpts, amount, period)
}

// Stake is a paid mutator transaction binding the contract method 0x10087fb1.
//
// Solidity: function stake(uint256 amount, uint8 period) returns()
func (_Contracts *ContractsTransactorSession) Stake(amount *big.Int, period uint8) (*types.Transaction, error) {
	return _Contracts.Contract.Stake(&_Contracts.TransactOpts, amount, period)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 stakeIndex) returns()
func (_Contracts *ContractsTransactor) Withdraw(opts *bind.TransactOpts, stakeIndex *big.Int) (*types.Transaction, error) {
	return _Contracts.contract.Transact(opts, "withdraw", stakeIndex)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 stakeIndex) returns()
func (_Contracts *ContractsSession) Withdraw(stakeIndex *big.Int) (*types.Transaction, error) {
	return _Contracts.Contract.Withdraw(&_Contracts.TransactOpts, stakeIndex)
}

// Withdraw is a paid mutator transaction binding the contract method 0x2e1a7d4d.
//
// Solidity: function withdraw(uint256 stakeIndex) returns()
func (_Contracts *ContractsTransactorSession) Withdraw(stakeIndex *big.Int) (*types.Transaction, error) {
	return _Contracts.Contract.Withdraw(&_Contracts.TransactOpts, stakeIndex)
}

// ContractsStakedIterator is returned from FilterStaked and is used to iterate over the raw logs and unpacked data for Staked events raised by the Contracts contract.
type ContractsStakedIterator struct {
	Event *ContractsStaked // Event containing the contract specifics and raw logger

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractsStakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractsStaked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractsStaked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractsStakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractsStakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractsStaked represents a Staked event raised by the Contracts contract.
type ContractsStaked struct {
	User       common.Address
	Amount     *big.Int
	Period     uint8
	StakeIndex *big.Int
	Timestamp  *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterStaked is a free logger retrieval operation binding the contract event 0x022dd619ee0d92140e5abde105761d6df71c05c4fcb761610ea0552064f91e6c.
//
// Solidity: event Staked(address indexed user, uint256 amount, uint8 period, uint256 stakeIndex, uint256 timestamp)
func (_Contracts *ContractsFilterer) FilterStaked(opts *bind.FilterOpts, user []common.Address) (*ContractsStakedIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Contracts.contract.FilterLogs(opts, "Staked", userRule)
	if err != nil {
		return nil, err
	}
	return &ContractsStakedIterator{contract: _Contracts.contract, event: "Staked", logs: logs, sub: sub}, nil
}

// WatchStaked is a free logger subscription operation binding the contract event 0x022dd619ee0d92140e5abde105761d6df71c05c4fcb761610ea0552064f91e6c.
//
// Solidity: event Staked(address indexed user, uint256 amount, uint8 period, uint256 stakeIndex, uint256 timestamp)
func (_Contracts *ContractsFilterer) WatchStaked(opts *bind.WatchOpts, sink chan<- *ContractsStaked, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Contracts.contract.WatchLogs(opts, "Staked", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New logger arrived, parse the event and forward to the user
				event := new(ContractsStaked)
				if err := _Contracts.contract.UnpackLog(event, "Staked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseStaked is a logger parse operation binding the contract event 0x022dd619ee0d92140e5abde105761d6df71c05c4fcb761610ea0552064f91e6c.
//
// Solidity: event Staked(address indexed user, uint256 amount, uint8 period, uint256 stakeIndex, uint256 timestamp)
func (_Contracts *ContractsFilterer) ParseStaked(log types.Log) (*ContractsStaked, error) {
	event := new(ContractsStaked)
	if err := _Contracts.contract.UnpackLog(event, "Staked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractsWithdrawnIterator is returned from FilterWithdrawn and is used to iterate over the raw logs and unpacked data for Withdrawn events raised by the Contracts contract.
type ContractsWithdrawnIterator struct {
	Event *ContractsWithdrawn // Event containing the contract specifics and raw logger

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractsWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractsWithdrawn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractsWithdrawn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractsWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractsWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractsWithdrawn represents a Withdrawn event raised by the Contracts contract.
type ContractsWithdrawn struct {
	User        common.Address
	StakeIndex  *big.Int
	TotalAmount *big.Int
	StandAmount *big.Int
	Reward      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterWithdrawn is a free logger retrieval operation binding the contract event 0x94ffd6b85c71b847775c89ef6496b93cee961bdc6ff827fd117f174f06f745ae.
//
// Solidity: event Withdrawn(address indexed user, uint256 stakeIndex, uint256 totalAmount, uint256 standAmount, uint256 reward)
func (_Contracts *ContractsFilterer) FilterWithdrawn(opts *bind.FilterOpts, user []common.Address) (*ContractsWithdrawnIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Contracts.contract.FilterLogs(opts, "Withdrawn", userRule)
	if err != nil {
		return nil, err
	}
	return &ContractsWithdrawnIterator{contract: _Contracts.contract, event: "Withdrawn", logs: logs, sub: sub}, nil
}

// WatchWithdrawn is a free logger subscription operation binding the contract event 0x94ffd6b85c71b847775c89ef6496b93cee961bdc6ff827fd117f174f06f745ae.
//
// Solidity: event Withdrawn(address indexed user, uint256 stakeIndex, uint256 totalAmount, uint256 standAmount, uint256 reward)
func (_Contracts *ContractsFilterer) WatchWithdrawn(opts *bind.WatchOpts, sink chan<- *ContractsWithdrawn, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _Contracts.contract.WatchLogs(opts, "Withdrawn", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New logger arrived, parse the event and forward to the user
				event := new(ContractsWithdrawn)
				if err := _Contracts.contract.UnpackLog(event, "Withdrawn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdrawn is a logger parse operation binding the contract event 0x94ffd6b85c71b847775c89ef6496b93cee961bdc6ff827fd117f174f06f745ae.
//
// Solidity: event Withdrawn(address indexed user, uint256 stakeIndex, uint256 totalAmount, uint256 standAmount, uint256 reward)
func (_Contracts *ContractsFilterer) ParseWithdrawn(log types.Log) (*ContractsWithdrawn, error) {
	event := new(ContractsWithdrawn)
	if err := _Contracts.contract.UnpackLog(event, "Withdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
