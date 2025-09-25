package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	cfgpkg "github.com/ZZHENJIE/AirDropsBot/internal/config"
	"github.com/ZZHENJIE/AirDropsBot/internal/email"
	"github.com/ZZHENJIE/AirDropsBot/internal/scheduler"
	"github.com/ZZHENJIE/AirDropsBot/internal/server"
	"github.com/ZZHENJIE/AirDropsBot/internal/task"
)

func main() {
	var (
		addr       string
		configPath string
	)
	flag.StringVar(&addr, "addr", ":8080", "HTTP listen address")
	flag.StringVar(&configPath, "config", "config.json", "Path to config file")
	flag.Parse()

	cfgMgr, err := cfgpkg.NewManager(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 创建邮件服务
	emailService := email.NewService(cfgMgr)

	// 创建任务实例
	taskInstance := task.NewTask(emailService)

	// 创建调度器
	sched := scheduler.New()

	// 创建服务器
	srv := server.New(addr, cfgMgr, sched, taskInstance)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Start(ctx); err != nil {
			if err != context.Canceled {
				log.Printf("server stopped: %v", err)
			}
		}
	}()

	<-ctx.Done()
	// Give some time for graceful shutdown
	time.Sleep(300 * time.Millisecond)
	os.Exit(0)
}
