package service

import (
	"strings"
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/rss2email/constants"
	"github.com/weirwei/rss2email/models"
)

func TestBuildConfigFromFeedSourceNil(t *testing.T) {
	_, err := BuildConfigFromFeedSource(nil)
	if err == nil {
		t.Fatal("expected error for nil feed source")
	}
}

func TestBuildConfigFromFeedSourceBuildFunc(t *testing.T) {
	t.Run("use description", func(t *testing.T) {
		cfg, err := BuildConfigFromFeedSource(&models.FeedSource{
			Subscription: constants.SubscriptionID("demo"),
			FeedURL:      "https://example.com/feed",
			ContentField: constants.FeedContentFieldDescription,
		})
		if err != nil {
			t.Fatal(err)
		}
		_, body := cfg.BuildFunc(&gofeed.Feed{
			Title: "demo",
			Items: []*gofeed.Item{{
				Title:       "item-1",
				Link:        "https://example.com/item-1",
				Description: "desc",
				Content:     "content",
			}},
		})
		if body == "" || !strings.Contains(body, "desc<br>") || strings.Contains(body, "content<br>") {
			t.Fatalf("unexpected body: %s", body)
		}
	})

	t.Run("use content", func(t *testing.T) {
		cfg, err := BuildConfigFromFeedSource(&models.FeedSource{
			Subscription: constants.SubscriptionID("demo"),
			FeedURL:      "https://example.com/feed",
			ContentField: constants.FeedContentFieldContent,
		})
		if err != nil {
			t.Fatal(err)
		}
		_, body := cfg.BuildFunc(&gofeed.Feed{
			Title: "demo",
			Items: []*gofeed.Item{{
				Title:       "item-1",
				Link:        "https://example.com/item-1",
				Description: "desc",
				Content:     "content",
			}},
		})
		if body == "" || !strings.Contains(body, "content<br>") || strings.Contains(body, "desc<br>") {
			t.Fatalf("unexpected body: %s", body)
		}
	})

	t.Run("unknown field fallback to description", func(t *testing.T) {
		cfg, err := BuildConfigFromFeedSource(&models.FeedSource{
			Subscription: constants.SubscriptionID("demo"),
			FeedURL:      "https://example.com/feed",
			ContentField: constants.FeedContentField("unknown"),
		})
		if err != nil {
			t.Fatal(err)
		}
		_, body := cfg.BuildFunc(&gofeed.Feed{
			Title: "demo",
			Items: []*gofeed.Item{{
				Title:       "item-1",
				Link:        "https://example.com/item-1",
				Description: "desc",
				Content:     "content",
			}},
		})
		if body == "" || !strings.Contains(body, "desc<br>") || strings.Contains(body, "content<br>") {
			t.Fatalf("unexpected body: %s", body)
		}
	})
}
