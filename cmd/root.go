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
	service.RunAllByDB(ctx)
	service.ScheduleFromDBDynamic(ctx, c, time.Minute)
	c.Start()
	defer c.Stop()
	select {} // 阻塞主线程，防止退出
}

func Execute() error {
	return rootCmd.Execute()
}
