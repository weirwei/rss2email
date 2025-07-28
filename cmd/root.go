package cmd

import (
	"context"
	"math/rand"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"github.com/weirwei/ikit/ilog"
	"github.com/weirwei/rss2email/service"
)

var rootCmd = &cobra.Command{
	Use:   "rss2email",
	Short: "rss to email",
	Run: func(cmd *cobra.Command, args []string) {
		exec()
	},
}

func exec() {
	ctx := context.Background()
	c := cron.New()
	ilog.Info("RSS2Email 启动...")

	c.AddFunc("0 * * * *", func() {
		// 探活
		ilog.Info("active")
	})
	triggerAtStartUp(ctx,
		service.RuanyifengService, // 阮一峰
		service.DecoHackService,   // DecoHack
		service.SspaiService,      // 少数派
		service.ZhihuService,      // 知乎
		service.KitekagiService,   // Kitekagi 世界
		service.KitekagiAIService, // Kitekagi 人工智能
	)
	live(ctx, c, service.DecoHackService)
	// 每天10:30
	customize(ctx, c, "30 10 * * *", service.SspaiService, service.ZhihuService)
	// 每天12:30
	customize(ctx, c, "30 12 * * *", service.KitekagiService, service.KitekagiAIService)
	// 每周五10点开始，每3个小时请求一次
	customize(ctx, c, "0 10/3 * * 5", service.RuanyifengService)
	c.Start()
	defer c.Stop()
	select {} // 阻塞主线程，防止退出
}

func Execute() error {
	return rootCmd.Execute()
}

func triggerAtStartUp(ctx context.Context, fns ...func(ctx context.Context) error) {
	ilog.Info("启动订阅...")
	for _, fn := range fns {
		// 使用指数退避策略重试
		if err := exponentialBackoffRetry(ctx, fn); err != nil {
			ilog.Errorf("订阅失败，%v", err)
		}
	}
	ilog.Info("结束启动订阅...")
}

// exponentialBackoffRetry 使用指数退避策略重试函数调用
func exponentialBackoffRetry(ctx context.Context, fn func(ctx context.Context) error) error {
	const (
		initialDelay = 1 * time.Second  // 初始延迟1秒
		maxDelay     = 5 * time.Minute  // 最大延迟5分钟
		maxRetries   = 5                // 最大重试次数
		multiplier   = 2.0              // 指数增长因子
		jitter       = 0.1              // 抖动因子
	)
	
	delay := initialDelay
	
	for i := 0; i <= maxRetries; i++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// 执行函数
		err := fn(ctx)
		if err == nil {
			// 成功执行，返回nil
			return nil
		}
		
		// 如果是最后一次重试，返回错误
		if i == maxRetries {
			return err
		}
		
		// 记录重试信息
		ilog.Warnf("函数执行失败，将在 %v 后进行第 %d 次重试: %v", delay, i+1, err)
		
		// 等待退避时间或上下文取消
		select {
		case <-time.After(delay):
			// 计算下一次的延迟时间（指数增长 + 随机抖动）
			delay = time.Duration(float64(delay) * multiplier)
			// 添加随机抖动避免惊群效应
			jitterAmount := time.Duration(float64(delay) * jitter * (0.5 - rand.Float64()) * 2)
			delay += jitterAmount
			// 确保不超过最大延迟时间
			if delay > maxDelay {
				delay = maxDelay
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	return nil
}

// 实时
func live(ctx context.Context, c *cron.Cron, fns ...func(ctx context.Context) error) {
	if c == nil {
		c = cron.New()
	}

	// 每个小时
	c.AddFunc("0 */1 * * *", func() {
		ilog.Info("开始实时订阅...")
		for _, fn := range fns {
			if err := fn(ctx); err != nil {
				ilog.Errorf("订阅失败，%v", err)
			}
		}
		ilog.Info("结束实时订阅...")
	})
}

func daily(ctx context.Context, c *cron.Cron, fns ...func(ctx context.Context) error) {
	if c == nil {
		c = cron.New()
	}
	// 每天10:30
	c.AddFunc("30 10 */ * *", func() {
		ilog.Info("开始每日订阅...")
		for _, fn := range fns {
			if err := fn(ctx); err != nil {
				ilog.Errorf("订阅失败，%v", err)
			}
		}
	})
}

func customize(ctx context.Context, c *cron.Cron, crontab string, fns ...func(ctx context.Context) error) {
	_, err := cron.ParseStandard(crontab)
	if err != nil {
		ilog.Errorf("cron 表达式不合法，%v", err)
		return
	}
	c.AddFunc(crontab, func() {
		ilog.Infof("开始自定义订阅...[%s]", crontab)
		for _, fn := range fns {
			if err := fn(ctx); err != nil {
				ilog.Errorf("订阅失败，%v", err)
			}
			time.Sleep(1 * time.Minute)
		}
	})
}
