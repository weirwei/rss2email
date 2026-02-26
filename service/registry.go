package service

import (
	"context"
	"fmt"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/rss2email/conf"
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
	translationCfg := conf.TranslationConf
	translator := newTranslator(translationCfg)
	return Config{
		FeedURL:      feedSource.FeedURL,
		Subscription: feedSource.Subscription,
		BuildFunc: func(feed *gofeed.Feed) (string, string) {
			return buildEmail(feed, contentField, translator, translationCfg)
		},
	}, nil
}

// buildEmail creates email subject and body from a feed using the specified content field
func buildEmail(feed *gofeed.Feed, contentField ContentField, translator Translator, translationCfg conf.TranslationConfig) (subject string, body string) {
	subject = feed.Title
	for _, item := range feed.Items {
		itemTitle := item.Title
		if translationCfg.Enabled && translationCfg.TranslateTitle {
			translatedTitle := safeTranslate(translator, item.Title)
			if translationCfg.AppendTranslated && translatedTitle != item.Title {
				itemTitle = fmt.Sprintf("%s（%s）", item.Title, translatedTitle)
			} else {
				itemTitle = translatedTitle
			}
		}
		body += fmt.Sprintf("<h1><a href=\"%s\">%s</a></h1><br>", item.Link, itemTitle)

		originalContent := item.Description
		switch contentField {
		case UseContent:
			originalContent = item.Content
		}
		if !translationCfg.Enabled || !translationCfg.TranslateContent {
			body += fmt.Sprintf("%s<br>", originalContent)
			continue
		}
		translatedContent := safeTranslate(translator, originalContent)
		if translationCfg.AppendTranslated && translatedContent != originalContent {
			body += fmt.Sprintf("%s<br><hr><div><strong>中文翻译：</strong></div>%s<br>", originalContent, translatedContent)
			continue
		}
		body += fmt.Sprintf("%s<br>", translatedContent)
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
