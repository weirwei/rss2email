package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/rss2email/constants"
)

const (
	decohackFeedURL = "https://decohack.com/feed"
)

var ErrDecoHackService = errors.New("decohack 推送失败")

func DecoHackService(ctx context.Context) error {
	if err := CommonService(ctx, Config{
		FeedURL:      decohackFeedURL,
		Subscription: constants.SubscriptionDecoHack,
		BuildFunc:    buildDecohack,
	}); err != nil {
		return ErrDecoHackService
	}

	return nil
}
func buildDecohack(feed *gofeed.Feed) (subject string, body string) {
	subject = feed.Title
	for _, item := range feed.Items {
		body += fmt.Sprintf("<h1><a href=\"%s\">%s</a></h1><br>", item.Link, item.Title)
		body += fmt.Sprintf("%s<br>", item.Content)
	}
	body = fmt.Sprintf(module, feed.Title, body)
	return
}
