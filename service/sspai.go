package service

import (
	"context"
	"errors"

	"github.com/weirwei/rss2email/constants"
)

var ErrSspaiService = errors.New("sspai 推送失败")

func SspaiService(ctx context.Context) error {
	if err := RunService(ctx, constants.SubscriptionSspai); err != nil {
		return ErrSspaiService
	}
	return nil
}
