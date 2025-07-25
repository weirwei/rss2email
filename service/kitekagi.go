package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/rss2email/conf"
	"github.com/weirwei/rss2email/constants"
)

var ErrKitekagiService = errors.New("kitekagi 推送失败")

func KitekagiService(ctx context.Context) error {
	feedURL := conf.FeedSourceConf.Kitekagi
	if err := CommonService(ctx, Config{
		FeedURL:      feedURL,
		Subscription: constants.SubscriptionKitekagi,
		BuildFunc:    buildKitekagi,
	}); err != nil {
		return ErrKitekagiService
	}

	return nil
}

func KitekagiAIService(ctx context.Context) error {
	feedURL := conf.FeedSourceConf.KitekagiAI
	if err := CommonService(ctx, Config{
		FeedURL:      feedURL,
		Subscription: constants.SubscriptionKitekagiAI,
		BuildFunc:    buildKitekagi,
	}); err != nil {
		return ErrKitekagiService
	}

	return nil
}

func buildKitekagi(feed *gofeed.Feed) (subject string, body string) {
	subject = feed.Title
	for _, item := range feed.Items {
		body += fmt.Sprintf("<h1><a href=\"%s\">%s</a></h1><br>", item.Link, item.Title)
		body += fmt.Sprintf("%s<br>", item.Description)
	}
	body = fmt.Sprintf(module, feed.Title, body)
	return
}
