package service

import (
	"context"
	"errors"

	"github.com/weirwei/rss2email/constants"
)

var ErrDecoHackService = errors.New("decohack 推送失败")

func DecoHackService(ctx context.Context) error {
	if err := RunService(ctx, constants.SubscriptionDecoHack); err != nil {
		return ErrDecoHackService
	}
	return nil
}
