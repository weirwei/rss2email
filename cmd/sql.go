package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weirwei/rss2email/service"
)

func init() {
	rootCmd.AddCommand(dbCmd)
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "db exec",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("缺少参数")
			return
		}
		ctx := context.Background()
		fmt.Println(args)
		service.SQLExec(ctx, args[0])
	},
}
