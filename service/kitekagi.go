package service

import (
	"context"
	"errors"

	"github.com/weirwei/rss2email/constants"
)

var ErrKitekagiService = errors.New("kitekagi 推送失败")

func KitekagiService(ctx context.Context) error {
	if err := RunService(ctx, constants.SubscriptionKitekagi); err != nil {
		return ErrKitekagiService
	}
	return nil
}

func KitekagiAIService(ctx context.Context) error {
	if err := RunService(ctx, constants.SubscriptionKitekagiAI); err != nil {
		return ErrKitekagiService
	}
	return nil
}
