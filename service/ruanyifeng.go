package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/rss2email/conf"
	"github.com/weirwei/rss2email/constants"
)

var ErrRuanyifengService = errors.New("ruanyifeng 推送失败")

func RuanyifengService(ctx context.Context) error {
	feedURL := conf.FeedSourceConf.Ruanyifeng
	if err := CommonService(ctx, Config{
		FeedURL:      feedURL,
		Subscription: constants.SubscriptionRuanyifeng,
		BuildFunc:    buildRuanyifeng,
	}); err != nil {
		return ErrRuanyifengService
	}

	return nil
}

func buildRuanyifeng(feed *gofeed.Feed) (subject string, body string) {
	subject = feed.Title
	for _, item := range feed.Items {
		body += fmt.Sprintf("<h1><a href=\"%s\">%s</a></h1><br>", item.Link, item.Title)
		body += fmt.Sprintf("%s<br>", item.Description)
	}
	body = fmt.Sprintf(module, feed.Title, body)
	return
}
