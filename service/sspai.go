package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/rss2email/conf"
	"github.com/weirwei/rss2email/constants"
)

var ErrSspaiService = errors.New("sspai 推送失败")

func SspaiService(ctx context.Context) error {
	feedURL := conf.FeedSourceConf.Sspai
	if err := CommonService(ctx, Config{
		FeedURL:      feedURL,
		Subscription: constants.SubscriptionSspai,
		BuildFunc:    buildSspai,
	}); err != nil {
		return ErrSspaiService
	}

	return nil
}

func buildSspai(feed *gofeed.Feed) (subject string, body string) {
	subject = feed.Title
	for _, item := range feed.Items {
		body += fmt.Sprintf("<h1><a href=\"%s\">%s</a></h1><br>", item.Link, item.Title)
		body += fmt.Sprintf("%s<br>", item.Description)
	}
	body = fmt.Sprintf(module, feed.Title, body)
	return
}
