package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	cfgpkg "github.com/ZZHENJIE/AirDropsBot/internal/config"
)

// EmailPayload 邮件发送请求的数据结构
type EmailPayload struct {
	FromTitle     string `json:"fromTitle"`
	Subject       string `json:"subject"`
	Content       string `json:"content"`
	IsTextContent bool   `json:"isTextContent"`
	ToMail        string `json:"tomail"`
	ColaKey       string `json:"ColaKey"`
	SmtpCode      string `json:"smtpCode"`
	SmtpCodeType  string `json:"smtpCodeType"`
	SmtpEmail     string `json:"smtpEmail"`
}

// Service 邮件服务结构体，提供邮件发送功能
type Service struct {
	cfg *cfgpkg.Manager
}

// NewService 创建新的邮件服务实例，需要提供配置管理器
func NewService(cfg *cfgpkg.Manager) *Service {
	return &Service{cfg: cfg}
}

// SendEmail 发送邮件到配置文件中指定的所有收件人
// fromTitle: 发件人的显示名称（如“AirDropsBot”）
// subject: 邮件的主题行
// content: 邮件的内容
// isTextContent: 内容是否为纯文本格式
func (s *Service) SendEmail(fromTitle, subject, content string, isTextContent bool) error {
	// 获取最新的配置
	config := s.cfg.Get()
	emailConfig := config.Email

	// 创建邮件模板
	template := EmailPayload{
		FromTitle:     fromTitle,
		Subject:       subject,
		Content:       content,
		IsTextContent: isTextContent,
		ColaKey:       emailConfig.ColaKey,
		SmtpCode:      emailConfig.SmtpCode,
		SmtpCodeType:  emailConfig.SmtpCodeType,
		SmtpEmail:     emailConfig.SmtpEmail,
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 为每个收件人发送邮件
	for _, tomail := range emailConfig.ToMail {
		template.ToMail = tomail

		// 序列化请求体
		body, err := json.Marshal(template)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}

		// 创建请求
		req, err := http.NewRequest("POST", "https://luckycola.com.cn/tools/customMail", bytes.NewBuffer(body))
		if err != nil {
			return fmt.Errorf("create request failed: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		// 发送请求
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("send request failed for %s: %w", tomail, err)
		}
		defer resp.Body.Close()

		// 检查响应状态
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("send email failed for %s with status code: %d", tomail, resp.StatusCode)
		}
	}

	return nil
}
