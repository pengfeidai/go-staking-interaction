package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateRandomAmount(count int, maxAmount *big.Int) ([]*big.Int, error) {
	// 创建结果数组
	result := make([]*big.Int, 0, count)

	// 计算随机数范围：[1, maxAmount]
	rangeSize := new(big.Int).Add(maxAmount, big.NewInt(1))

	for i := 0; i < count; i++ {
		// 生成 [0, maxAmount) 的随机数
		randomNum, err := rand.Int(rand.Reader, rangeSize)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random number: %w", err)
		}
		randomNum.Add(randomNum, big.NewInt(1))

		result = append(result, randomNum)
	}

	return result, nil
}

func CalculateSumOfAmounts(amounts []*big.Int) *big.Int {
	sum := new(big.Int)
	for _, amount := range amounts {
		sum = sum.Add(sum, amount)
	}
	return sum
}
