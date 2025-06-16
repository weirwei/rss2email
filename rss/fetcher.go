package rss

import (
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/ikit/ilog"
)

func Fetch(url string) (*gofeed.Feed, error) {
	parser := gofeed.NewParser()

	const maxRetries = 3
	var feed *gofeed.Feed
	var err error
	for i := 0; i < maxRetries; i++ {
		feed, err = parser.ParseURL(url)
		if err != nil {
			ilog.Warnf("获取RSS失败，error: %v, retry: %d", err, i+1)
		} else {
			break
		}
		time.Sleep(1 * time.Minute) // 避免过快触发
	}
	if err != nil {
		return nil, err
	}

	// 排序，时间倒序
	// sort.Sort(sort.Reverse(feed))
	return feed, nil
}

type UpdateCheckHandler func(feed *gofeed.Feed) []*gofeed.Item

// 根据guid检查更新
func GUIDUpdateCheckHandler(guid string) UpdateCheckHandler {
	return func(feed *gofeed.Feed) []*gofeed.Item {
		var items []*gofeed.Item
		for _, item := range feed.Items {
			if item.GUID == guid {
				break
			}
			items = append(items, item)
		}
		return items
	}
}

// 根据发布时间检查更新
func PublishedParsedUpdateCheckHandler(publishedParsed time.Time) UpdateCheckHandler {
	return func(feed *gofeed.Feed) []*gofeed.Item {
		var items []*gofeed.Item
		for _, item := range feed.Items {
			if item.PublishedParsed.After(publishedParsed) {
				items = append(items, item)
			}
		}
		return items
	}
}

// 如果已经接受过这个feed，只获取更新部分
func CheckUpdate(feed gofeed.Feed, checkHandler UpdateCheckHandler) gofeed.Feed {
	if checkHandler == nil {
		return feed
	}
	feed.Items = checkHandler(&feed)
	return feed
}
