package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/ikit/ilog"
	"github.com/weirwei/rss2email/constants"
	"github.com/weirwei/rss2email/helpers"
	"github.com/weirwei/rss2email/models"
	"github.com/weirwei/rss2email/rss"
)

const (
	ruanyifengFeedURL = "http://feeds.feedburner.com/ruanyifeng"
)

var ErrRuanyifengService = errors.New("ruanyifeng 推送失败")

func RuanyifengService(ctx context.Context) error {
	ilog.Info("RuanyifengService 运行中...")
	defer ilog.Info("RuanyifengService 运行结束...")
	// 从数据库获取订阅信息
	subscriptionList, err := models.NewUserSubscriptionDao().ListBySubscriptionIDAndSubscriptionType(ctx, constants.SubscriptionRuanyifeng, constants.SubscriptionTypeRss)
	if err != nil {
		ilog.Warnf("从数据库获取订阅信息 失败，%v", err)
		return ErrRuanyifengService
	}
	if len(subscriptionList) == 0 {
		return nil
	}
	// 获取feed
	feed, err := rss.Fetch(ruanyifengFeedURL)
	if err != nil {
		ilog.Warnf("获取rss 失败，%v", err)
		return ErrRuanyifengService
	}
	for _, v := range subscriptionList {
		var handler rss.UpdateCheckHandler
		switch v.ProcessType {
		case constants.ProcessTypeGUID:
			handler = rss.GUIDUpdateCheckHandler(v.Process)
		case constants.ProcessTypeTime:
			// 时间戳转时间类型
			ts, err := strconv.ParseInt(v.Process, 10, 64)
			if err != nil {
				ilog.Error(err)
				continue
			}
			t := time.Unix(ts, 0)
			handler = rss.PublishedParsedUpdateCheckHandler(t)
		}
		// 检查需要更新的内容
		f := rss.CheckUpdate(*feed, handler)
		if len(f.Items) > 0 {
			// 邮件样式构建
			subject, body := buildRuanyifeng(&f)
			// 发邮件
			err := helpers.EmailHelper.Send([]string{v.Email}, subject, body)
			if err != nil {
				ilog.Error(err)
			}
			// 更新数据库
			updates := map[string]interface{}{
				"updated_at": time.Now(),
			}
			switch v.ProcessType {
			case constants.ProcessTypeGUID:
				updates["process"] = f.Items[0].GUID
			case constants.ProcessTypeTime:
				updates["process"] = strconv.FormatInt(f.Items[0].PublishedParsed.Unix(), 10)
			}
			err = models.NewUserSubscriptionDao().Update(ctx, v.ID, updates)
			if err != nil {
				ilog.Error(err)
			}
		}
	}

	return nil
}

func buildRuanyifeng(feed *gofeed.Feed) (subject string, body string) {
	subject = feed.Title
	for _, item := range feed.Items {
		body += fmt.Sprintf("<h1><a href=\"%s\">%s</a></h1><br>", item.Link, item.Title)
		body += fmt.Sprintf("%s<br>", item.Description)
	}
	body = fmt.Sprintf(ruanyifengModule, body)
	return
}

var ruanyifengModule = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>阮一峰的网络日志</title>
    <style>
        /* 为页面主体内容创建一个容器 */
        .container {
            max-width: 600px; /* 设置最大宽度为600px */
        }

        div[width="600px"] {
            width: auto !important; /* 覆盖原有的width属性 */
        }

        /* 其他一些基本的样式重置或调整，可选 */
        body {
        }

        img {
            max-width: 100%%; /* 确保图片不会超出容器 */
            height: auto; /* 保持图片比例 */
        }

        a {
            color: #007bff; /* 链接颜色 */
            text-decoration: none; /* 移除下划线 */
        }

        a:hover {
            text-decoration: underline; /* 鼠标悬停时显示下划线 */
        }

        h1, h2 {
            color: #333;
        }

        hr {
            border: 0;
            height: 1px;
            background: #ccc;
            margin: 20px 0;
        }

    </style>
</head>
<body>
%s
</body>
</html>`
