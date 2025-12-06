package config

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
	"time"
)

// Config 应用程序配置
type Config struct {
	AppConfig        `yaml:"app"`
	DatabaseConfig   `yaml:"database"`
	RedisConfig      `yaml:"redis"`
	BlockchainConfig `yaml:"blockchain"`
	AuthConfig       `yaml:"auth"`
	LogConfig        `yaml:"logger"`
}

// ===== 应用配置模块 =====
type AppConfig struct {
	Environment string        `yaml:"environment"`
	Debug       bool          `yaml:"debug"`
	Version     string        `yaml:"version"`
	Host        string        `yaml:"host"`
	Port        int           `yaml:"port"`
	LogLevel    string        `yaml:"log_level"`
	Timeout     time.Duration `yaml:"timeout"`
}

// ===== 数据库配置模块 =====
type DatabaseConfig struct {
	Driver          string        `yaml:"driver"`
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	DatabaseConfig  string        `yaml:"database"`
	Charset         string        `yaml:"charset"`
	TimeZone        string        `yaml:"timezone"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	LogLevel        string        `yaml:"log_level"`
}

// DSN 生成数据库连接字符串
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
		c.Username, c.Password, c.Host, c.Port,
		c.DatabaseConfig, c.Charset, c.TimeZone,
	)
}

// ===== Redis配置模块 =====
type RedisConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	Password       string        `yaml:"password"`
	DatabaseConfig int           `yaml:"database"`       // 数据库编号 (0-15)
	PoolSize       int           `yaml:"pool_size"`      // 连接池大小
	MinIdleConns   int           `yaml:"min_idle_conns"` // 最小空闲连接数
	DialTimeout    time.Duration `yaml:"dial_timeout"`   // 建立连接超时
	ReadTimeout    time.Duration `yaml:"read_timeout"`   // 读操作超时
	WriteTimeout   time.Duration `yaml:"write_timeout"`  // 写操作超时
	IdleTimeout    time.Duration `yaml:"idle_timeout"`   // 空闲连接超时

	// 锁配置
	Locks map[string]LockConfig `yaml:"lock"`
}

type LockConfig struct {
	Expiration time.Duration `yaml:"expiration"`
	Timeout    time.Duration `yaml:"timeout"`
}

// Addr 返回Redis地址
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetLockConfig 获取锁配置
func (c *RedisConfig) GetLockConfig(lockType string) LockConfig {
	if config, exists := c.Locks[lockType]; exists {
		return config
	}
	// 返回默认配置
	return LockConfig{
		Expiration: 30 * time.Second,
		Timeout:    5 * time.Second,
	}
}

// ===== 区块链配置模块 =====
type BlockchainConfig struct {
	Network    string   `yaml:"network"`
	RpcURL     string   `yaml:"rpc_url"`
	RawURL     string   `yaml:"raw_url"`
	PrivateKey string   `yaml:"private_key"`
	GasPrice   string   `yaml:"gas_price"`
	Owners     []string `yaml:"owners"`

	// 同步配置
	Sync SyncConfig `yaml:"sync"`

	// 合约地址
	Contracts ContractAddresses `yaml:"contracts"`
}

type SyncConfig struct {
	BatchSize     int           `yaml:"batch_size"`
	BlockBuffer   uint64        `yaml:"block_buffer"`
	Workers       int           `yaml:"workers"`
	SyncInterval  time.Duration `yaml:"sync_interval"`
	RetryAttempts int           `yaml:"retry_attempts"`
	RetryDelay    time.Duration `yaml:"retry_delay"`
}

type ContractAddresses struct {
	Stake   string `yaml:"stake_address"`
	Airdrop string `yaml:"airdrop_address"`
	Token   string `yaml:"token_address"`
}

type AuthConfig struct {
	JwtSecret        string             `yaml:"jwt_secret"`
	JwtExpiration    time.Duration      `yaml:"jwt_expiration"`
	Prefix           string             `yaml:"prefix"`
	MaxDelta         int                `yaml:"max_delta"` //分钟
	EcdsaPublicKey   *ecdsa.PublicKey   `yaml:"ecdsa_public_key"`
	EcdsaPrivateKey  string             `yaml:"ecdsa_private_key"`
	Ed25519PublicKey *ed25519.PublicKey `yaml:"ed25519_public_key"`
	Ed25519Seed      string             `yaml:"ed25519_seed"`
}

type LogConfig struct {
	Level        logrus.Level `yaml:"level"`
	IsJsonFormat bool         `yaml:"is_json_format"`
	LogPath      string       `yaml:"log_path"`
	LogFile      string       `yaml:"log_file"`
	MaxAge       int          `yaml:"maxAge"`
}

// ===== 配置加载和管理 =====
var (
	globalConfig *Config
	once         sync.Once
)

func init() {
	once.Do(func() {
		configPath := getDefaultConfigPath()
		cfg, err := Load(configPath)
		if err != nil {
			panic(fmt.Sprintf("Failed to load config: %v", err))
		}
		globalConfig = cfg
	})
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	_ = godotenv.Load(".env")
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", configPath, err)
	}

	// 替换环境变量
	content := expandEnvVars(string(data))

	var config Config
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	// 设置默认值
	setDefaults(&config)

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	globalConfig = &config
	return &config, nil
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		panic("Config is not initialized")
	}
	return globalConfig
}

// ===== 工具函数 =====
func getDefaultConfigPath() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}
	return fmt.Sprintf("common/config/config.%s.yaml", env)
}

// expandEnvVars 替换配置文件中的环境变量
func expandEnvVars(content string) string {
	return os.ExpandEnv(content)
}

// setDefaults 设置默认值
func setDefaults(config *Config) {
	// AppConfig 默认值
	if config.AppConfig.Environment == "" {
		config.AppConfig.Environment = "local"
	}
	if config.AppConfig.Host == "" {
		config.AppConfig.Host = "localhost"
	}
	if config.AppConfig.Port == 0 {
		config.AppConfig.Port = 8084
	}
	if config.AppConfig.LogLevel == "" {
		config.AppConfig.LogLevel = "info"
	}
	if config.AppConfig.Timeout == 0 {
		config.AppConfig.Timeout = 30 * time.Second
	}

	// DatabaseConfig 默认值
	if config.DatabaseConfig.Driver == "" {
		config.DatabaseConfig.Driver = "mysql"
	}
	if config.DatabaseConfig.Port == 0 {
		config.DatabaseConfig.Port = 3306
	}
	if config.DatabaseConfig.Charset == "" {
		config.DatabaseConfig.Charset = "utf8mb4"
	}
	if config.DatabaseConfig.TimeZone == "" {
		config.DatabaseConfig.TimeZone = "Local"
	}
	if config.DatabaseConfig.MaxOpenConns == 0 {
		config.DatabaseConfig.MaxOpenConns = 20
	}
	if config.DatabaseConfig.MaxIdleConns == 0 {
		config.DatabaseConfig.MaxIdleConns = 10
	}
	if config.DatabaseConfig.ConnMaxLifetime == 0 {
		config.DatabaseConfig.ConnMaxLifetime = time.Hour
	}

	// RedisConfig 默认值
	if config.RedisConfig.Port == 0 {
		config.RedisConfig.Port = 6379
	}
	if config.RedisConfig.PoolSize == 0 {
		config.RedisConfig.PoolSize = 10
	}
	if config.RedisConfig.MinIdleConns == 0 {
		config.RedisConfig.MinIdleConns = 5
	}
	if config.RedisConfig.DialTimeout == 0 {
		config.RedisConfig.DialTimeout = 5 * time.Second
	}
	if config.RedisConfig.ReadTimeout == 0 {
		config.RedisConfig.ReadTimeout = 3 * time.Second
	}
	if config.RedisConfig.WriteTimeout == 0 {
		config.RedisConfig.WriteTimeout = 3 * time.Second
	}
	if config.RedisConfig.IdleTimeout == 0 {
		config.RedisConfig.IdleTimeout = 5 * time.Minute
	}

	// 设置默认锁配置
	if config.RedisConfig.Locks == nil {
		config.RedisConfig.Locks = map[string]LockConfig{
			"lock": {
				Expiration: 10 * time.Second,
				Timeout:    10 * time.Second,
			},
			"withdraw_lock": {
				Expiration: 10 * time.Second,
				Timeout:    10 * time.Second,
			},
			"transaction_lock": {
				Expiration: 10 * time.Second,
				Timeout:    10 * time.Second,
			},
			"asset_lock": {
				Expiration: 10 * time.Second,
				Timeout:    10 * time.Second,
			},
		}
	}

	// BlockchainConfig 默认值
	if config.BlockchainConfig.Network == "" {
		config.BlockchainConfig.Network = "testnet"
	}
	if config.BlockchainConfig.Sync.BatchSize == 0 {
		config.BlockchainConfig.Sync.BatchSize = 100
	}
	if config.BlockchainConfig.Sync.BlockBuffer == 0 {
		config.BlockchainConfig.Sync.BlockBuffer = 30
	}
	if config.BlockchainConfig.Sync.Workers == 0 {
		config.BlockchainConfig.Sync.Workers = 5
	}
	if config.BlockchainConfig.Sync.SyncInterval == 0 {
		config.BlockchainConfig.Sync.SyncInterval = 10 * time.Second
	}
	if config.BlockchainConfig.Sync.RetryAttempts == 0 {
		config.BlockchainConfig.Sync.RetryAttempts = 3
	}
	if config.BlockchainConfig.Sync.RetryDelay == 0 {
		config.BlockchainConfig.Sync.RetryDelay = 5 * time.Second
	}

	if config.LogConfig.Level == 0 {
		if config.AppConfig.Environment == "local" {
			config.LogConfig.Level = logrus.InfoLevel
		} else {
			config.LogConfig.Level = logrus.DebugLevel
		}
	}
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	if config.AppConfig.Environment == "" {
		return fmt.Errorf("app.environment is required")
	}

	if config.DatabaseConfig.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if config.DatabaseConfig.Username == "" {
		return fmt.Errorf("database.username is required")
	}
	if config.DatabaseConfig.DatabaseConfig == "" {
		return fmt.Errorf("database.database is required")
	}

	if config.RedisConfig.Host == "" {
		return fmt.Errorf("redis.host is required")
	}

	if config.BlockchainConfig.RpcURL == "" {
		return fmt.Errorf("blockchain.rpc_url is required")
	}

	return nil
}

func SetEcdsaPublicKey(config *Config, publicKey *ecdsa.PublicKey) {
	config.AuthConfig.EcdsaPublicKey = publicKey
}

func SetEd25519PublicKey(config *Config, publicKey *ed25519.PublicKey) {
	config.AuthConfig.Ed25519PublicKey = publicKey
}
