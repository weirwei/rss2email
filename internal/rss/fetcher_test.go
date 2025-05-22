package rss

import (
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestFetch(t *testing.T) {
	parser := gofeed.NewParser()
	feed, err := parser.ParseURL("https://decohack.com/feed/")
	if err != nil {
		t.Fatal(err)
	}
	// t.Log(feed)
	t.Log(feed.Items[0].Title)
	t.Log(feed.Items[0].Content)
	t.Log(feed.Items[0].Link)
	t.Log(feed.Items[0].PublishedParsed)
	t.Log(feed.Items[0].Published)
	t.Log(feed.Items[0].PublishedParsed.Format(time.RFC3339))
}
