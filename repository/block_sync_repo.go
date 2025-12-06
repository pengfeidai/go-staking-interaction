package repository

import (
	"errors"
	"os"
	"strconv"
	"sync"
)

// 区块同步管理器
type BlockSyncManager struct {
	// 存储最后同步区块高度的文件路径
	lastSyncedFile string
	// 用于并发安全的互斥锁
	mu sync.Mutex
}

// 新建区块同步管理器
func NewBlockSyncManager(filePath string) *BlockSyncManager {
	return &BlockSyncManager{
		lastSyncedFile: filePath,
	}
}

// GetLastSyncedBlock 获取最后同步的区块高度
// 如果文件不存在或读取失败，返回 0 和错误信息
func (m *BlockSyncManager) GetLastSyncedBlock() (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(m.lastSyncedFile); os.IsNotExist(err) {
		// 文件不存在，返回 0 表示需要从头开始同步
		return 0, nil
	}

	// 读取文件内容
	data, err := os.ReadFile(m.lastSyncedFile)
	if err != nil {
		return 0, errors.New("读取最后同步区块文件失败: " + err.Error())
	}

	// 转换为数字
	blockHeight, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return 0, errors.New("解析区块高度失败: " + err.Error())
	}

	return blockHeight, nil
}

// SaveSyncedBlock 保存当前同步到的区块高度
func (m *BlockSyncManager) SaveSyncedBlock(height uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 将区块高度转换为字符串
	heightStr := strconv.FormatUint(height, 10)

	// 写入文件
	if err := os.WriteFile(m.lastSyncedFile, []byte(heightStr), 0644); err != nil {
		return errors.New("保存区块高度失败: " + err.Error())
	}

	return nil
}
