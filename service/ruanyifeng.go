package service

import (
	"context"
	"errors"

	"github.com/weirwei/rss2email/constants"
)

var ErrRuanyifengService = errors.New("ruanyifeng 推送失败")

func RuanyifengService(ctx context.Context) error {
	if err := RunService(ctx, constants.SubscriptionRuanyifeng); err != nil {
		return ErrRuanyifengService
	}
	return nil
}
