package cmd

import (
	"context"
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
	)
	live(ctx, c, service.DecoHackService)
	// 每天10:30
	customize(ctx, c, "30 10 * * *", service.SspaiService, service.ZhihuService)
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
		if err := fn(ctx); err != nil {
			ilog.Errorf("订阅失败，%v", err)
		}
		time.Sleep(1 * time.Minute) // 避免过快触发
	}
	ilog.Info("结束启动订阅...")
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
