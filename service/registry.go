package service

import (
	"context"
	"fmt"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/rss2email/conf"
	"github.com/weirwei/rss2email/constants"
)

// ContentField specifies which field to use from feed items for email body
type ContentField int

const (
	// UseDescription uses item.Description for email body
	UseDescription ContentField = iota
	// UseContent uses item.Content for email body
	UseContent
)

// ServiceConfig holds the configuration for an RSS service
type ServiceConfig struct {
	Name         string
	Subscription constants.SubscriptionID
	GetFeedURL   func() string
	ContentField ContentField
}

// registry maps subscription IDs to their configurations
var registry = map[constants.SubscriptionID]ServiceConfig{
	constants.SubscriptionRuanyifeng: {
		Name:         "ruanyifeng",
		Subscription: constants.SubscriptionRuanyifeng,
		GetFeedURL:   func() string { return conf.FeedSourceConf.Ruanyifeng },
		ContentField: UseDescription,
	},
	constants.SubscriptionDecoHack: {
		Name:         "decohack",
		Subscription: constants.SubscriptionDecoHack,
		GetFeedURL:   func() string { return conf.FeedSourceConf.DecoHack },
		ContentField: UseContent,
	},
	constants.SubscriptionSspai: {
		Name:         "sspai",
		Subscription: constants.SubscriptionSspai,
		GetFeedURL:   func() string { return conf.FeedSourceConf.Sspai },
		ContentField: UseDescription,
	},
	constants.SubscriptionZhihu: {
		Name:         "zhihu",
		Subscription: constants.SubscriptionZhihu,
		GetFeedURL:   func() string { return conf.FeedSourceConf.Zhihu },
		ContentField: UseDescription,
	},
	constants.SubscriptionV2ex: {
		Name:         "v2ex",
		Subscription: constants.SubscriptionV2ex,
		GetFeedURL:   func() string { return conf.FeedSourceConf.V2ex },
		ContentField: UseDescription,
	},
	constants.SubscriptionKitekagi: {
		Name:         "kitekagi",
		Subscription: constants.SubscriptionKitekagi,
		GetFeedURL:   func() string { return conf.FeedSourceConf.Kitekagi },
		ContentField: UseDescription,
	},
	constants.SubscriptionKitekagiAI: {
		Name:         "kitekagi-ai",
		Subscription: constants.SubscriptionKitekagiAI,
		GetFeedURL:   func() string { return conf.FeedSourceConf.KitekagiAI },
		ContentField: UseDescription,
	},
	constants.SubscriptionAIInsightDaily: {
		Name:         "ai-insight-daily",
		Subscription: constants.SubscriptionAIInsightDaily,
		GetFeedURL:   func() string { return conf.FeedSourceConf.AIInsightDaily },
		ContentField: UseContent,
	},
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
	cfg, ok := registry[subscriptionID]
	if !ok {
		return fmt.Errorf("unknown subscription: %s", subscriptionID)
	}

	return CommonService(ctx, Config{
		FeedURL:      cfg.GetFeedURL(),
		Subscription: cfg.Subscription,
		BuildFunc: func(feed *gofeed.Feed) (string, string) {
			return buildEmail(feed, cfg.ContentField)
		},
	})
}

// GetServiceFunc returns a service function for the given subscription ID.
// This is useful for maintaining backward compatibility with existing code.
func GetServiceFunc(subscriptionID constants.SubscriptionID) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return RunService(ctx, subscriptionID)
	}
}
