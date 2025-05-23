package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/rss2email/conf"
	"github.com/weirwei/rss2email/constants"
)

var ErrV2exService = errors.New("v2ex 推送失败")

func V2exService(ctx context.Context) error {
	feedURL := conf.FeedSourceConf.V2ex
	if err := CommonService(ctx, Config{
		FeedURL:      feedURL,
		Subscription: constants.SubscriptionV2ex,
		BuildFunc:    buildV2ex,
	}); err != nil {
		return ErrV2exService
	}

	return nil
}

func buildV2ex(feed *gofeed.Feed) (subject string, body string) {
	subject = feed.Title
	for _, item := range feed.Items {
		body += fmt.Sprintf("<h1><a href=\"%s\">%s</a></h1><br>", item.Link, item.Title)
		body += fmt.Sprintf("%s<br>", item.Description)
	}
	body = fmt.Sprintf(module, feed.Title, body)
	return
}
