package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/rss2email/conf"
	"github.com/weirwei/rss2email/constants"
)

var ErrAIInsightDailyService = errors.New("ai-insight-daily 推送失败")

func AIInsightDailyService(ctx context.Context) error {
	feedURL := conf.FeedSourceConf.AIInsightDaily
	if err := CommonService(ctx, Config{
		FeedURL:      feedURL,
		Subscription: constants.SubscriptionAIInsightDaily,
		BuildFunc:    buildAIInsightDaily,
	}); err != nil {
		return ErrAIInsightDailyService
	}

	return nil
}

func buildAIInsightDaily(feed *gofeed.Feed) (subject string, body string) {
	subject = feed.Title
	for _, item := range feed.Items {
		body += fmt.Sprintf("<h1><a href=\"%s\">%s</a></h1><br>", item.Link, item.Title)
		body += fmt.Sprintf("%s<br>", item.Content)
	}
	body = fmt.Sprintf(module, feed.Title, body)
	return
}
