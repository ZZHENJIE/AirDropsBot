package airdrop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	AirdropListURL = "https://www.binance.com/bapi/defi/v1/friendly/wallet-direct/buw/growth/query-alpha-airdrop"
	TokenInfoURL   = "https://www.maxweb.systems/bapi/defi/v1/public/wallet-direct/buw/wallet/cex/alpha/token/full/info"
)

// Service 空投服务结构体，管理空投监控和通知功能
type Service struct {
	client *http.Client  // HTTP客户端，用于调用币安API
	cache  *AirdropCache // 缓存管理器
}

// NewService 创建新的空投服务实例
func NewService() *Service {
	return &Service{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: NewAirdropCache(),
	}
}

// FetchAirdrops 从币安API获取最新的空投列表
func (s *Service) FetchAirdrops() ([]AirdropConfig, error) {
	payload := map[string]interface{}{
		"page": 1,
		"rows": 20,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	req, err := http.NewRequest("POST", AirdropListURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result AirdropResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if !result.Success || result.Code != "000000" {
		return nil, fmt.Errorf("api error: %s", result.Code)
	}

	return result.Data.Configs, nil
}

// FetchTokenInfo 获取代币信息
func (s *Service) FetchTokenInfo(chainId, contractAddress string) (*TokenData, error) {
	url := fmt.Sprintf("%s?chainId=%s&contractAddress=%s", TokenInfoURL, chainId, contractAddress)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if !result.Success || result.Code != "000000" {
		return nil, fmt.Errorf("api error: %s", result.Code)
	}

	return &result.Data, nil
}

// getNotifyTypeForTime 根据剩余时间获取通知类型
func (s *Service) getNotifyTypeForTime(remainingTime time.Duration) (NotifyType, bool) {
	switch {
	case remainingTime <= 0:
		return NotifyStart, true
	case remainingTime <= time.Minute:
		return Notify1Min, true
	case remainingTime <= 5*time.Minute:
		return Notify5Min, true
	case remainingTime <= 10*time.Minute:
		return Notify10Min, true
	default:
		return NotifyStart, false
	}
}

// FormatNotification 根据空投信息和通知类型格式化邮件通知内容
func (s *Service) FormatNotification(config *AirdropConfig, token *TokenData, notifyType NotifyType) string {
	var timeMsg string
	switch notifyType {
	case NotifyStart:
		timeMsg = "空投现在开始啦！"
	case Notify1Min:
		timeMsg = "还有 1 分钟开始"
	case Notify5Min:
		timeMsg = "还有 5 分钟开始"
	case Notify10Min:
		timeMsg = "还有 10 分钟开始"
	}

	return fmt.Sprintf(`🚀 空投提醒：%s (%s)

%s
项目简介：%s

• 空投数量：%d
• 链名称：%s
• 合约地址：%s
• 总供应量：%s
• 流通量：%s
• 持有人数：%s

时间安排：
开始时间：%s
结束时间：%s

赶紧准备参与吧！ 🎉`,
		token.MetaInfo.Name,
		token.MetaInfo.Symbol,
		timeMsg,
		token.MetaInfo.Description,
		config.AirdropAmount,
		token.MetaInfo.ChainName,
		token.MetaInfo.ContractAddress,
		token.StatsInfo.TotalSupply,
		token.StatsInfo.CirculatingSupply,
		token.StatsInfo.Holders,
		time.UnixMilli(config.ClaimStartTime).Format("2006-01-02 15:04:05"),
		time.UnixMilli(config.ClaimEndTime).Format("2006-01-02 15:04:05"),
	)
}

// ProcessAirdrops 处理最新的空投信息并返回需要发送的通知列表
type NotificationInfo struct {
	Config     *AirdropConfig
	Token      *TokenData
	NotifyType NotifyType
}

func (s *Service) ProcessAirdrops() ([]NotificationInfo, error) {
	// 清理过期缓存
	s.cache.CleanExpired()

	// 获取最新空投列表
	airdrops, err := s.FetchAirdrops()
	if err != nil {
		return nil, err
	}

	var notifications []NotificationInfo

	for _, drop := range airdrops {
		// 跳过已结束的空投
		if drop.Status == "ended" {
			continue
		}

		// 更新配置缓存
		s.cache.UpdateConfig(&drop)

		// 计算距离开始时间还有多长时间
		remainingTime := time.Until(time.UnixMilli(drop.ClaimStartTime))

		// 获取应该发送的通知类型
		notifyType, shouldCheck := s.getNotifyTypeForTime(remainingTime)
		if !shouldCheck {
			continue
		}

		// 检查是否已经发送过该类型的通知
		if !s.cache.ShouldNotify(drop.ConfigID, notifyType) {
			continue
		}

		// 获取或更新代币信息
		_, token, exists := s.cache.GetAirdropInfo(drop.ConfigID)
		if !exists {
			var err error
			token, err = s.FetchTokenInfo(drop.BinanceChainID, drop.ContractAddress)
			if err != nil {
				continue // 跳过出错的项目
			}
			s.cache.UpdateToken(drop.ConfigID, token)
		}

		notifications = append(notifications, NotificationInfo{
			Config:     &drop,
			Token:      token,
			NotifyType: notifyType,
		})
	}

	return notifications, nil
}
