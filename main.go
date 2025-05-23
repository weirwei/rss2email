package main

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/weirwei/ikit/ilog"
	"github.com/weirwei/rss2email/service"
)

func main() {
	ctx := context.Background()
	c := cron.New()
	live(ctx, c, service.DecoHackService)

	c.Start()
	select {} // 阻塞主线程，防止退出
}

// 实时
func live(ctx context.Context, c *cron.Cron, fns ...func(ctx context.Context) error) {
	if c == nil {
		c = cron.New()
	}
	// 每个小时
	c.AddFunc("0 0 */1 * * *", func() {
		ilog.Info("开始实时订阅...")
		for _, fn := range fns {
			if err := fn(ctx); err != nil {
				ilog.Errorf("订阅失败，%v", err)
			}
		}
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
		ilog.Info("开始自定义订阅...")
		for _, fn := range fns {
			if err := fn(ctx); err != nil {
				ilog.Errorf("订阅失败，%v", err)
			}
		}
	})
}
