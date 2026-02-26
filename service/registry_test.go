package service

import (
	"strings"
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/rss2email/conf"
	"github.com/weirwei/rss2email/constants"
	"github.com/weirwei/rss2email/models"
)

func TestBuildConfigFromFeedSourceNil(t *testing.T) {
	_, err := BuildConfigFromFeedSource(nil)
	if err == nil {
		t.Fatal("expected error for nil feed source")
	}
}

type fakeTranslator struct {
	prefix string
}

func (f fakeTranslator) Translate(text string) (string, error) {
	return f.prefix + text, nil
}

func TestBuildEmailWithTranslation(t *testing.T) {
	feed := &gofeed.Feed{
		Title: "demo",
		Items: []*gofeed.Item{{
			Title:       "English Title",
			Link:        "https://example.com/item-1",
			Description: "English description",
		}},
	}

	t.Run("append translation", func(t *testing.T) {
		_, body := buildEmail(feed, UseDescription, fakeTranslator{prefix: "中文:"}, conf.TranslationConfig{
			Enabled:          true,
			TranslateTitle:   true,
			TranslateContent: true,
			AppendTranslated: true,
		})
		if !strings.Contains(body, "English Title（中文:English Title）") {
			t.Fatalf("unexpected title in body: %s", body)
		}
		if !strings.Contains(body, "中文翻译") || !strings.Contains(body, "中文:English description") {
			t.Fatalf("unexpected translated content in body: %s", body)
		}
	})

	t.Run("replace translation", func(t *testing.T) {
		_, body := buildEmail(feed, UseDescription, fakeTranslator{prefix: "中文:"}, conf.TranslationConfig{
			Enabled:          true,
			TranslateTitle:   true,
			TranslateContent: true,
			AppendTranslated: false,
		})
		if strings.Contains(body, "English Title（中文:English Title）") {
			t.Fatalf("unexpected appended title in body: %s", body)
		}
		if !strings.Contains(body, ">中文:English Title</a>") {
			t.Fatalf("unexpected replaced title in body: %s", body)
		}
		if strings.Contains(body, "中文翻译") {
			t.Fatalf("unexpected append marker in replace mode: %s", body)
		}
		if !strings.Contains(body, "中文:English description<br>") {
			t.Fatalf("unexpected replaced content in body: %s", body)
		}
	})
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
