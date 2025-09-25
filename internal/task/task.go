package task

import (
	"context"
	"log"

	"air-drops-bot/internal/airdrop"
	"air-drops-bot/internal/email"
)

// Task 定时任务结构体，管理空投监控和邮件通知
type Task struct {
	emailService   *email.Service   // 邮件服务实例
	airdropService *airdrop.Service // 空投服务实例
}

// NewTask 创建新的定时任务实例
func NewTask(emailService *email.Service) *Task {
	return &Task{
		emailService:   emailService,
		airdropService: airdrop.NewService(),
	}
}

// MainTask 主任务函数，由调度器定期执行，负责检查新空投并发送邮件通知
func (t *Task) MainTask(ctx context.Context) {
	// 处理空投并获取需要发送的通知
	notifications, err := t.airdropService.ProcessAirdrops()
	if err != nil {
		log.Printf("处理空投失败: %v", err)
		return
	}

	// 发送通知
	for _, notif := range notifications {
		message := t.airdropService.FormatNotification(notif.Config, notif.Token, notif.NotifyType)

		err := t.emailService.SendEmail(
			"AirDropsBot 空投提醒",
			"新空投提醒："+notif.Token.MetaInfo.Name,
			message,
			true,
		)

		if err != nil {
			log.Printf("发送通知失败 [%s]: %v", notif.Config.ConfigID, err)
			continue
		}

		log.Printf("已发送空投通知: %s (%s)", notif.Config.ConfigName, notif.Config.TokenSymbol)
	}
}
