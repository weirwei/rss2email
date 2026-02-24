package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/weirwei/rss2email/constants"
	"github.com/weirwei/rss2email/models"
)

func init() {
	rootCmd.AddCommand(addFeedCmd)
}

var addFeedCmd = &cobra.Command{
	Use:   "add-feed <subscription_id> <feed_url> <name> [content_field] [schedule_type] [cron_spec]",
	Short: "add or update feed source",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 3 {
			cmd.Help()
			return
		}
		subscriptionID := constants.SubscriptionID(strings.TrimSpace(args[0]))
		feedURL := strings.TrimSpace(args[1])
		name := strings.TrimSpace(args[2])
		contentField := constants.FeedContentFieldDescription
		scheduleType := constants.ScheduleTypeLive
		cronSpec := ""
		if len(args) >= 4 {
			contentField = constants.FeedContentField(strings.ToLower(strings.TrimSpace(args[3])))
		}
		if len(args) >= 5 {
			scheduleType = constants.ScheduleType(strings.ToLower(strings.TrimSpace(args[4])))
		}
		if len(args) >= 6 {
			cronSpec = strings.TrimSpace(args[5])
		}
		if subscriptionID == "" || feedURL == "" || name == "" {
			cmd.Printf("subscription_id, feed_url, name cannot be empty\n")
			return
		}
		if contentField != constants.FeedContentFieldDescription && contentField != constants.FeedContentFieldContent {
			cmd.Printf("content_field must be description or content\n")
			return
		}
		if scheduleType != constants.ScheduleTypeStartup && scheduleType != constants.ScheduleTypeLive && scheduleType != constants.ScheduleTypeCron {
			cmd.Printf("schedule_type must be startup, live or cron\n")
			return
		}
		if scheduleType == constants.ScheduleTypeCron && cronSpec == "" {
			cmd.Printf("cron_spec required when schedule_type is cron\n")
			return
		}
		feedSource := &models.FeedSource{
			Subscription: subscriptionID,
			Name:         name,
			FeedURL:      feedURL,
			ContentField: contentField,
			ScheduleType: scheduleType,
			CronSpec:     cronSpec,
			Deleted:      0,
		}
		if err := models.NewFeedSourceDao().Upsert(cmd.Context(), feedSource); err != nil {
			cmd.Printf("add feed failed, %v\n", err)
			return
		}
		cmd.Printf("add feed success, subscription: %s\n", subscriptionID)
	},
}
