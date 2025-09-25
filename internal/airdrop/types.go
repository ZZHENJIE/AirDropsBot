package airdrop

import (
	"sync"
	"time"
)

// NotifyType 通知类型枚举
type NotifyType int

const (
	Notify10Min NotifyType = iota // 提前10分钟通知
	Notify5Min                    // 提前5分钟通知
	Notify1Min                    // 提前1分钟通知
	NotifyStart                   // 开始时通知
)

// AirdropResponse 币安空投API响应结构
type AirdropResponse struct {
	Code    string      `json:"code"`
	Data    AirdropData `json:"data"`
	Success bool        `json:"success"`
}

type AirdropData struct {
	Configs []AirdropConfig `json:"configs"`
}

type AirdropConfig struct {
	ConfigID        string `json:"configId"`
	ConfigName      string `json:"configName"`
	BinanceChainID  string `json:"binanceChainId"`
	ContractAddress string `json:"contractAddress"`
	TokenSymbol     string `json:"tokenSymbol"`
	AirdropAmount   int    `json:"airdropAmount"`
	ClaimStartTime  int64  `json:"claimStartTime"` // 毫秒时间戳
	ClaimEndTime    int64  `json:"claimEndTime"`   // 毫秒时间戳
	Status          string `json:"status"`
}

// TokenResponse 代币信息API响应结构
type TokenResponse struct {
	Code    string    `json:"code"`
	Data    TokenData `json:"data"`
	Success bool      `json:"success"`
}

type TokenData struct {
	MetaInfo  MetaInfo  `json:"metaInfo"`
	StatsInfo StatsInfo `json:"statsInfo"`
}

type MetaInfo struct {
	Name            string `json:"name"`
	Symbol          string `json:"symbol"`
	ChainName       string `json:"chainName"`
	ContractAddress string `json:"contractAddress"`
	Description     string `json:"cnDescription"`
}

type StatsInfo struct {
	TotalSupply       string `json:"totalSupply"`
	CirculatingSupply string `json:"circulatingSupply"`
	Holders           string `json:"holders"`
}

// NotificationRecord 通知记录结构体，用于记录每个空投项目的通知状态
type NotificationRecord struct {
	ConfigID    string              // 空投配置ID
	NotifyTypes map[NotifyType]bool // 各时间点的通知状态记录
	LastCheck   time.Time           // 最后检查时间
}

// AirdropCache 空投缓存管理器，用于内存优化和避免重复请求
type AirdropCache struct {
	sync.RWMutex                                // 读写锁，保证并发安全
	Records      map[string]*NotificationRecord // 通知记录缓存：ConfigID -> NotificationRecord
	Tokens       map[string]*TokenData          // 代币信息缓存：ConfigID -> TokenData
	Config       map[string]*AirdropConfig      // 空投配置缓存：ConfigID -> AirdropConfig
}

// NewAirdropCache 创建新的缓存管理器
func NewAirdropCache() *AirdropCache {
	return &AirdropCache{
		Records: make(map[string]*NotificationRecord),
		Tokens:  make(map[string]*TokenData),
		Config:  make(map[string]*AirdropConfig),
	}
}

// GetNotificationRecord 获取通知记录
func (c *AirdropCache) GetNotificationRecord(configID string) *NotificationRecord {
	c.RLock()
	record, exists := c.Records[configID]
	c.RUnlock()

	if !exists {
		c.Lock()
		record = &NotificationRecord{
			ConfigID:    configID,
			NotifyTypes: make(map[NotifyType]bool),
			LastCheck:   time.Now(),
		}
		c.Records[configID] = record
		c.Unlock()
	}

	return record
}

// CleanExpired 清理过期的缓存
func (c *AirdropCache) CleanExpired() {
	c.Lock()
	defer c.Unlock()

	now := time.Now()
	for configID, config := range c.Config {
		// 如果空投已经结束超过1小时，清理相关缓存
		if now.UnixMilli() > config.ClaimEndTime+3600000 {
			delete(c.Records, configID)
			delete(c.Tokens, configID)
			delete(c.Config, configID)
		}
	}
}

// UpdateConfig 更新空投配置
func (c *AirdropCache) UpdateConfig(config *AirdropConfig) {
	c.Lock()
	defer c.Unlock()
	c.Config[config.ConfigID] = config
}

// UpdateToken 更新代币信息
func (c *AirdropCache) UpdateToken(configID string, token *TokenData) {
	c.Lock()
	defer c.Unlock()
	c.Tokens[configID] = token
}

// GetAirdropInfo 获取空投信息
func (c *AirdropCache) GetAirdropInfo(configID string) (*AirdropConfig, *TokenData, bool) {
	c.RLock()
	defer c.RUnlock()

	config, configExists := c.Config[configID]
	token, tokenExists := c.Tokens[configID]

	if !configExists || !tokenExists {
		return nil, nil, false
	}

	return config, token, true
}

// ShouldNotify 检查是否应该发送特定类型的通知
func (c *AirdropCache) ShouldNotify(configID string, notifyType NotifyType) bool {
	record := c.GetNotificationRecord(configID)

	c.Lock()
	defer c.Unlock()

	if record.NotifyTypes[notifyType] {
		return false
	}

	record.NotifyTypes[notifyType] = true
	return true
}
