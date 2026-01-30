package service

import (
	"context"
	"errors"

	"github.com/weirwei/rss2email/constants"
)

var ErrAIInsightDailyService = errors.New("ai-insight-daily 推送失败")

func AIInsightDailyService(ctx context.Context) error {
	if err := RunService(ctx, constants.SubscriptionAIInsightDaily); err != nil {
		return ErrAIInsightDailyService
	}
	return nil
}
