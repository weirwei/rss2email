package service

import (
	"context"
	"errors"

	"github.com/weirwei/rss2email/constants"
)

var ErrV2exService = errors.New("v2ex 推送失败")

func V2exService(ctx context.Context) error {
	if err := RunService(ctx, constants.SubscriptionV2ex); err != nil {
		return ErrV2exService
	}
	return nil
}
