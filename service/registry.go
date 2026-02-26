package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/ikit/ilog"
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
	maxItems := translationCfg.MaxItemsPerPush
	if maxItems <= 0 {
		maxItems = 10
	}
	maxTextChars := translationCfg.MaxTextCharsPerPush
	if maxTextChars <= 0 {
		maxTextChars = 30000
	}
	var (
		selectedItems int
		usedTextChars int
	)
	ilog.Infof(
		"build email start title=%q items=%d content_field=%d translation_enabled=%t title=%t content=%t max_items=%d max_text_chars=%d",
		feed.Title,
		len(feed.Items),
		contentField,
		translationCfg.Enabled,
		translationCfg.TranslateTitle,
		translationCfg.TranslateContent,
		maxItems,
		maxTextChars,
	)
	for _, item := range feed.Items {
		if selectedItems >= maxItems {
			ilog.Infof("build email stop by max_items=%d", maxItems)
			break
		}

		originalContent := item.Description
		switch contentField {
		case UseContent:
			originalContent = item.Content
		}
		itemTextChars := len([]rune(strings.TrimSpace(item.Title) + strings.TrimSpace(originalContent)))
		if maxTextChars > 0 && usedTextChars+itemTextChars > maxTextChars {
			ilog.Infof("build email stop by max_text_chars=%d used=%d next_item_chars=%d", maxTextChars, usedTextChars, itemTextChars)
			break
		}
		usedTextChars += itemTextChars
		selectedItems++

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
	ilog.Infof("build email done title=%q body_chars=%d selected_items=%d source_text_chars=%d", feed.Title, len(body), selectedItems, usedTextChars)
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
