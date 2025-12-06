package utils

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math"
	"math/big"
	"strconv"
	"strings"
)

func IsEmptyOrSpaceString(s string) bool {
	// 去除所有空白字符后判断是否为空
	return strings.TrimSpace(s) == ""
}

func StringToBigInt(s string) (*big.Int, error) {
	res := new(big.Int)
	if _, ok := res.SetString(s, 10); !ok {
		return nil, fmt.Errorf("amount format is invalid: %s", s)
	}
	return res, nil
}

func BigIntToString(i *big.Int) string {
	return i.String()
}

// BigIntToInt64 将 *big.Int 转换为 int64，若超出范围返回错误
func BigIntToInt64(b *big.Int) (int64, error) {
	// 检查是否超出int64范围
	if b.Cmp(big.NewInt(math.MaxInt64)) > 0 || b.Cmp(big.NewInt(math.MinInt64)) < 0 {
		return 0, fmt.Errorf("值超出int64范围")
	}
	return b.Int64(), nil
}

// Int64ToBigInt 将 int64 转换为 *big.Int
func Int64ToBigInt(n int64) *big.Int {
	return big.NewInt(n)
}

func Uint64ToBigInt(n uint64) *big.Int {
	//return new(big.Int).SetInt64(int64(n))
	return big.NewInt(int64(n))
}

func uintToString(i uint64) string {
	return strconv.FormatUint(i, 10)
}

func stringToHash(s string) common.Hash {
	return common.HexToHash(s)
}
