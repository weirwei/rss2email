package service

import (
	"context"
	"errors"

	"github.com/weirwei/rss2email/constants"
)

var ErrZhihuService = errors.New("zhihu 推送失败")

func ZhihuService(ctx context.Context) error {
	if err := RunService(ctx, constants.SubscriptionZhihu); err != nil {
		return ErrZhihuService
	}
	return nil
}
