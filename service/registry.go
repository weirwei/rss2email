package service

import (
	"context"
	"fmt"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/rss2email/constants"
	"github.com/weirwei/rss2email/models"
)

// ContentField specifies which field to use from feed items for email body
type ContentField int

const (
	// UseDescription uses item.Description for email body
	UseDescription ContentField = iota
	// UseContent uses item.Content for email body
	UseContent
)

func contentFieldFromDB(value constants.FeedContentField) ContentField {
	switch value {
	case constants.FeedContentFieldContent:
		return UseContent
	case constants.FeedContentFieldDescription:
		return UseDescription
	default:
		return UseDescription
	}
}

func BuildConfigFromFeedSource(feedSource *models.FeedSource) (Config, error) {
	if feedSource == nil {
		return Config{}, fmt.Errorf("feed source is nil")
	}
	contentField := contentFieldFromDB(feedSource.ContentField)
	return Config{
		FeedURL:      feedSource.FeedURL,
		Subscription: feedSource.Subscription,
		BuildFunc: func(feed *gofeed.Feed) (string, string) {
			return buildEmail(feed, contentField)
		},
	}, nil
}

// buildEmail creates email subject and body from a feed using the specified content field
func buildEmail(feed *gofeed.Feed, contentField ContentField) (subject string, body string) {
	subject = feed.Title
	for _, item := range feed.Items {
		body += fmt.Sprintf("<h1><a href=\"%s\">%s</a></h1><br>", item.Link, item.Title)
		switch contentField {
		case UseContent:
			body += fmt.Sprintf("%s<br>", item.Content)
		default:
			body += fmt.Sprintf("%s<br>", item.Description)
		}
	}
	body = fmt.Sprintf(module, feed.Title, body)
	return
}

// RunService executes the service for the given subscription ID
func RunService(ctx context.Context, subscriptionID constants.SubscriptionID) error {
	feedSource, err := models.NewFeedSourceDao().GetBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		return err
	}
	if feedSource == nil {
		return fmt.Errorf("unknown subscription: %s", subscriptionID)
	}
	config, err := BuildConfigFromFeedSource(feedSource)
	if err != nil {
		return err
	}
	return CommonService(ctx, config)
}

// GetServiceFunc returns a service function for the given subscription ID.
// This is useful for maintaining backward compatibility with existing code.
func GetServiceFunc(subscriptionID constants.SubscriptionID) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return RunService(ctx, subscriptionID)
	}
}
